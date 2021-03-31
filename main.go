package main

import "fmt"

func main() {
	fmt.Println("OK")

  url := "https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/57860403?scoringPeriodId=71&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav"
  resp, err := httpGet(
}
