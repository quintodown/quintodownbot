package games

import (
	"github.com/quintodown/quintodownbot/internal/clock"
	"time"
)

const (
	NoChanges GameChange = iota
	NewGame
	Started
	Rescheduled
	HomeScore
	AwayScore
	PeriodFinished
	Finished
)

const (
	NFL Competition = iota
	CFL
	NCAA
)

const (
	ScheduledState GameState = iota
	RescheduledState
	InProgressState
	FinishedState
	CancelledState
)

type GameChange int

type Competition int

type GameState int

func GetCompetitions() []Competition {
	return []Competition{NFL}
}

type Game struct {
	Id          string
	Start       time.Time
	Name        string
	Venue       Venue
	Status      GameStatus
	Weather     GameWeather
	HomeTeam    TeamScore
	AwayTeam    TeamScore
	WeekName    string
	Competition Competition
}

type Week struct {
	Name  string
	Start time.Time
	End   time.Time
}

type TeamScore struct {
	Score            int
	Name             string
	ShortDisplayName string
	Logo             string
	Record           string
}

type Venue struct {
	FullName string
	Address  VenueAddress
	Capacity int
	Indoor   bool
}

type VenueAddress struct {
	City  string
	State string
}

type GameStatus struct {
	Clock        float64
	DisplayClock string
	Period       int
	State        GameState
}

type GameWeather struct {
	DisplayValue string
	Temperature  int
}

func (g *Game) isGameInProgress(clk clock.Clock) bool {
	return clk.Now().After(g.Start.UTC()) && !g.hasFinishedGame()
}

func (g *Game) hasFinishedGame() bool {
	return g.Status.State == FinishedState
}
