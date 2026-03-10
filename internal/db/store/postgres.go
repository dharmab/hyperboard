package store

import (
	"context"

	"github.com/stephenafamo/bob"
)

// PostgresSQLStore implements the Store interface using a PostgreSQL database via Bob ORM.
type PostgresSQLStore struct {
	db                  bob.DB
	similarityThreshold int
}

var _ SQLStore = &PostgresSQLStore{}

// NewPostgresSQLStore creates a new PostgresStore backed by the given bob.DB.
func NewPostgresSQLStore(db bob.DB, similarityThreshold int) *PostgresSQLStore {
	return &PostgresSQLStore{
		db:                  db,
		similarityThreshold: similarityThreshold,
	}
}

func (s *PostgresSQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
