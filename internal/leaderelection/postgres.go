package leaderelection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type postgresLeader struct {
	pool *pgxpool.Pool
	cfg  Config
}

var _ Leader = (*postgresLeader)(nil)

// NewPostgres constructs a PostgreSQL leader using the supplied connection pool.
func NewPostgres(pool *pgxpool.Pool, cfg Config) (Leader, error) {
	cfg, err := prepareConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &postgresLeader{pool: pool, cfg: cfg}, nil
}

func (l *postgresLeader) Lead(ctx context.Context, task Task) error {
	cfg := l.cfg
	pool := l.pool
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
				if err := <-taskErr; err != nil && !errors.Is(err, context.Canceled) {
					log.Warn().Err(err).Msg("task failed during stop")
				}
				log.Warn().Err(err).Msg("Lost PostgreSQL leadership connection")
				return errLeadershipLost
			}
		case <-ctx.Done():
			cancel()
			if err := <-taskErr; err != nil && !errors.Is(err, context.Canceled) {
				log.Warn().Err(err).Msg("task failed during stop")
			}
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
