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
	Id             string      `json:"id"`
	Start          time.Time   `json:"start"`
	Name           string      `json:"name"`
	Venue          GameVenue   `json:"venue"`
	Status         GameStatus  `json:"status"`
	Weather        GameWeather `json:"weather"`
	HomeTeam       TeamScore   `json:"homeTeam"`
	AwayTeam       TeamScore   `json:"awayTeam"`
	WeekName       string      `json:"weekName"`
	Competition    string      `json:"competition"`
	LastGameChange string      `json:"lastGameChange"`
}

//easyjson:json
type TeamScore struct {
	Score            int    `json:"score"`
	Name             string `json:"name"`
	ShortDisplayName string `json:"shortDisplayName"`
	Logo             string `json:"logo"`
	Record           string `json:"record"`
}

//easyjson:json
type GameVenue struct {
	FullName string `json:"fullName"`
	City     string `json:"city"`
	State    string `json:"state"`
	Capacity int    `json:"capacity"`
	Indoor   bool   `json:"indoor"`
}

//easyjson:json
type GameStatus struct {
	Clock        float64 `json:"clock"`
	DisplayClock string  `json:"displayClock"`
	Period       int     `json:"period"`
	State        string  `json:"state"`
}

//easyjson:json
type GameWeather struct {
	DisplayValue string `json:"displayValue"`
	Temperature  int    `json:"temperature"`
}
