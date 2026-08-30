// Package heartbeat runs controller reconciliation on an adaptive interval.
package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ReconcileFunc performs one reconciliation pass and reports whether it found work.
type ReconcileFunc func(context.Context) (bool, error)

// Run calls reconcile at an adaptive interval. Work speeds up the next
// heartbeat; an idle pass backs it off, up to maxInterval.
func Run(
	ctx context.Context,
	minInterval time.Duration,
	maxInterval time.Duration,
	reconcile ReconcileFunc,
) error {
	if minInterval <= 0 {
		return errors.New("minimum heartbeat interval must be positive")
	}
	if maxInterval < minInterval {
		return errors.New("maximum heartbeat interval must be greater than or equal to minimum interval")
	}
	if reconcile == nil {
		return errors.New("heartbeat reconcile function must not be nil")
	}

	interval := minInterval
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			didWork, err := reconcile(ctx)
			if err != nil {
				return fmt.Errorf("reconcile controller: %w", err)
			}
			interval = nextInterval(interval, minInterval, maxInterval, didWork)
			timer.Reset(interval)
		}
	}
}

func nextInterval(current, minInterval, maxInterval time.Duration, didWork bool) time.Duration {
	if didWork {
		if current <= minInterval*2 {
			return minInterval
		}
		return current / 2
	}
	if current >= maxInterval/2 {
		return maxInterval
	}
	return current * 2
}
