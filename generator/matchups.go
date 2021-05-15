package generator

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

const csvFilesPathPrefix = "generated/csv/"
const matchupsFilePath = csvFilesPathPrefix + "matchups/"
const matchupsFileNamePrefix = "matchup-"

var ScoringPeriods = map[int][]string{
	1:  {"1", "2", "4", "5", "6"},
	2:  {"7", "8", "9", "10", "11", "12", "13"},
	3:  {"14", "15", "16", "17", "18", "19", "20"},
	4:  {"21", "22", "23", "24", "25", "26", "27"},
	5:  {"28", "29", "30", "31", "32", "33", "34"},
	6:  {"35", "36", "37", "38", "39", "40", "41"},
	7:  {"42", "43", "44", "45", "46", "47", "48"},
	8:  {"49", "50", "51", "52", "53", "54", "55"},
	9:  {"56", "57", "58", "59", "60", "61", "62"},
	10: {"63", "64", "65", "66", "67", "68", "69"},
	11: {"70", "71", "72", "73", "79", "80", "81", "82", "83"},
	12: {"84", "85", "86", "87", "88", "89", "90"},
	13: {"91", "92", "93", "94", "95", "96", "97"},
	14: {"98", "99", "100", "101", "102", "103", "104"},
	15: {"105", "106", "107", "108", "109", "110", "111"},
	16: {"112", "113", "114", "115", "116", "117", "118"},
}

type MatchupsGenerator struct {
	MatchupPeriod       int
	MatchupStatsFetcher MatchupStatsFetcher
}

type MatchupStatsFetcher interface {
	FetchStatsInMatchupPeriod(int) (string, error)
}

type matchup struct {
	HomeTeamID int64 `json:"home_team_id"`
	AwayTeamID int64 `json:"away_team_id"`
	HomePoints int64 `json:"home_points"`
	AwayPoints int64 `json:"away_points"`
	PeriodID   int64 `json:"period_id"`
}

func (mg *MatchupsGenerator) Generate() error {
	stats, err := mg.MatchupStatsFetcher.FetchStatsInMatchupPeriod(mg.MatchupPeriod)
	if err != nil {
		return fmt.Errorf("could not generate matchups: %w", err)
	}

	schedules := gjson.Get(stats, "schedule")
	matchups := []matchup{}
	schedules.ForEach(func(_, schedule gjson.Result) bool {
		aTeam := schedule.Get("away")
		hTeam := schedule.Get("home")
		matchup := matchup{
			HomeTeamID: hTeam.Get("teamId").Int(),
			AwayTeamID: aTeam.Get("teamId").Int(),
			HomePoints: hTeam.Get("totalPoints").Int(),
			AwayPoints: aTeam.Get("totalPoints").Int(),
			PeriodID:   int64(mg.MatchupPeriod),
		}
		matchups = append(matchups, matchup)

		return true
	})

	err = mg.generateCSV(matchups)
	if err != nil {
		return fmt.Errorf("could not generate matchups: %w", err)
	}

	return nil
}

func (mg *MatchupsGenerator) generateCSV(matchups []matchup) error {
	var rows [][]string
	for _, matchup := range matchups {
		row := []string{
			strconv.FormatInt(matchup.HomeTeamID, 10),
			strconv.FormatInt(matchup.AwayTeamID, 10),
			strconv.FormatInt(matchup.HomePoints, 10),
			strconv.FormatInt(matchup.AwayPoints, 10),
			strconv.FormatInt(matchup.PeriodID, 10),
		}

		rows = append(rows, row)
	}

	writer := CSVWriter{
		Path:     matchupsFilePath,
		Filename: matchupsFileNamePrefix + strconv.Itoa(mg.MatchupPeriod),
		Header:   []string{"home_team_id", "away_team_id", "home_points", "away_points", "period_id"},
		Rows:     rows,
	}

	err := writer.WriteToCSV()
	if err != nil {
		return fmt.Errorf("could not generate csv: %w", err)
	}

	return nil
}
