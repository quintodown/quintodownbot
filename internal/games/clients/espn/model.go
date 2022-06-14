package espn

import (
	"strconv"
	"time"

	"github.com/quintodown/quintodownbot/internal/games"
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
