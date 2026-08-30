package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextInterval(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		current time.Duration
		didWork bool
		want    time.Duration
	}{
		"idle backs off":       {current: 2 * time.Second, want: 4 * time.Second},
		"idle caps at maximum": {current: 8 * time.Second, want: 10 * time.Second},
		"work speeds up":       {current: 8 * time.Second, didWork: true, want: 4 * time.Second},
		"work caps at minimum": {current: 2 * time.Second, didWork: true, want: time.Second},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, nextInterval(test.current, time.Second, 10*time.Second, test.didWork))
		})
	}
}

func TestRunStopsWithContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := Run(ctx, time.Hour, time.Hour, func(context.Context) (bool, error) {
		require.Fail(t, "reconcile should not be called")
		return false, nil
	})
	require.NoError(t, err)
}
