# airball

Run with 

```shell
$ go run main.go <lowerBoundMatchupID> <upperBoundMatchupID>
```



### Example

```shell
$ go run 2 4
```

Will produce (folders should exist in the current directory) :

```
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