package espn

import (
	"fmt"
	"github.com/quintodown/quintodownbot/internal/clock"
	"net/http"
	"net/url"
	"strconv"
	"time"

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
)

type urlParameters map[string]string

//easyjson:json
type gameScore struct {
	Boxscore struct {
		Teams []struct {
			Team struct {
				ID               string `json:"id"`
				UID              string `json:"uid"`
				Slug             string `json:"slug"`
				Location         string `json:"location"`
				Name             string `json:"name"`
				Abbreviation     string `json:"abbreviation"`
				DisplayName      string `json:"displayName"`
				ShortDisplayName string `json:"shortDisplayName"`
				Color            string `json:"color"`
				AlternateColor   string `json:"alternateColor"`
				Logo             string `json:"logo"`
			} `json:"team"`
		} `json:"teams"`
	} `json:"boxscore"`
	GameInfo struct {
		Venue struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
			Address  struct {
				City    string `json:"city"`
				State   string `json:"state"`
				ZipCode string `json:"zipCode"`
			} `json:"address"`
			Capacity int  `json:"capacity"`
			Grass    bool `json:"grass"`
		} `json:"venue"`
		Attendance int `json:"attendance"`
		Weather    struct {
			Temperature   int    `json:"temperature"`
			ConditionID   string `json:"conditionId"`
			Gust          int    `json:"gust"`
			Precipitation int    `json:"precipitation"`
		} `json:"weather"`
	} `json:"gameInfo"`
	Header header `json:"header"`
}

//easyjson:json
type scoreboard struct {
	Leagues []struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		Season       struct {
			Year      int    `json:"year"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Type      struct {
				ID           string `json:"id"`
				Type         int    `json:"type"`
				Name         string `json:"name"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
		} `json:"season"`
		CalendarType        string `json:"calendarType"`
		CalendarIsWhitelist bool   `json:"calendarIsWhitelist"`
		CalendarStartDate   string `json:"calendarStartDate"`
		CalendarEndDate     string `json:"calendarEndDate"`
		Calendar            []struct {
			Label     string `json:"label"`
			Value     string `json:"value"`
			StartDate string `json:"startDate"`
			EndDate   string `json:"endDate"`
			Entries   []struct {
				Label          string `json:"label"`
				AlternateLabel string `json:"alternateLabel"`
				Detail         string `json:"detail"`
				Value          string `json:"value"`
				StartDate      string `json:"startDate"`
				EndDate        string `json:"endDate"`
			} `json:"entries"`
		} `json:"calendar"`
	} `json:"leagues"`
	Season struct {
		Type int `json:"type"`
		Year int `json:"year"`
	} `json:"season"`
	Week struct {
		Number int `json:"number"`
	} `json:"week"`
	Events []event `json:"events"`
}

//easyjson:json
type competition struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	Date       string `json:"date"`
	Attendance int    `json:"attendance"`
	Type       struct {
		ID           string `json:"id"`
		Abbreviation string `json:"abbreviation"`
	} `json:"type"`
	TimeValid             bool  `json:"timeValid"`
	NeutralSite           bool  `json:"neutralSite"`
	ConferenceCompetition bool  `json:"conferenceCompetition"`
	Recent                bool  `json:"recent"`
	Venue                 venue `json:"venue"`
	Competitors           []struct {
		ID       string `json:"id"`
		UID      string `json:"uid"`
		Type     string `json:"type"`
		Order    int    `json:"order"`
		HomeAway string `json:"homeAway"`
		Team     struct {
			ID               string `json:"id"`
			UID              string `json:"uid"`
			Location         string `json:"location"`
			Name             string `json:"name"`
			Abbreviation     string `json:"abbreviation"`
			DisplayName      string `json:"displayName"`
			ShortDisplayName string `json:"shortDisplayName"`
			Color            string `json:"color"`
			AlternateColor   string `json:"alternateColor"`
			IsActive         bool   `json:"isActive"`
			Venue            struct {
				ID string `json:"id"`
			} `json:"venue"`
			Logo string `json:"logo"`
		} `json:"team"`
		Score      string        `json:"score"`
		Statistics []interface{} `json:"statistics"`
		Records    []struct {
			Name         string `json:"name"`
			Abbreviation string `json:"abbreviation,omitempty"`
			Type         string `json:"type"`
			Summary      string `json:"summary"`
		} `json:"records"`
	} `json:"competitors"`
	Status struct {
		Clock        float64 `json:"clock"`
		DisplayClock string  `json:"displayClock"`
		Period       int     `json:"period"`
		Type         struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			State       string `json:"state"`
			Completed   bool   `json:"completed"`
			Description string `json:"description"`
			Detail      string `json:"detail"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	} `json:"status"`
	StartDate string `json:"startDate"`
}

//easyjson:json
type venue struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Address  struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"address"`
	Capacity int  `json:"capacity"`
	Indoor   bool `json:"indoor"`
}

//easyjson:json
type event struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Date      string `json:"date"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Season    struct {
		Year int    `json:"year"`
		Type int    `json:"type"`
		Slug string `json:"slug"`
	} `json:"season"`
	Competitions []competition `json:"competitions"`
	Weather      struct {
		DisplayValue string `json:"displayValue"`
		Temperature  int    `json:"temperature"`
		ConditionID  string `json:"conditionId"`
	} `json:"weather"`
	Status struct {
		Clock        float64 `json:"clock"`
		DisplayClock string  `json:"displayClock"`
		Period       int     `json:"period"`
		Type         struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			State       string `json:"state"`
			Completed   bool   `json:"completed"`
			Description string `json:"description"`
			Detail      string `json:"detail"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
	} `json:"status"`
}

//easyjson:json
type header struct {
	ID     string `json:"id"`
	UID    string `json:"uid"`
	Season struct {
		Year int `json:"year"`
		Type int `json:"type"`
	} `json:"season"`
	TimeValid    bool                `json:"timeValid"`
	Competitions []headerCompetition `json:"competitions"`
	Week         int                 `json:"week"`
	League       struct {
		ID           string `json:"id"`
		UID          string `json:"uid"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Slug         string `json:"slug"`
		IsTournament bool   `json:"isTournament"`
		Links        []struct {
			Rel  []string `json:"rel"`
			Href string   `json:"href"`
			Text string   `json:"text"`
		} `json:"links"`
	} `json:"league"`
}

//easyjson:json
//nolint:tagliatelle
type headerCompetition struct {
	ID                    string `json:"id"`
	UID                   string `json:"uid"`
	Date                  string `json:"date"`
	NeutralSite           bool   `json:"neutralSite"`
	ConferenceCompetition bool   `json:"conferenceCompetition"`
	BoxscoreAvailable     bool   `json:"boxscoreAvailable"`
	CommentaryAvailable   bool   `json:"commentaryAvailable"`
	LiveAvailable         bool   `json:"liveAvailable"`
	OnWatchESPN           bool   `json:"onWatchESPN"`
	Recent                bool   `json:"recent"`
	BoxscoreSource        string `json:"boxscoreSource"`
	PlayByPlaySource      string `json:"playByPlaySource"`
	Competitors           []struct {
		ID       string `json:"id"`
		UID      string `json:"uid"`
		Order    int    `json:"order"`
		HomeAway string `json:"homeAway"`
		Winner   bool   `json:"winner"`
		Team     struct {
			ID             string `json:"id"`
			UID            string `json:"uid"`
			Location       string `json:"location"`
			Name           string `json:"name"`
			Nickname       string `json:"nickname"`
			Abbreviation   string `json:"abbreviation"`
			DisplayName    string `json:"displayName"`
			Color          string `json:"color"`
			AlternateColor string `json:"alternateColor"`
			Logos          []struct {
				Href   string   `json:"href"`
				Width  int      `json:"width"`
				Height int      `json:"height"`
				Alt    string   `json:"alt"`
				Rel    []string `json:"rel"`
			} `json:"logos"`
		} `json:"team"`
		Score      string `json:"score"`
		Linescores []struct {
			DisplayValue string `json:"displayValue"`
		} `json:"linescores"`
		Record []struct {
			Type         string `json:"type"`
			Summary      string `json:"summary"`
			DisplayValue string `json:"displayValue"`
		} `json:"record"`
		Possession bool `json:"possession"`
	} `json:"competitors"`
	Status struct {
		Type struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			State       string `json:"state"`
			Completed   bool   `json:"completed"`
			Description string `json:"description"`
			Detail      string `json:"detail"`
			ShortDetail string `json:"shortDetail"`
		} `json:"type"`
		DisplayClock string `json:"displayClock"`
		Period       int    `json:"period"`
	} `json:"status"`
}

func (v gameScore) toGame(c games.Competition) (games.Game, error) {
	t, err := time.Parse(timeLayout, v.Header.Competitions[0].Date)
	if err != nil {
		return games.Game{}, err
	}

	g := games.Game{
		Id: v.Header.ID,
		Venue: games.Venue{
			FullName: v.GameInfo.Venue.FullName,
			Address: games.VenueAddress{
				City:  v.GameInfo.Venue.Address.City,
				State: v.GameInfo.Venue.Address.State,
			},
			Capacity: v.GameInfo.Venue.Capacity,
		},
		Start: t.UTC(),
		Status: games.GameStatus{
			State:        getGameStatus(v.Header.Competitions[0].Status.Type.Name),
			DisplayClock: v.Header.Competitions[0].Status.DisplayClock,
			Period:       v.Header.Competitions[0].Status.Period,
		},
		Weather: games.GameWeather{
			Temperature: (v.GameInfo.Weather.Temperature - farenheitConversionFactor) * 5 / farenheitDivider,
		},
		Competition: c,
	}

	for _, v := range v.Header.Competitions[0].Competitors {
		logo := ""
		if len(v.Team.Logos) > 0 {
			logo = v.Team.Logos[0].Href
		}

		record := ""
		if len(v.Record) > 0 {
			record = v.Record[0].Summary
		}

		score, _ := strconv.Atoi(v.Score)
		team := games.TeamScore{
			Name:             v.Team.DisplayName,
			ShortDisplayName: v.Team.Name,
			Score:            score,
			Logo:             logo,
			Record:           record,
		}

		if v.HomeAway == "home" {
			g.HomeTeam = team
		} else {
			g.AwayTeam = team
		}
	}

	g.Name = g.AwayTeam.Name + " @ " + g.HomeTeam.Name

	return g, nil
}

func (v scoreboard) toCalendar() []games.Week {
	var weeks []games.Week

	for _, c := range v.Leagues[0].Calendar {
		for _, v := range c.Entries {
			start, err := time.Parse(timeLayout, v.StartDate)
			if err != nil {
				continue
			}

			end, err := time.Parse(timeLayout, v.EndDate)
			if err != nil {
				continue
			}

			weeks = append(weeks, games.Week{
				Name:  v.Label,
				Start: start,
				End:   end,
			})
		}
	}

	return weeks
}

func (v scoreboard) toGames(c games.Competition, calendar []games.Week) []games.Game {
	found := make([]games.Game, 0, len(v.Events))

	for _, event := range v.Events {
		t, err := time.Parse(timeLayout, event.Date)
		if err != nil {
			continue
		}

		g := games.Game{
			Id:    event.ID,
			Name:  event.Name,
			Start: t.UTC(),
			Venue: v.getVenue(event),
			Status: games.GameStatus{
				Clock:        event.Status.Clock,
				DisplayClock: event.Status.DisplayClock,
				Period:       event.Status.Period,
				State:        getGameStatus(event.Competitions[0].Status.Type.Name),
			},
			Weather: games.GameWeather{
				DisplayValue: event.Weather.DisplayValue,
				Temperature:  event.Weather.Temperature,
			},
			Competition: c,
		}

		g.WeekName = v.getWeekName(calendar, t.UTC())

		for _, competitor := range event.Competitions[0].Competitors {
			record := ""
			if len(competitor.Records) > 0 {
				record = competitor.Records[0].Summary
			}

			score, _ := strconv.Atoi(competitor.Score)
			teamScore := games.TeamScore{
				Score:            score,
				Name:             competitor.Team.DisplayName,
				ShortDisplayName: competitor.Team.ShortDisplayName,
				Logo:             competitor.Team.Logo,
				Record:           record,
			}

			if competitor.HomeAway == "home" {
				g.HomeTeam = teamScore
			} else {
				g.AwayTeam = teamScore
			}
		}

		found = append(found, g)
	}

	return found
}

func (v scoreboard) getVenue(event event) games.Venue {
	return games.Venue{
		FullName: event.Competitions[0].Venue.FullName,
		Address: games.VenueAddress{
			City:  event.Competitions[0].Venue.Address.City,
			State: event.Competitions[0].Venue.Address.State,
		},
		Capacity: event.Competitions[0].Venue.Capacity,
		Indoor:   event.Competitions[0].Venue.Indoor,
	}
}

func (v scoreboard) getWeekName(calendar []games.Week, t time.Time) string {
	for _, w := range calendar {
		if t.UTC().After(w.Start.UTC()) && t.UTC().Before(w.End.UTC()) {
			return w.Name
		}
	}

	return ""
}

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
	req.Header.Set("Origin", "https://espndeportes.espn.com")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Referer", "https://espndeportes.espn.com/")
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
