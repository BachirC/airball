package generator

import (
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

const rosteredPlayersFilePath = csvFilesPathPrefix + "rostered_players/"
const rosteredPlayersFileNamePrefix = "rostered-players-"

type RosteredPlayersGenerator struct {
	MatchupPeriod               int
	RosteredPlayersStatsFetcher RosteredPlayersStatsFetcher
}

type RosteredPlayersStatsFetcher interface {
	FetchStatsInScoringPeriod(int, int) (string, error)
}

type rosteredPlayer struct {
	ProPlayerID     int64 `json:"pro_player_id"`
	TeamID          int64 `json:"team_id"`
	FantasyPoints   int64 `json:"fantasy_points"`
	LineupSlotID    int64 `json:"lineup_slot_id"`
	Minutes         int64 `json:"minutes_played"`
	ScoringPeriodID int   `json:"scoring_period_id"`
}

func (rpg *RosteredPlayersGenerator) Generate() error {
	var players []rosteredPlayer

	for _, scoringPeriod := range ScoringPeriods[rpg.MatchupPeriod] {
		scoringPeriodInt, err := strconv.Atoi(scoringPeriod)
		if err != nil {
			return fmt.Errorf("error parsing scoring period while generating rostered players stats: %w", err)
		}

		data, err := rpg.RosteredPlayersStatsFetcher.FetchStatsInScoringPeriod(rpg.MatchupPeriod, scoringPeriodInt)
		if err != nil {
			return fmt.Errorf("error generating rostered players stats: %w", err)
		}

		schedules := gjson.Get(data, "schedule")
		schedules.ForEach(func(_, schedule gjson.Result) bool {
			awayTeam := schedule.Get("away")
			homeTeam := schedule.Get("home")

			awayPlayers := parseTeamPlayers(awayTeam, scoringPeriodInt)
			homePlayers := parseTeamPlayers(homeTeam, scoringPeriodInt)

			players = append(players, awayPlayers...)
			players = append(players, homePlayers...)

			return true
		})
	}

	err := rpg.generateCSV(players)
	if err != nil {
		return fmt.Errorf("could not generate matchups: %w", err)
	}

	return nil
}

func (rpg *RosteredPlayersGenerator) generateCSV(players []rosteredPlayer) error {
	var rows [][]string
	for _, player := range players {
		row := []string{
			strconv.FormatInt(player.ProPlayerID, 10),
			strconv.FormatInt(player.TeamID, 10),
			strconv.FormatInt(player.FantasyPoints, 10),
			strconv.FormatInt(player.LineupSlotID, 10),
			strconv.FormatInt(player.Minutes, 10),
			strconv.Itoa(player.ScoringPeriodID),
		}

		rows = append(rows, row)
	}

	writer := CSVWriter{
		Path:     rosteredPlayersFilePath,
		Filename: rosteredPlayersFileNamePrefix + strconv.Itoa(rpg.MatchupPeriod),
		Header:   []string{"pro_player_id", "team_id", "fantasy_points", "lineup_slot_id", "minutes_played", "scoring_period_id"},
		Rows:     rows,
	}

	err := writer.WriteToCSV()
	if err != nil {
		return fmt.Errorf("could not generate csv: %w", err)
	}

	return nil
}

func parseTeamPlayers(team gjson.Result, scoringPeriod int) []rosteredPlayer {
	var res []rosteredPlayer
	players := team.Get("rosterForCurrentScoringPeriod.entries")
	players.ForEach(func(_, player gjson.Result) bool {
		rp := rosteredPlayer{
			ProPlayerID:     player.Get("playerId").Int(),
			TeamID:          team.Get("teamId").Int(),
			FantasyPoints:   player.Get("playerPoolEntry.appliedStatTotal").Int(),
			LineupSlotID:    player.Get("lineupSlotId").Int(),
			Minutes:         player.Get("playerPoolEntry.player.stats.0.stats.28").Int(),
			ScoringPeriodID: scoringPeriod,
		}

		res = append(res, rp)

		return true
	})

	return res
}
