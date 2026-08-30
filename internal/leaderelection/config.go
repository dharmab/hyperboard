package leaderelection

import (
	"errors"
	"fmt"
	"time"
)

// Backend identifies a leader-election implementation.
type Backend string

const (
	// Kubernetes uses a coordination.k8s.io Lease.
	Kubernetes Backend = "kubernetes"
	// Postgres uses a session-level PostgreSQL advisory lock.
	Postgres Backend = "postgres"
)

// Config configures leader election.
type Config struct {
	Backend       Backend
	Identity      string
	LockName      string
	Namespace     string
	Kubeconfig    string
	LeaseDuration time.Duration
	RenewDeadline time.Duration
	RetryPeriod   time.Duration
}

// Validate checks that the leader-election configuration is internally consistent.
func (cfg Config) Validate() error {
	if cfg.Backend != Kubernetes && cfg.Backend != Postgres {
		return fmt.Errorf("leader election backend must be %q or %q", Kubernetes, Postgres)
	}
	if cfg.LockName == "" {
		return errors.New("leader election lock name must not be empty")
	}
	if cfg.RetryPeriod <= 0 {
		return errors.New("leader election retry period must be positive")
	}
	if cfg.Backend == Kubernetes {
		if cfg.Namespace == "" {
			return errors.New("leader election namespace must not be empty")
		}
		if cfg.LeaseDuration <= cfg.RenewDeadline {
			return errors.New("leader election lease duration must be greater than renew deadline")
		}
		if cfg.RenewDeadline <= cfg.RetryPeriod {
			return errors.New("leader election renew deadline must be greater than retry period")
		}
	}
	return nil
}
