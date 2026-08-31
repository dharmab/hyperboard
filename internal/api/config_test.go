package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidation(t *testing.T) {
	t.Parallel()

	require.EqualError(t, (&config{}).validate(), "admin password is required")
	assert.NoError(t, (&config{AdminPassword: "hyperboard", SimilarityThreshold: 0}).validate())
	assert.NoError(t, (&config{AdminPassword: "hyperboard", SimilarityThreshold: 64}).validate())
	require.EqualError(t, (&config{AdminPassword: "hyperboard", SimilarityThreshold: -1}).validate(), "similarity threshold must be between 0 and 64")
	require.EqualError(t, (&config{AdminPassword: "hyperboard", SimilarityThreshold: 65}).validate(), "similarity threshold must be between 0 and 64")
}

func TestSQLStoreDSNEscapesCredentials(t *testing.T) {
	t.Parallel()

	cfg := sqlStoreConfig{
		Host:     "database:5432",
		User:     "user@example.com",
		Password: "p@ss word",
		Database: "hyperboard",
		SSLMode:  "require",
	}

	assert.Equal(t, "postgres://user%40example.com:p%40ss%20word@database:5432/hyperboard?sslmode=require", cfg.dsn())
}
