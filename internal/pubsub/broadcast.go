package pubsub

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

type TopicName int

const (
	ErrorTopic TopicName = iota
	PhotoTopic
	TextTopic
	GamesTopic
)

type Queue interface {
	Publish(topic string, messages ...*message.Message) error
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
	Close() error
}

//easyjson:json
type ErrorEvent struct {
	Err string `json:"error"`
}

//easyjson:json
type PhotoEvent struct {
	Caption     string `json:"caption"`
	FileID      string `json:"fileId"`
	FileURL     string `json:"fileUrl"`
	FileSize    int    `json:"fileSize"`
	FileContent []byte `json:"fileContent"`
}

//easyjson:json
type TextEvent struct {
	Text string `json:"text"`
}

//easyjson:json
type GameEvent struct {
	Id    string    `json:"id"`
	Start time.Time `json:"start"`
	Name  string    `json:"name"`
	Venue struct {
		FullName string `json:"full_name"`
		City     string `json:"city"`
		State    string `json:"state"`
		Capacity int    `json:"capacity"`
		Indoor   bool   `json:"indoor"`
	} `json:"venue"`
	Status struct {
		Clock        float64 `json:"clock"`
		DisplayClock string  `json:"display_clock"`
		Period       int     `json:"period"`
		State        string  `json:"state"`
	} `json:"status"`
	Weather struct {
		DisplayValue string `json:"display_value"`
		Temperature  int    `json:"temperature"`
	} `json:"weather"`
	HomeTeam       TeamScore `json:"home_team"`
	AwayTeam       TeamScore `json:"away_team"`
	WeekName       string    `json:"week_name"`
	Competition    string    `json:"competition"`
	LastGameChange string    `json:"last_game_change"`
}

//easyjson:json
type TeamScore struct {
	Score            int    `json:"score"`
	Name             string `json:"name"`
	ShortDisplayName string `json:"short_display_name"`
	Logo             string `json:"logo"`
	Record           string `json:"record"`
}
