package app

import "time"

type Clock interface {
	Now() time.Time
}

type UTCClock struct{}

func NewUTCClock() UTCClock {
	return UTCClock{}
}

func (U UTCClock) Now() time.Time {
	return time.Now().UTC()
}
