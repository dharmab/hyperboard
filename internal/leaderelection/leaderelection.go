// Package leaderelection runs work exclusively on one elected process.
package leaderelection

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sleaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var errLeadershipLost = errors.New("leadership lost")

// Task runs while this process holds leadership.
type Task func(context.Context) error

// Run elects a leader with the configured backend and runs task while this process is leader.
func Run(ctx context.Context, pool *pgxpool.Pool, cfg Config, task Task) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.Identity == "" {
		identity, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("determine leader election identity: %w", err)
		}
		cfg.Identity = identity
	}

	switch cfg.Backend {
	case Kubernetes:
		return runKubernetes(ctx, cfg, task)
	case Postgres:
		return runPostgres(ctx, pool, cfg, task)
	default:
		return fmt.Errorf("unsupported leader election backend %q", cfg.Backend)
	}
}

func runKubernetes(ctx context.Context, cfg Config, task Task) error {
	restConfig, err := kubernetesConfig(cfg.Kubeconfig)
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("create Kubernetes client: %w", err)
	}
	lock, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		cfg.Namespace,
		cfg.LockName,
		client.CoreV1(),
		client.CoordinationV1(),
		resourcelock.ResourceLockConfig{Identity: cfg.Identity},
	)
	if err != nil {
		return fmt.Errorf("create Kubernetes Lease lock: %w", err)
	}

	electionCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	taskErr := make(chan error, 1)
	elector, err := k8sleaderelection.NewLeaderElector(k8sleaderelection.LeaderElectionConfig{
		Lock:            lock,
		LeaseDuration:   cfg.LeaseDuration,
		RenewDeadline:   cfg.RenewDeadline,
		RetryPeriod:     cfg.RetryPeriod,
		ReleaseOnCancel: false,
		Name:            cfg.LockName,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: func(leaderCtx context.Context) {
				log.Info().Str("identity", cfg.Identity).Msg("Acquired Kubernetes leadership")
				taskErr <- task(leaderCtx)
			},
			OnStoppedLeading: func() {
				log.Warn().Str("identity", cfg.Identity).Msg("Lost Kubernetes leadership")
			},
			OnNewLeader: func(identity string) {
				if identity != cfg.Identity {
					log.Info().Str("identity", identity).Msg("Observed new Kubernetes leader")
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create Kubernetes leader elector: %w", err)
	}

	electionDone := make(chan struct{})
	go func() {
		elector.Run(electionCtx)
		close(electionDone)
	}()

	select {
	case err := <-taskErr:
		if err != nil {
			return err
		}
		return errors.New("leader task stopped unexpectedly")
	case <-electionDone:
		if ctx.Err() != nil {
			return nil
		}
		return errLeadershipLost
	case <-ctx.Done():
		cancel()
		<-electionDone
		return nil
	}
}

func kubernetesConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
		return cfg, nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster Kubernetes configuration: %w", err)
	}
	return cfg, nil
}

func runPostgres(ctx context.Context, pool *pgxpool.Pool, cfg Config, task Task) error {
	for ctx.Err() == nil {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil //nolint:nilerr // context cancellation is a clean shutdown
			}
			log.Warn().Err(err).Msg("Failed to acquire PostgreSQL connection for leader election")
			if !waitForRetry(ctx, cfg.RetryPeriod) {
				return nil
			}
			continue
		}

		acquired, electionErr := tryPostgresLock(ctx, conn, cfg.LockName)
		if electionErr != nil {
			log.Warn().Err(electionErr).Msg("Failed to attempt PostgreSQL leader election")
			conn.Release()
		} else if acquired {
			electionErr = holdPostgresLeadership(ctx, conn, cfg, task)
			conn.Release()
			if electionErr != nil && !errors.Is(electionErr, errLeadershipLost) {
				return electionErr
			}
		} else {
			conn.Release()
		}

		if !waitForRetry(ctx, cfg.RetryPeriod) {
			return nil
		}
	}
	return nil
}

func tryPostgresLock(ctx context.Context, conn *pgxpool.Conn, lockName string) (bool, error) {
	var acquired bool
	if err := conn.QueryRow(
		ctx,
		"SELECT pg_try_advisory_lock(hashtextextended($1, 0))",
		lockName,
	).Scan(&acquired); err != nil {
		return false, fmt.Errorf("acquire PostgreSQL advisory lock: %w", err)
	}
	return acquired, nil
}

func holdPostgresLeadership(ctx context.Context, conn *pgxpool.Conn, cfg Config, task Task) error {
	log.Info().Str("identity", cfg.Identity).Msg("Acquired PostgreSQL leadership")
	leaderCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	taskErr := make(chan error, 1)
	go func() { taskErr <- task(leaderCtx) }()

	ticker := time.NewTicker(cfg.RetryPeriod)
	defer ticker.Stop()
	for {
		select {
		case err := <-taskErr:
			if err != nil {
				return err
			}
			return errors.New("leader task stopped unexpectedly")
		case <-ticker.C:
			var one int
			if err := conn.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
				cancel()
				<-taskErr
				log.Warn().Err(err).Msg("Lost PostgreSQL leadership connection")
				return errLeadershipLost
			}
		case <-ctx.Done():
			cancel()
			<-taskErr
			releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer releaseCancel()
			//nolint:contextcheck // the parent is canceled; lock release needs a short-lived fresh context
			if _, err := conn.Exec(
				releaseCtx,
				"SELECT pg_advisory_unlock(hashtextextended($1, 0))",
				cfg.LockName,
			); err != nil {
				log.Warn().Err(err).Msg("Failed to release PostgreSQL leadership")
			}
			return nil
		}
	}
}

func waitForRetry(ctx context.Context, retryPeriod time.Duration) bool {
	timer := time.NewTimer(retryPeriod)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}
