package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
)

type playerScoring struct {
	ProPlayerID     int64 `json:"pro_player_id"`
	TeamID          int64 `json:"team_id"`
	FantasyPoints   int64 `json:"fantasy_points"`
	LineupSlotID    int64 `json:"lineup_slot_id"`
	Minutes         int64 `json:"minutes_played"`
	ScoringPeriodID int   `json:"scoring_period_id"`
}

type scoring struct {
	HomeTeamID      int64 `json:"home_team_id"`
	AwayTeamID      int64 `json:"away_team_id"`
	HomePoints      int64 `json:"home_points"`
	AwayPoints      int64 `json:"away_points"`
	MatchupPeriodID int64 `json:"matchup_period_id"`
	PeriodID        int64 `json:"period_id"`
}

type matchup struct {
	HomeTeamID int64 `json:"home_team_id"`
	AwayTeamID int64 `json:"away_team_id"`
	HomePoints int64 `json:"home_points"`
	AwayPoints int64 `json:"away_points"`
	PeriodID   int64 `json:"period_id"`
}

var matchupPeriods = map[int][]string{
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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Unable to load env variables. Make sure you have a .env file in the current directory")
	}
	leagueID := os.Getenv("LEAGUE_ID")
	swidCkie := os.Getenv("SWID_COOKIE")
	espns2Ckie := os.Getenv("ESPNS2_COOKIE")
	if leagueID == "" || swidCkie == "" || espns2Ckie == "" {
		log.Fatalln("LEAGUE_ID, SWID_COOKIE and ESPNS2_COOKIE env vars must be set")
	}

	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalln("Please specify two arguments with a lower and upper bound")
	}

	lowerMatchupID := parseMatchupID(args[0])
	upperMatchupID := parseMatchupID(args[1])
	if lowerMatchupID > upperMatchupID {
		log.Fatalln("Invalid range: first matchupID should be lower or equal to the second one")
	}

	var wg sync.WaitGroup

	for matchupID := lowerMatchupID; matchupID <= upperMatchupID; matchupID++ {
		wg.Add(1)
		go handleMatchup(matchupID, &wg)
	}

	wg.Wait()
}

func leagueID() int {
	l, err := strconv.Atoi(os.Getenv("LEAGUE_ID"))
	if err != nil {
		log.Fatalln("LEAGUE_ID must be a valid integer", err)
	}

	return l
}

func swidCkie() string {
	return os.Getenv("SWID_COOKIE")
}

func espns2Ckie() string {
	return os.Getenv("ESPNS2_COOKIE")
}

func parseMatchupID(arg string) int {
	matchupID, err := strconv.Atoi(arg)
	if err != nil {
		log.Fatalln("invalid integer arg matchupID", err)
	}
	if matchupID < 1 || matchupID > 16 {
		log.Fatalln("matchupID must be between 1 and 16")
	}

	return matchupID
}

func handleMatchup(matchupID int, wg *sync.WaitGroup) {
	defer wg.Done()

	scoringPeriods := matchupPeriods[matchupID]

	// MATCHUP AND DAILY SCORINGS
	scorings, matchups := getScorings(matchupID, scoringPeriods)
	if scorings == nil {
		fmt.Println("Could not get errors")
		return
	}

	f, err := os.Create("dump/matchups/" + strconv.Itoa(matchupID) + "-dump.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"home_team_id", "away_team_id", "home_points", "away_points", "period_id"}); err != nil {
		log.Fatalln("error writing record to file", err)
	}
	for _, r := range matchups {
		if err := w.Write(
			[]string{strconv.FormatInt(r.HomeTeamID, 10), strconv.FormatInt(r.AwayTeamID, 10), strconv.FormatInt(r.HomePoints, 10), strconv.FormatInt(r.AwayPoints, 10), strconv.FormatInt(r.PeriodID, 10)}); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	f2, err := os.Create("dump/scorings/matchup" + strconv.Itoa(matchupID) + "-dump.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer f2.Close()

	w2 := csv.NewWriter(f2)
	defer w2.Flush()
	if err := w2.Write([]string{"home_team_id", "away_team_id", "home_points", "away_points", "matchup_period_id", "period_id"}); err != nil {
		log.Fatalln("error writing record to file", err)
	}
	for _, s := range scorings {
		if err := w2.Write(
			[]string{strconv.FormatInt(s.HomeTeamID, 10), strconv.FormatInt(s.AwayTeamID, 10), strconv.FormatInt(s.HomePoints, 10), strconv.FormatInt(s.AwayPoints, 10), strconv.FormatInt(s.MatchupPeriodID, 10), strconv.FormatInt(s.PeriodID, 10)}); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	// PLAYER SCORING
	pscorings := getPlayersScorings(matchupID, scoringPeriods)

	f3, err := os.Create("dump/player_scorings/matchup" + strconv.Itoa(matchupID) + "-dump.csv")
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer f3.Close()
	w3 := csv.NewWriter(f3)
	defer w3.Flush()
	if err := w3.Write([]string{"pro_player_id", "team_id", "fantasy_points", "lineup_slot_id", "minutes_played", "scoring_period_id"}); err != nil {
		log.Fatalln("error writing record to file", err)
	}
	for _, ps := range pscorings {
		if err := w3.Write(
			[]string{strconv.FormatInt(ps.ProPlayerID, 10), strconv.FormatInt(ps.TeamID, 10), strconv.FormatInt(ps.FantasyPoints, 10), strconv.FormatInt(ps.LineupSlotID, 10), strconv.FormatInt(ps.Minutes, 10), strconv.Itoa(ps.ScoringPeriodID)}); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

  fmt.Println("Matchup " + strconv.Itoa(matchupID) + " done")
}

func getScorings(matchupID int, scoringPeriods []string) ([]scoring, []matchup) {
	swidCookie := &http.Cookie{
		Name:  "swid",
		Value: swidCkie(),
	}
	espns2Cookie := &http.Cookie{
		Name:  "espn_s2",
		Value: espns2Ckie(),
	}

	url := fmt.Sprintf("https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/%d?&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav", leagueID())
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("x-fantasy-filter", fmt.Sprintf("{\"schedule\":{\"filterMatchupPeriodIds\":{\"value\":[%d]}}}", matchupID))
	req.AddCookie(swidCookie)
	req.AddCookie(espns2Cookie)
	if err != nil {
		fmt.Println("Could not create new request %w", err)
		return nil, nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while doing request %w", err)
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("HTTP Status &d: %w", resp.StatusCode, err)
		return nil, nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while parsing body")
		return nil, nil
	}

	schedules := gjson.Get(string(data), "schedule")
	records := []scoring{}
	matchupRecs := []matchup{}
	schedules.ForEach(func(_, schedule gjson.Result) bool {
		aTeam := schedule.Get("away")
		hTeam := schedule.Get("home")
		aScoring := aTeam.Get("pointsByScoringPeriod")
		hScoring := hTeam.Get("pointsByScoringPeriod")
		mr := matchup{
			HomeTeamID: hTeam.Get("teamId").Int(),
			AwayTeamID: aTeam.Get("teamId").Int(),
			HomePoints: hTeam.Get("totalPoints").Int(),
			AwayPoints: aTeam.Get("totalPoints").Int(),
			PeriodID:   int64(matchupID),
		}
		matchupRecs = append(matchupRecs, mr)

		for _, period := range scoringPeriods {
			periodI, _ := strconv.Atoi(period)
			ps := scoring{
				HomeTeamID:      hTeam.Get("teamId").Int(),
				AwayTeamID:      aTeam.Get("teamId").Int(),
				HomePoints:      hScoring.Map()[period].Int(),
				AwayPoints:      aScoring.Map()[period].Int(),
				MatchupPeriodID: int64(matchupID),
				PeriodID:        int64(periodI),
			}
			records = append(records, ps)
		}

		return true
	})

	return records, matchupRecs
}

func getPlayersScorings(matchupID int, scoringPeriods []string) []playerScoring {
	swidCookie := &http.Cookie{
		Name:  "swid",
		Value: swidCkie(),
	}
	espns2Cookie := &http.Cookie{
		Name:  "espn_s2",
		Value: espns2Ckie(),
	}

	records := []playerScoring{}
	for _, scoringPeriod := range scoringPeriods {
		url := fmt.Sprintf("https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/%d?scoringPeriodId=%s&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav", leagueID(), scoringPeriod)
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Set("x-fantasy-filter", fmt.Sprintf("{\"schedule\":{\"filterMatchupPeriodIds\":{\"value\":[%d]}}}", matchupID))
		req.AddCookie(swidCookie)
		req.AddCookie(espns2Cookie)
		if err != nil {
			fmt.Println("Could not create new request %w", err)
			return nil
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error while doing request %w", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("HTTP Status &d: %w", resp.StatusCode, err)
			return nil
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while parsing body")
			return nil
		}

		schedules := gjson.Get(string(data), "schedule")
		schedules.ForEach(func(_, schedule gjson.Result) bool {
			awayTeam := schedule.Get("away")
			homeTeam := schedule.Get("home")

			awayPlayers := awayTeam.Get("rosterForCurrentScoringPeriod.entries")
			homePlayers := homeTeam.Get("rosterForCurrentScoringPeriod.entries")
			scPeriodI, err := strconv.Atoi(scoringPeriod)
			if err != nil {
				fmt.Println("Error while parsing body")
				return false
			}

			homePlayers.ForEach(func(_, player gjson.Result) bool {
				ps := playerScoring{
					ProPlayerID:     player.Get("playerId").Int(),
					TeamID:          homeTeam.Get("teamId").Int(),
					FantasyPoints:   player.Get("playerPoolEntry.appliedStatTotal").Int(),
					LineupSlotID:    player.Get("lineupSlotId").Int(),
					Minutes:         player.Get("playerPoolEntry.player.stats.0.stats.28").Int(),
					ScoringPeriodID: scPeriodI,
				}

				records = append(records, ps)

				return true
			})

			awayPlayers.ForEach(func(_, player gjson.Result) bool {
				ps := playerScoring{
					ProPlayerID:     player.Get("playerId").Int(),
					TeamID:          awayTeam.Get("teamId").Int(),
					FantasyPoints:   player.Get("playerPoolEntry.appliedStatTotal").Int(),
					LineupSlotID:    player.Get("lineupSlotId").Int(),
					Minutes:         player.Get("playerPoolEntry.player.stats.0.stats.28").Int(),
					ScoringPeriodID: scPeriodI,
				}

				records = append(records, ps)

				return true
			})

			return true
		})
	}

	return records
}
