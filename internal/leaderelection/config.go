package leaderelection

import (
	"errors"
	"time"
)

// Config configures leader election.
type Config struct {
	Identity    string
	LockName    string
	RetryPeriod time.Duration
}

// Validate checks that the leader-election configuration is internally consistent.
func (cfg Config) Validate() error {
	if cfg.LockName == "" {
		return errors.New("leader election lock name must not be empty")
	}
	if cfg.RetryPeriod <= 0 {
		return errors.New("leader election retry period must be positive")
	}
	return nil
}
