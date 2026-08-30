package heartbeat

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWaitsForMinimumInterval(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		const interval = 100 * time.Millisecond
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		calls := make(chan struct{}, 1)
		go func() {
			done <- Run(ctx, interval, time.Hour, func(context.Context) (bool, error) {
				calls <- struct{}{}
				return false, nil
			})
		}()

		synctest.Wait()
		assertNoCall(t, calls)
		time.Sleep(interval - time.Nanosecond)
		synctest.Wait()
		assertNoCall(t, calls)
		time.Sleep(time.Nanosecond)
		synctest.Wait()
		requireCall(t, calls)

		cancel()
		require.NoError(t, <-done)
	})
}

func TestRunBacksOffWhileIdleAndCapsAtMaximum(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		const (
			minInterval = 100 * time.Millisecond
			maxInterval = 800 * time.Millisecond
		)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		start := time.Now()
		calls := make(chan time.Duration, 5)
		go func() {
			done <- Run(ctx, minInterval, maxInterval, func(context.Context) (bool, error) {
				calls <- time.Since(start)
				return false, nil
			})
		}()

		time.Sleep(2300 * time.Millisecond)
		assert.Equal(t, []time.Duration{
			100 * time.Millisecond,
			300 * time.Millisecond,
			700 * time.Millisecond,
			1500 * time.Millisecond,
			2300 * time.Millisecond,
		}, receiveCalls(calls, 5))

		cancel()
		require.NoError(t, <-done)
	})
}

func TestRunSpeedsUpWhenWorkIsFoundAndCapsAtMinimum(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		const (
			minInterval = 100 * time.Millisecond
			maxInterval = 800 * time.Millisecond
		)
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		start := time.Now()
		calls := make(chan time.Duration, 6)
		go func() {
			callCount := 0
			done <- Run(ctx, minInterval, maxInterval, func(context.Context) (bool, error) {
				callCount++
				calls <- time.Since(start)
				return callCount >= 3, nil
			})
		}()

		time.Sleep(1100 * time.Millisecond)
		assert.Equal(t, []time.Duration{
			100 * time.Millisecond,
			300 * time.Millisecond,
			700 * time.Millisecond,
			900 * time.Millisecond,
			1000 * time.Millisecond,
			1100 * time.Millisecond,
		}, receiveCalls(calls, 6))

		cancel()
		require.NoError(t, <-done)
	})
}

func TestRunReturnsReconcileError(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		reconcileErr := errors.New("failed")
		done := make(chan error, 1)
		go func() {
			done <- Run(t.Context(), time.Second, time.Hour, func(context.Context) (bool, error) {
				return false, reconcileErr
			})
		}()

		time.Sleep(time.Second)
		err := <-done
		require.ErrorIs(t, err, reconcileErr)
		assert.EqualError(t, err, "reconcile controller: failed")
	})
}

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		calls := make(chan struct{}, 1)
		go func() {
			done <- Run(ctx, time.Hour, time.Hour, func(context.Context) (bool, error) {
				calls <- struct{}{}
				return false, nil
			})
		}()

		synctest.Wait()
		cancel()
		require.NoError(t, <-done)
		assertNoCall(t, calls)
	})
}

func TestRunValidatesArguments(t *testing.T) {
	t.Parallel()

	reconcile := func(context.Context) (bool, error) { return false, nil }
	tests := map[string]struct {
		min       time.Duration
		max       time.Duration
		reconcile ReconcileFunc
		want      string
	}{
		"non-positive minimum": {
			max:       time.Second,
			reconcile: reconcile,
			want:      "minimum heartbeat interval must be positive",
		},
		"maximum below minimum": {
			min:       time.Second,
			max:       time.Millisecond,
			reconcile: reconcile,
			want:      "maximum heartbeat interval must be greater than or equal to minimum interval",
		},
		"nil reconcile": {
			min:  time.Second,
			max:  time.Hour,
			want: "heartbeat reconcile function must not be nil",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := Run(t.Context(), test.min, test.max, test.reconcile)
			require.EqualError(t, err, test.want)
		})
	}
}

func assertNoCall(t *testing.T, calls <-chan struct{}) {
	t.Helper()
	select {
	case <-calls:
		t.Fatal("unexpected reconciliation")
	default:
	}
}

func requireCall(t *testing.T, calls <-chan struct{}) {
	t.Helper()
	select {
	case <-calls:
	default:
		t.Fatal("expected reconciliation")
	}
}

func receiveCalls(calls <-chan time.Duration, count int) []time.Duration {
	result := make([]time.Duration, 0, count)
	for range count {
		result = append(result, <-calls)
	}
	return result
}
