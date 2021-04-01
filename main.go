package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

func main() {
	leagueID := 57860403
	matchupID := 13
	scoringPeriods := []int{91, 92, 93, 94, 95, 96, 97}
	swidCookie := &http.Cookie{
		Name:  "swid",
		Value: "{C36EA600-DE29-4AC3-B0E2-94069370EC1D}",
	}
	espns2Cookie := &http.Cookie{
		Name:  "espn_s2",
		Value: "AECninTC3IGAkv47ONJa%2B8bLbqom2yMJq2LwmSL9fRYCqiSwU8GUKeYV76wL4HR%2B2utBq9YlOJd3kl8FV3tv2FmaKMhdgkHm4pmAbuWLQrCpPW0tvqZYvGb8oH3ju1H3vgZ5vcmJic1Y4AURBGS1PSf7Fw5ACEk6itkWM4Qx66dPrBaRu8VVVsWQabh%2FGTLx8z0a8L%2B5c6N4M1IHGJT9RhKNq3a%2FnAjY6Q5jh75eZQvh393994x0N9OJElk2qYtziEefAnH1bD1KarSbz%2FQ92fIJ",
	}

	for _, scoringPeriod := range scoringPeriods {
		url := fmt.Sprintf("https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/%d?scoringPeriodId=%d&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav", leagueID, scoringPeriod)
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Set("x-fantasy-filter", fmt.Sprintf("{\"schedule\":{\"filterMatchupPeriodIds\":{\"value\":[%d]}}}", matchupID))
		req.AddCookie(swidCookie)
		req.AddCookie(espns2Cookie)
		if err != nil {
			fmt.Println("Could not create new request %w", err)
			return
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error while doing request %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("HTTP Status &d: %w", resp.StatusCode, err)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while parsing body")
			return
		}

		schedules := gjson.Get(string(data), "schedule")
		records := []playerScoring{}
		schedules.ForEach(func(_, schedule gjson.Result) bool {
			awayTeam := schedule.Get("away")
			homeTeam := schedule.Get("home")

			awayPlayers := awayTeam.Get("rosterForCurrentScoringPeriod.entries")
			homePlayers := homeTeam.Get("rosterForCurrentScoringPeriod.entries")
			homePlayers.ForEach(func(_, player gjson.Result) bool {
				ps := playerScoring{
					ProPlayerID:     player.Get("playerId").Int(),
					TeamID:          homeTeam.Get("teamId").Int(),
					FantasyPoints:   player.Get("playerPoolEntry.appliedStatTotal").Int(),
					LineupSlotID:    player.Get("lineupSlotId").Int(),
					Minutes:         player.Get("playerPoolEntry.player.stats.0.stats.28").Int(),
					ScoringPeriodID: scoringPeriod,
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
					ScoringPeriodID: scoringPeriod,
				}

				records = append(records, ps)

				return true
			})

			return true
		})

		jsoned, _ := json.Marshal(records)
		fmt.Println(string(jsoned))
	}
}
