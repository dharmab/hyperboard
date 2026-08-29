package store

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationSQLState = "23505"

// mapConflictError wraps relevant PostgreSQL unique-constraint violations
// with ErrConflict while preserving the original database error in the chain.
func mapConflictError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolationSQLState {
		return err
	}
	switch pgErr.ConstraintName {
	case "tag_aliases_alias_name_key", "tag_categories_name_key", "tags_name_key":
		return fmt.Errorf("%w: %w", ErrConflict, err)
	default:
		return err
	}
}
