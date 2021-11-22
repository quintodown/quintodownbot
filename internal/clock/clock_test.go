package clock_test

import (
	"testing"
	"time"

	"github.com/quintodown/quintodownbot/internal/clock"

	"github.com/stretchr/testify/require"
)

func TestUTCClock_Now(t *testing.T) {
	require.WithinDuration(t, time.Now().UTC(), clock.NewUTCClock().Now(), time.Minute)
}
