package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRequiresAdminPassword(t *testing.T) {
	t.Parallel()

	err := (&config{}).validate()
	require.EqualError(t, err, "admin password is required")
	assert.NoError(t, (&config{AdminPassword: "hyperboard"}).validate())
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
