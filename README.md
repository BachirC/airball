# airball

Get CSV-formatted player information about your ESPN fantasy basketball league.

## How to run

- Create `.env` file in the root directory of this project and add from your ESPN fantasy account

  ```shell
  LEAGUE_ID=12345678
  SWID_COOKIE={...}
  ESPNS2_COOKIE=AWRHJwejf...
  ```

- Run the following command

```shell
$ go run main.go <lowerBoundMatchupID> <upperBoundMatchupID>
```

#### Example

```shell
$ go run main.go 1 3
```

Will produce (folders should exist in the current directory)

```shell
 # home_team_id, away_team_id, home_points, away_points, period_id
generated/matchups/matchup-1.csv # number indicates the matchup period
generated/matchups/matchup-2.csv
generated/matchups/matchup-3.csv

# home_team_id, away_team_id, home_points, away_points, period_id, matchup_period_id, period_id
generated/scorings/scorings-1.csv # number indicates the matchup period
generated/scorings/scorings-2.csv
generated/scorings/scorings-3.csv

# pro_player_id,team_id,fantasy_points,lineup_slot_id,minutes_played,scoring_period_id
generated/rostered_players/rostered-players-1.csv # number indicates the matchup period
generated/rostered_players/rostered-players-2.csv
generated/rostered_players/rostered-players-3.csv
```

**Matchup period** : Period during which head-to-heads are happening. There are 16 matchup periods in total.

**Scoring period** : duration of a face-to-face on a given day. A matchup period is composed of several scoring periods.