package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type ESPNAPIClient struct {
	LeagueID     int
	SWIDCookie   string
	ESPNS2Cookie string
}

func (client *ESPNAPIClient) FetchStatsInScoringPeriod(matchupPeriod, scoringPeriod int) (string, error) {
	url := fmt.Sprintf("https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/%d?scoringPeriodId=%d&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav", client.LeagueID, scoringPeriod)
	headers := map[string]string{
		"x-fantasy-filter": fmt.Sprintf("{\"schedule\":{\"filterMatchupPeriodIds\":{\"value\":[%d]}}}", matchupPeriod),
	}

	stats, err := client.doRequest("GET", url, nil, headers)
	if err != nil {
		return "", fmt.Errorf("could not fetch players stats for matchup %d and period %d: %w", matchupPeriod, scoringPeriod, err)
	}

	return string(stats), nil
}

func (client *ESPNAPIClient) FetchStatsInMatchupPeriod(matchupPeriod int) (string, error) {
	url := fmt.Sprintf("https://fantasy.espn.com/apis/v3/games/fba/seasons/2021/segments/0/leagues/%d?&view=mBoxscore&view=mMatchupScore&view=mRoster&view=mSettings&view=mStatus&view=mTeam&view=modular&view=mNav", client.LeagueID)
	headers := map[string]string{
		"x-fantasy-filter": fmt.Sprintf("{\"schedule\":{\"filterMatchupPeriodIds\":{\"value\":[%d]}}}", matchupPeriod),
	}

	stats, err := client.doRequest("GET", url, nil, headers)
	if err != nil {
		return "", fmt.Errorf("could not fetch team stats for matchup %d: %w", matchupPeriod, err)
	}

	return string(stats), nil
}

func (client *ESPNAPIClient) doRequest(method string, path string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %w", err)
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	// Mandatory for authenticating the requests to the ESPN API.
	cookies := map[string]string{
		"swid":    client.SWIDCookie,
		"espn_s2": client.ESPNS2Cookie,
	}
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not do http request: %w", err)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse http response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status=%d, body=%v", resp.StatusCode, respBody)
	}

	return respBody, nil
}
