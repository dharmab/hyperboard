// Package leaderelection runs work exclusively on one elected process.
package leaderelection

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var errLeadershipLost = errors.New("leadership lost")

// Task runs while this process holds leadership.
type Task func(context.Context) error

// Leader elects a leader and runs work while this process holds leadership.
type Leader interface {
	Lead(context.Context, Task) error
}

func prepareConfig(cfg Config) (Config, error) {
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	if cfg.Identity == "" {
		identity, err := os.Hostname()
		if err != nil {
			return cfg, fmt.Errorf("determine leader election identity: %w", err)
		}
		cfg.Identity = identity
	}
	return cfg, nil
}
