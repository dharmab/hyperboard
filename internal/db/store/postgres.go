package store

import (
	"context"
	"database/sql"
)

// PostgresSQLStore implements the Store interface using a PostgreSQL database.
type PostgresSQLStore struct {
	db                  *sql.DB
	similarityThreshold int
}

var _ SQLStore = &PostgresSQLStore{}

// NewPostgresSQLStore creates a new PostgresStore backed by the given *sql.DB.
func NewPostgresSQLStore(db *sql.DB, similarityThreshold int) *PostgresSQLStore {
	return &PostgresSQLStore{
		db:                  db,
		similarityThreshold: similarityThreshold,
	}
}

// Ping checks database connectivity.
func (s *PostgresSQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
