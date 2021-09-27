package generator

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

const ppScoringsFileName = "pro_players_scorings"

type ProPlayerScoringsGenerator struct {
	ProPlayerStatsFetcher ProPlayerStatsFetcher
}

type ProPlayerStatsFetcher interface {
	FetchPlayerStats() (string, error)
}

type proPlayerScoring struct {
	ProPlayerID     int64 `json:"player_id"`
	TeamID          int64 `json:"team_id"`
	FantasyPoints   int64 `json:"fantasy_points"`
	Minutes         int64 `json:"minutes_played"`
	ScoringPeriodID int64 `json:"scoring_period_id"`
}

func (psg *ProPlayerScoringsGenerator) Generate() error {
	var res []proPlayerScoring

	data, err := psg.ProPlayerStatsFetcher.FetchPlayerStats()
	if err != nil {
		return fmt.Errorf("error generating pro players stats: %w", err)
	}

	gjson.Get(data, "players").ForEach(func(_, rawPlayer gjson.Result) bool {
		gjson.Get(rawPlayer.String(), "player.stats").ForEach(func(_, rawStats gjson.Result) bool {
			p := proPlayerScoring{
				ProPlayerID:     rawPlayer.Get("player.id").Int(),
				TeamID:          rawStats.Get("proTeamId").Int(),
				FantasyPoints:   rawStats.Get("appliedTotal").Int(),
				Minutes:         rawStats.Get("stats.28").Int(),
				ScoringPeriodID: rawStats.Get("scoringPeriodId").Int(),
			}

			res = append(res, p)

			return true
		})

		return true
	})

	err = psg.generateCSV(res)
	if err != nil {
		return fmt.Errorf("could not generate pro players: %w", err)
	}

	return nil
}

func (psg *ProPlayerScoringsGenerator) generateCSV(players []proPlayerScoring) error {
	var rows [][]string
	for _, player := range players {
		row := []string{
			strconv.FormatInt(player.ProPlayerID, 10),
			strconv.FormatInt(player.TeamID, 10),
			strconv.FormatInt(player.FantasyPoints, 10),
			strconv.FormatInt(player.Minutes, 10),
			strconv.FormatInt(player.ScoringPeriodID, 10),
		}

		rows = append(rows, row)
	}

	writer := CSVWriter{
		Path:     csvFilesPathPrefix,
		Filename: ppScoringsFileName,
		Header:   []string{"pro_player_id", "team_id", "fantasy_points", "minutes_played", "scoring_period_id"},
		Rows:     rows,
	}

	err := writer.WriteToCSV()
	if err != nil {
		return fmt.Errorf("could not generate csv: %w", err)
	}

	return nil
}
