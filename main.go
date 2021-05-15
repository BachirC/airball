package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"bachirc/airball/generator"
	"bachirc/airball/http"

	"github.com/joho/godotenv"
)

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

	// MATCHUPS
	err := getMatchup(matchupID)
	if err != nil {
		log.Fatalln("could not get matchup for matchup period"+strconv.Itoa(matchupID), err)
	}

	// SCORINGS
	err = getScorings(matchupID)
	if err != nil {
		log.Fatalln("could not get scorings for matchup "+strconv.Itoa(matchupID), err)
	}

	// ROSTERTED PLAYERS SCORING
	err = getRosteredPlayers(matchupID)
	if err != nil {
		log.Fatalln("could not get rosterted players for matchup "+strconv.Itoa(matchupID), err)
	}

	fmt.Println("Matchup " + strconv.Itoa(matchupID) + " done")
}

func getMatchup(matchupID int) error {
	client := &http.ESPNAPIClient{LeagueID: leagueID(), SWIDCookie: swidCkie(), ESPNS2Cookie: espns2Ckie()}
	matchupsGen := &generator.MatchupsGenerator{
		MatchupPeriod:       matchupID,
		MatchupStatsFetcher: client,
	}
	err := matchupsGen.Generate()
	if err != nil {
		return fmt.Errorf("could not get matchups: %w", err)
	}

	return nil
}

func getScorings(matchupID int) error {
	client := &http.ESPNAPIClient{LeagueID: leagueID(), SWIDCookie: swidCkie(), ESPNS2Cookie: espns2Ckie()}
	scoringsGen := &generator.ScoringsGenerator{
		MatchupPeriod:       matchupID,
		MatchupStatsFetcher: client,
	}
	err := scoringsGen.Generate()
	if err != nil {
		return fmt.Errorf("could not get scorings: %w", err)
	}

	return nil
}

func getRosteredPlayers(matchupPeriod int) error {
	client := &http.ESPNAPIClient{LeagueID: leagueID(), SWIDCookie: swidCkie(), ESPNS2Cookie: espns2Ckie()}
	rosterPlayersGen := &generator.RosteredPlayersGenerator{
		MatchupPeriod:               matchupPeriod,
		RosteredPlayersStatsFetcher: client,
	}

	err := rosterPlayersGen.Generate()
	if err != nil {
		fmt.Println("could not get rostered players scorings: %w", err)
		return nil
	}

	return nil
}
