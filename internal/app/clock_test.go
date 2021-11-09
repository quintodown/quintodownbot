package app_test

import (
	"testing"
	"time"

	"github.com/quintodown/quintodownbot/internal/app"
	"github.com/stretchr/testify/require"
)

func TestUTCClock_Now(t *testing.T) {
	require.WithinDuration(t, time.Now().UTC(), app.NewUTCClock().Now(), time.Minute)
}
