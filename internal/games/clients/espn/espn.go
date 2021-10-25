package espn

import (
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/quintodown/quintodownbot/internal/app"
	"github.com/quintodown/quintodownbot/internal/games"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	endpointForCalendar = "https://site.api.espn.com/apis/site/v2/sports/football/%s/scoreboard"
	endpointForGames    = "https://site.api.espn.com/apis/site/v2/sports/football/%s/summary"
	timeLayout          = "2006-01-02T15:04Z"
	statusFinal         = "STATUS_FINAL"
	statusInProgress    = "STATUS_IN_PROGRESS"
)

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
			Statistics []struct {
				Name         string `json:"name"`
				DisplayValue string `json:"displayValue"`
				Label        string `json:"label"`
			} `json:"statistics"`
		} `json:"teams"`
		Players []struct {
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
			Statistics []struct {
				Name         string   `json:"name"`
				Text         string   `json:"text"`
				Labels       []string `json:"labels"`
				Descriptions []string `json:"descriptions"`
				Athletes     []struct {
					Athlete struct {
						ID          string `json:"id"`
						UID         string `json:"uid"`
						GUID        string `json:"guid"`
						FirstName   string `json:"firstName"`
						LastName    string `json:"lastName"`
						DisplayName string `json:"displayName"`
						Links       []struct {
							Href string `json:"href"`
							Text string `json:"text"`
						} `json:"links"`
					} `json:"athlete"`
					Stats []string `json:"stats"`
				} `json:"athletes"`
				Totals []string `json:"totals"`
			} `json:"statistics"`
		} `json:"players"`
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
			Images   []struct {
				Href   string   `json:"href"`
				Width  int      `json:"width"`
				Height int      `json:"height"`
				Alt    string   `json:"alt"`
				Rel    []string `json:"rel"`
			} `json:"images"`
		} `json:"venue"`
		Attendance int `json:"attendance"`
		Weather    struct {
			Temperature   int    `json:"temperature"`
			ConditionID   string `json:"conditionId"`
			Gust          int    `json:"gust"`
			Precipitation int    `json:"precipitation"`
		} `json:"weather"`
	} `json:"gameInfo"`
	Drives struct {
		Previous []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Team        struct {
				Name             string `json:"name"`
				Abbreviation     string `json:"abbreviation"`
				DisplayName      string `json:"displayName"`
				ShortDisplayName string `json:"shortDisplayName"`
				Logos            []struct {
					Href   string   `json:"href"`
					Width  int      `json:"width"`
					Height int      `json:"height"`
					Alt    string   `json:"alt"`
					Rel    []string `json:"rel"`
				} `json:"logos"`
			} `json:"team"`
			Start struct {
				Period struct {
					Type   string `json:"type"`
					Number int    `json:"number"`
				} `json:"period"`
				Clock struct {
					DisplayValue string `json:"displayValue"`
				} `json:"clock"`
				YardLine int    `json:"yardLine"`
				Text     string `json:"text"`
			} `json:"start"`
			End struct {
				Period struct {
					Type   string `json:"type"`
					Number int    `json:"number"`
				} `json:"period"`
				Clock struct {
					DisplayValue string `json:"displayValue"`
				} `json:"clock"`
				YardLine int    `json:"yardLine"`
				Text     string `json:"text"`
			} `json:"end,omitempty"`
			TimeElapsed struct {
				DisplayValue string `json:"displayValue"`
			} `json:"timeElapsed"`
			Yards              int    `json:"yards"`
			IsScore            bool   `json:"isScore"`
			OffensivePlays     int    `json:"offensivePlays"`
			Result             string `json:"result"`
			ShortDisplayResult string `json:"shortDisplayResult"`
			DisplayResult      string `json:"displayResult"`
			Plays              []struct {
				ID   string `json:"id"`
				Type struct {
					ID           string `json:"id"`
					Text         string `json:"text"`
					Abbreviation string `json:"abbreviation"`
				} `json:"type,omitempty"`
				Text      string `json:"text"`
				AwayScore int    `json:"awayScore"`
				HomeScore int    `json:"homeScore"`
				Period    struct {
					Number int `json:"number"`
				} `json:"period"`
				Clock struct {
					DisplayValue string `json:"displayValue"`
				} `json:"clock"`
				ScoringPlay bool      `json:"scoringPlay"`
				Priority    bool      `json:"priority"`
				Modified    string    `json:"modified"`
				Wallclock   time.Time `json:"wallclock"`
				End         struct {
					Down                  int    `json:"down"`
					Distance              int    `json:"distance"`
					YardLine              int    `json:"yardLine"`
					YardsToEndzone        int    `json:"yardsToEndzone"`
					DownDistanceText      string `json:"downDistanceText"`
					ShortDownDistanceText string `json:"shortDownDistanceText"`
					PossessionText        string `json:"possessionText"`
					Team                  struct {
						ID string `json:"id"`
					} `json:"team"`
				} `json:"end"`
				StatYardage int `json:"statYardage"`
				Start       struct {
					Down                  int    `json:"down"`
					Distance              int    `json:"distance"`
					YardLine              int    `json:"yardLine"`
					YardsToEndzone        int    `json:"yardsToEndzone"`
					DownDistanceText      string `json:"downDistanceText"`
					ShortDownDistanceText string `json:"shortDownDistanceText"`
					PossessionText        string `json:"possessionText"`
					Team                  struct {
						ID string `json:"id"`
					} `json:"team"`
				} `json:"start,omitempty"`
			} `json:"plays"`
		} `json:"previous"`
	} `json:"drives"`
	Leaders []struct {
		Team struct {
			ID           string `json:"id"`
			UID          string `json:"uid"`
			DisplayName  string `json:"displayName"`
			Abbreviation string `json:"abbreviation"`
			Links        []struct {
				Href string `json:"href"`
				Text string `json:"text"`
			} `json:"links"`
			Logo string `json:"logo"`
		} `json:"team"`
		Leaders []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			Leaders     []struct {
				DisplayValue string `json:"displayValue"`
				Athlete      struct {
					ID          string `json:"id"`
					UID         string `json:"uid"`
					GUID        string `json:"guid"`
					LastName    string `json:"lastName"`
					FullName    string `json:"fullName"`
					DisplayName string `json:"displayName"`
					ShortName   string `json:"shortName"`
					Links       []struct {
						Rel  []string `json:"rel"`
						Href string   `json:"href"`
						Text string   `json:"text"`
					} `json:"links"`
					Headshot struct {
						Href string `json:"href"`
						Alt  string `json:"alt"`
					} `json:"headshot"`
					Jersey   string `json:"jersey"`
					Position struct {
						Abbreviation string `json:"abbreviation"`
					} `json:"position"`
					Team struct {
						Ref string `json:"$ref"`
					} `json:"team"`
				} `json:"athlete"`
			} `json:"leaders"`
		} `json:"leaders"`
	} `json:"leaders"`
	Header struct {
		ID     string `json:"id"`
		UID    string `json:"uid"`
		Season struct {
			Year int `json:"year"`
			Type int `json:"type"`
		} `json:"season"`
		TimeValid    bool `json:"timeValid"`
		Competitions []struct {
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
					Links []struct {
						Rel  []string `json:"rel"`
						Href string   `json:"href"`
						Text string   `json:"text"`
					} `json:"links"`
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
			Broadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"broadcasts"`
		} `json:"competitions"`
		Links []struct {
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Week   int `json:"week"`
		League struct {
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
	} `json:"header"`
	ScoringPlays []struct {
		ID   string `json:"id"`
		Type struct {
			ID           string `json:"id"`
			Text         string `json:"text"`
			Abbreviation string `json:"abbreviation"`
		} `json:"type"`
		Text      string `json:"text"`
		AwayScore int    `json:"awayScore"`
		HomeScore int    `json:"homeScore"`
		Period    struct {
			Number int `json:"number"`
		} `json:"period"`
		Clock struct {
			Value        float64 `json:"value"`
			DisplayValue string  `json:"displayValue"`
		} `json:"clock"`
		Team struct {
			ID           string `json:"id"`
			UID          string `json:"uid"`
			DisplayName  string `json:"displayName"`
			Abbreviation string `json:"abbreviation"`
			Links        []struct {
				Href string `json:"href"`
				Text string `json:"text"`
			} `json:"links"`
			Logo string `json:"logo"`
		} `json:"team"`
		ScoringType struct {
			Name         string `json:"name"`
			DisplayName  string `json:"displayName"`
			Abbreviation string `json:"abbreviation"`
		} `json:"scoringType"`
	} `json:"scoringPlays"`
	Standings struct {
		FullViewLink struct {
			Text string `json:"text"`
			Href string `json:"href"`
		} `json:"fullViewLink"`
		Groups []struct {
			Standings struct {
				Entries []struct {
					Team  string `json:"team"`
					Link  string `json:"link"`
					ID    string `json:"id"`
					UID   string `json:"uid"`
					Stats []struct {
						Name             string  `json:"name"`
						DisplayName      string  `json:"displayName,omitempty"`
						ShortDisplayName string  `json:"shortDisplayName,omitempty"`
						Description      string  `json:"description,omitempty"`
						Abbreviation     string  `json:"abbreviation"`
						Type             string  `json:"type"`
						Value            float64 `json:"value,omitempty"`
						DisplayValue     string  `json:"displayValue"`
						ID               string  `json:"id,omitempty"`
						Summary          string  `json:"summary,omitempty"`
					} `json:"stats"`
					Logo []struct {
						Href   string   `json:"href"`
						Width  int      `json:"width"`
						Height int      `json:"height"`
						Alt    string   `json:"alt"`
						Rel    []string `json:"rel"`
					} `json:"logo"`
				} `json:"entries"`
			} `json:"standings"`
			Header string `json:"header"`
			Href   string `json:"href"`
		} `json:"groups"`
	} `json:"standings"`
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
	Events []struct {
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
		Competitions []struct {
			ID         string `json:"id"`
			UID        string `json:"uid"`
			Date       string `json:"date"`
			Attendance int    `json:"attendance"`
			Type       struct {
				ID           string `json:"id"`
				Abbreviation string `json:"abbreviation"`
			} `json:"type"`
			TimeValid             bool `json:"timeValid"`
			NeutralSite           bool `json:"neutralSite"`
			ConferenceCompetition bool `json:"conferenceCompetition"`
			Recent                bool `json:"recent"`
			Venue                 struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Address  struct {
					City  string `json:"city"`
					State string `json:"state"`
				} `json:"address"`
				Capacity int  `json:"capacity"`
				Indoor   bool `json:"indoor"`
			} `json:"venue"`
			Competitors []struct {
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
					Links []struct {
						Rel        []string `json:"rel"`
						Href       string   `json:"href"`
						Text       string   `json:"text"`
						IsExternal bool     `json:"isExternal"`
						IsPremium  bool     `json:"isPremium"`
					} `json:"links"`
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
			Notes  []interface{} `json:"notes"`
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
			Broadcasts []struct {
				Market string   `json:"market"`
				Names  []string `json:"names"`
			} `json:"broadcasts"`
			Tickets []struct {
				Summary         string `json:"summary"`
				NumberAvailable int    `json:"numberAvailable"`
				Links           []struct {
					Href string `json:"href"`
				} `json:"links"`
			} `json:"tickets"`
			StartDate     string `json:"startDate"`
			GeoBroadcasts []struct {
				Type struct {
					ID        string `json:"id"`
					ShortName string `json:"shortName"`
				} `json:"type"`
				Market struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"market"`
				Media struct {
					ShortName string `json:"shortName"`
				} `json:"media"`
				Lang   string `json:"lang"`
				Region string `json:"region"`
			} `json:"geoBroadcasts"`
			Odds []struct {
				Provider struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Priority int    `json:"priority"`
				} `json:"provider"`
				Details   string  `json:"details"`
				OverUnder float64 `json:"overUnder"`
			} `json:"odds"`
		} `json:"competitions"`
		Links []struct {
			Language   string   `json:"language"`
			Rel        []string `json:"rel"`
			Href       string   `json:"href"`
			Text       string   `json:"text"`
			ShortText  string   `json:"shortText"`
			IsExternal bool     `json:"isExternal"`
			IsPremium  bool     `json:"isPremium"`
		} `json:"links"`
		Weather struct {
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
	} `json:"events"`
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
			Temperature: (v.GameInfo.Weather.Temperature - 32) * 5 / 9,
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
	var found []games.Game
	for _, v := range v.Events {
		t, err := time.Parse(timeLayout, v.Date)
		if err != nil {
			continue
		}
		g := games.Game{
			Id:    v.ID,
			Name:  v.Name,
			Start: t.UTC(),
			Venue: games.Venue{
				FullName: v.Competitions[0].Venue.FullName,
				Address: games.VenueAddress{
					City:  v.Competitions[0].Venue.Address.City,
					State: v.Competitions[0].Venue.Address.State,
				},
				Capacity: v.Competitions[0].Venue.Capacity,
				Indoor:   v.Competitions[0].Venue.Indoor,
			},
			Status: games.GameStatus{
				Clock:        v.Status.Clock,
				DisplayClock: v.Status.DisplayClock,
				Period:       v.Status.Period,
				State:        getGameStatus(v.Competitions[0].Status.Type.Name),
			},
			Weather: games.GameWeather{
				DisplayValue: v.Weather.DisplayValue,
				Temperature:  v.Weather.Temperature,
			},
			Competition: c,
		}

		for _, w := range calendar {
			if t.UTC().After(w.Start.UTC()) && t.UTC().Before(w.End.UTC()) {
				g.WeekName = w.Name
			}
		}

		for _, v := range v.Competitions[0].Competitors {
			record := ""
			if len(v.Records) > 0 {
				record = v.Records[0].Summary
			}

			score, _ := strconv.Atoi(v.Score)
			t := games.TeamScore{
				Score:            score,
				Name:             v.Team.DisplayName,
				ShortDisplayName: v.Team.ShortDisplayName,
				Logo:             v.Team.Logo,
				Record:           record,
			}

			if v.HomeAway == "home" {
				g.HomeTeam = t
			} else {
				g.AwayTeam = t
			}
		}

		found = append(found, g)
	}

	return found
}

type Client struct {
	client *http.Client
	clk    app.Clock
}

func NewESPNClient(c *http.Client, clk app.Clock) games.GameInfoClient {
	return &Client{client: c, clk: clk}
}

func (ec *Client) GetGames(competition games.Competition) ([]games.Game, error) {
	const datesLayout = "20060102"
	var scb scoreboard

	if err := ec.executeCall(endpointForCalendar, competition, map[string]string{}, &scb); err != nil {
		return nil, err
	}

	calendar := scb.toCalendar()
	params := map[string]string{}
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

func (ec Client) executeCall(endpoint string, c games.Competition, parameters map[string]string, response easyjson.Unmarshaler) error {
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

func (ec *Client) getRequest(endpoint string, c games.Competition, parameters map[string]string) (*http.Request, error) {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
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