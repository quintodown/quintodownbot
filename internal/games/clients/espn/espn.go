package espn

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/quintodown/quintodownbot/internal/clock"

	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/games"
)

const (
	endpointForCalendar       = "https://site.api.espn.com/apis/site/v2/sports/football/%s/scoreboard"
	endpointForGames          = "https://site.api.espn.com/apis/site/v2/sports/football/%s/summary"
	timeLayout                = "2006-01-02T15:04Z"
	statusFinal               = "STATUS_FINAL"
	statusInProgress          = "STATUS_IN_PROGRESS"
	farenheitConversionFactor = 32
	farenheitDivider          = 9
	userAgent                 = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like " +
		"Gecko) Chrome/92.0.4515.159 Safari/537.36"
	referer = "https://espndeportes.espn.com"
)

type urlParameters map[string]string

type Client struct {
	client *http.Client
	clk    clock.Clock
}

func NewESPNClient(c *http.Client, clk clock.Clock) games.GameInfoClient {
	return &Client{client: c, clk: clk}
}

func (ec *Client) GetGames(competition games.Competition) ([]games.Game, error) {
	const datesLayout = "20060102"

	var scb scoreboard

	if err := ec.executeCall(endpointForCalendar, competition, urlParameters{}, &scb); err != nil {
		return nil, err
	}

	calendar := scb.toCalendar()
	params := urlParameters{}
	now := ec.clk.Now()

	for _, v := range calendar {
		if v.Start.UTC().Before(now) && v.End.UTC().After(now) {
			params["dates"] = fmt.Sprintf("%s-%s", v.Start.Format(datesLayout), v.End.Format(datesLayout))

			break
		}
	}

	if err := ec.executeCall(endpointForCalendar, competition, params, &scb); err != nil {
		return nil, err
	}

	return scb.toGames(competition, calendar), nil
}

func (ec *Client) GetGameInformation(competition games.Competition, id string) (games.Game, error) {
	var gsc gameScore
	if err := ec.executeCall(endpointForGames, competition, map[string]string{"event": id}, &gsc); err != nil {
		return games.Game{}, err
	}

	return gsc.toGame(competition)
}

func (ec Client) executeCall(
	endpoint string,
	c games.Competition,
	parameters urlParameters,
	response easyjson.Unmarshaler,
) error {
	req, err := ec.getRequest(endpoint, c, parameters)
	if err != nil {
		return err
	}

	resp, err := ec.client.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if err := easyjson.UnmarshalFromReader(resp.Body, response); err != nil {
		return err
	}

	return nil
}

func (ec *Client) getRequest(
	endpoint string,
	c games.Competition,
	parameters urlParameters,
) (*http.Request, error) {
	ep, err := url.Parse(fmt.Sprintf(endpoint, ec.getCompetitionSlug(c)))
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	for k, v := range parameters {
		params.Add(k, v)
	}

	params.Add("lang", "es")
	params.Add("region", "us")

	ep.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", ep.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authority", "site.api.espn.com")
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"92\", \" Not A;Brand\";v=\"99\", \"Google Chrome\";v=\"92\"")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", referer)
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept-Language", "en,es-ES;q=0.9,es;q=0.8,ca;q=0.7,ja;q=0.6")

	return req, nil
}

func (ec *Client) getCompetitionSlug(c games.Competition) string {
	equivalents := map[games.Competition]string{
		games.NFL:  "nfl",
		games.NCAA: "college-football",
		games.CFL:  "cfl",
	}

	return equivalents[c]
}

func getGameStatus(status string) games.GameState {
	switch status {
	case statusInProgress:
		return games.InProgressState
	case statusFinal:
		return games.FinishedState
	default:
		return games.ScheduledState
	}
}
