package generator

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

const scoringsFilePath = csvFilesPathPrefix + "scorings/"
const scoringsFileNamePrefix = "scorings-"

type ScoringsGenerator struct {
	MatchupPeriod       int
	MatchupStatsFetcher MatchupStatsFetcher
}

type scoring struct {
	HomeTeamID      int64 `json:"home_team_id"`
	AwayTeamID      int64 `json:"away_team_id"`
	HomePoints      int64 `json:"home_points"`
	AwayPoints      int64 `json:"away_points"`
	MatchupPeriodID int64 `json:"matchup_period_id"`
	PeriodID        int64 `json:"period_id"`
}

func (sg *ScoringsGenerator) Generate() error {
	stats, err := sg.MatchupStatsFetcher.FetchStatsInMatchupPeriod(sg.MatchupPeriod)
	if err != nil {
		return fmt.Errorf("could not generate scorings: %w", err)
	}

	var scorings []scoring
	schedules := gjson.Get(stats, "schedule")
	schedules.ForEach(func(_, schedule gjson.Result) bool {
		for _, scoringPeriod := range ScoringPeriods[sg.MatchupPeriod] {
			aTeam := schedule.Get("away")
			hTeam := schedule.Get("home")
			aScoring := aTeam.Get("pointsByScoringPeriod")
			hScoring := hTeam.Get("pointsByScoringPeriod")
			sPeriodInt, err := strconv.Atoi(scoringPeriod)
			if err != nil {
				fmt.Println("could not parse period while generating scorings: %w", err)
				return false
			}

			ps := scoring{
				HomeTeamID:      hTeam.Get("teamId").Int(),
				AwayTeamID:      aTeam.Get("teamId").Int(),
				HomePoints:      hScoring.Map()[scoringPeriod].Int(),
				AwayPoints:      aScoring.Map()[scoringPeriod].Int(),
				MatchupPeriodID: int64(sg.MatchupPeriod),
				PeriodID:        int64(sPeriodInt),
			}
			scorings = append(scorings, ps)
		}

		return true
	})

	err = sg.generateCSV(scorings)
	if err != nil {
		return fmt.Errorf("could not generate scorings: %w", err)
	}

	return nil
}

func (sg *ScoringsGenerator) generateCSV(scorings []scoring) error {
	var rows [][]string
	for _, scoring := range scorings {
		row := []string{
			strconv.FormatInt(scoring.HomeTeamID, 10),
			strconv.FormatInt(scoring.AwayTeamID, 10),
			strconv.FormatInt(scoring.HomePoints, 10),
			strconv.FormatInt(scoring.AwayPoints, 10),
			strconv.FormatInt(scoring.MatchupPeriodID, 10),
			strconv.FormatInt(scoring.PeriodID, 10),
		}

		rows = append(rows, row)
	}

	writer := CSVWriter{
		Path:     scoringsFilePath,
		Filename: scoringsFileNamePrefix + strconv.Itoa(sg.MatchupPeriod),
		Header:   []string{"home_team_id", "away_team_id", "home_points", "away_points", "matchup_period_id", "period_id"},
		Rows:     rows,
	}

	err := writer.WriteToCSV()
	if err != nil {
		return fmt.Errorf("could not generate csv: %w", err)
	}

	return nil
}
