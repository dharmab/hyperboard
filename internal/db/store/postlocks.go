package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"uuid"
)

// postgresPostMutationLock keeps its dedicated connection for the full lifetime
// of the session-level advisory lock.
type postgresPostMutationLock struct {
	conn *sql.Conn
	key  int64
	once sync.Once
	err  error
}

// AcquirePostMutationLock serializes mutations for postID. Advisory locks use
// a dedicated pool so blocked lock waiters cannot consume connections needed by
// the lock holder's data operations.
func (s *PostgresSQLStore) AcquirePostMutationLock(ctx context.Context, postID uuid.UUID) (PostMutationLock, error) {
	conn, err := s.postMutationLockDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection for post mutation lock: %w", err)
	}

	key := postMutationAdvisoryKey(postID)
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("acquire post mutation lock: %w", err)
	}

	return &postgresPostMutationLock{conn: conn, key: key}, nil
}

func (l *postgresPostMutationLock) Unlock(ctx context.Context) error {
	l.once.Do(func() {
		var unlocked bool
		unlockFailed := false
		if err := l.conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", l.key).Scan(&unlocked); err != nil {
			l.err = fmt.Errorf("release post mutation lock: %w", err)
			unlockFailed = true
		} else if !unlocked {
			l.err = errors.New("release post mutation lock: lock was not held by the session")
			unlockFailed = true
		}

		// A session-level advisory lock survives returning a connection to the
		// pool. If unlocking failed, mark the driver connection bad so the pool
		// discards the session instead of reusing it with a stale lock.
		if unlockFailed {
			err := l.conn.Raw(func(any) error { return driver.ErrBadConn })
			if err != nil && !errors.Is(err, driver.ErrBadConn) {
				l.err = errors.Join(l.err, fmt.Errorf("discard post mutation lock connection: %w", err))
			}
		}
		if err := l.conn.Close(); err != nil {
			l.err = errors.Join(l.err, fmt.Errorf("close post mutation lock connection: %w", err))
		}
	})
	return l.err
}

func postMutationAdvisoryKey(postID uuid.UUID) int64 {
	digest := sha256.Sum256([]byte(postID.String()))
	return int64(binary.BigEndian.Uint64(digest[:8]))
}
