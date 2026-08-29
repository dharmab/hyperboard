package store

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapConflictError(t *testing.T) {
	t.Parallel()

	for _, constraint := range []string{
		"tags_name_key",
		"tag_categories_name_key",
		"tag_aliases_alias_name_key",
	} {
		t.Run(constraint, func(t *testing.T) {
			t.Parallel()
			pgErr := &pgconn.PgError{
				Code:           uniqueViolationSQLState,
				ConstraintName: constraint,
				Message:        "duplicate key value violates unique constraint",
			}
			wrapped := fmt.Errorf("database operation failed: %w", pgErr)

			got := mapConflictError(wrapped)

			require.Error(t, got)
			require.ErrorIs(t, got, ErrConflict)
			require.ErrorIs(t, got, pgErr)
			assert.Contains(t, got.Error(), pgErr.Message)
		})
	}
}

func TestMapConflictErrorLeavesUnrelatedErrorsUnchanged(t *testing.T) {
	t.Parallel()

	ordinaryErr := errors.New("ordinary failure")
	unrelatedUniqueViolation := &pgconn.PgError{
		Code:           uniqueViolationSQLState,
		ConstraintName: "future_unique_constraint",
	}
	nonUniqueViolation := &pgconn.PgError{
		Code:           "23503",
		ConstraintName: "tags_name_key",
	}

	require.NoError(t, mapConflictError(nil))
	assert.Same(t, ordinaryErr, mapConflictError(ordinaryErr))
	assert.Same(t, unrelatedUniqueViolation, mapConflictError(unrelatedUniqueViolation))
	assert.Same(t, nonUniqueViolation, mapConflictError(nonUniqueViolation))
	assert.NotErrorIs(t, mapConflictError(unrelatedUniqueViolation), ErrConflict)
}
