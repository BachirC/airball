# airball

Get CSV-formatted player information about your ESPN fantasy basketball league.

## How to run

- Create `.env` file in the root directory of this project and add from your ESPN fantasy account

  ```shell
  LEAGUE_ID=57860403
  SWID_COOKIE={...}
  ESPNS2_COOKIE=AWRHJwejf...
  ```

- Create `dump/matchups`, `dump/scorings` and `dump/player_scorings` folders.

- Run the following command

```shell
$ go run main.go <lowerBoundMatchupID> <upperBoundMatchupID>
```

#### Example

```shell
$ go run 2 4
```

Will produce (folders should exist in the current directory)

```shell
dump/matchups/1-dump.csv
dump/matchups/2-dump.csv
dump/matchups/3-dump.csv

dump/scorings/matchup1-dump.csv
dump/scorings/matchup2-dump.csv
dump/scorings/matchup3-dump.csv

dump/player_scorings/matchup1-dump.csv
dump/player_scorings/matchup2-dump.csv
dump/player_scorings/matchup3-dump.csv
```