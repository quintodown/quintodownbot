package app_test

import (
	"github.com/quintodown/quintodownbot/internal/app"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUTCClock_Now(t *testing.T) {
	require.WithinDuration(t, time.Now().UTC(), app.NewUTCClock().Now(), time.Minute)
}
