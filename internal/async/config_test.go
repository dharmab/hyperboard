package async

import (
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/leaderelection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	values := viper.New()
	cmd := &cobra.Command{Use: "test"}
	require.NoError(t, bindConfig(cmd, values))

	cfg, err := loadConfig(values)
	require.NoError(t, err)
	assert.Equal(t, leaderelection.Postgres, cfg.LeaderElection.Backend)
	assert.Equal(t, "hyperboard-async", cfg.LeaderElection.LockName)
	assert.Equal(t, 100*time.Millisecond, cfg.Controller.MinInterval)
	assert.Equal(t, time.Hour, cfg.Controller.MaxInterval)
}

func TestLoadConfigRejectsInvalidLeaderElectionBackend(t *testing.T) {
	t.Parallel()

	values := viper.New()
	cmd := &cobra.Command{Use: "test"}
	require.NoError(t, bindConfig(cmd, values))
	require.NoError(t, cmd.Flags().Set("leader-election", "consul"))

	_, err := loadConfig(values)
	require.EqualError(t, err, `leader election backend must be "kubernetes" or "postgres"`)
}

func TestDSNEscapesCredentials(t *testing.T) {
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
