package store

import (
	"context"
	"database/sql"
)

// PostgresSQLStore implements the Store interface using a PostgreSQL database.
type PostgresSQLStore struct {
	db                  *sql.DB
	postMutationLockDB  *sql.DB
	similarityThreshold int
}

var _ SQLStore = &PostgresSQLStore{}

// NewPostgresSQLStore creates a new PostgresStore backed by the given data and
// advisory-lock connection pools.
func NewPostgresSQLStore(db, postMutationLockDB *sql.DB, similarityThreshold int) *PostgresSQLStore {
	return &PostgresSQLStore{
		db:                  db,
		postMutationLockDB:  postMutationLockDB,
		similarityThreshold: similarityThreshold,
	}
}

// Ping checks database connectivity.
func (s *PostgresSQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
