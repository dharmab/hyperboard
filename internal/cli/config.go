package cli

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds CLI configuration for API access.
type Config struct {
	APIURL        string
	AdminPassword string
}

// bindConfig registers persistent CLI flags and environment variable bindings.
func bindConfig(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()

	flags.String("api-url", "", "Hyperboard API URL")
	flags.String("admin-password", "", "Admin password")

	viper.SetEnvPrefix("HYPERBOARDCTL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	_ = viper.BindPFlags(flags)
}

// loadConfig reads and validates configuration values from viper.
func loadConfig() (*Config, error) {
	cfg := &Config{
		APIURL:        viper.GetString("api-url"),
		AdminPassword: viper.GetString("admin-password"),
	}

	if cfg.APIURL == "" {
		return nil, errors.New("API URL is required (set --api-url flag, HYPERBOARDCTL_API_URL env var, or api-url in config file)")
	}
	if cfg.AdminPassword == "" {
		return nil, errors.New("admin password is required (set --admin-password flag, HYPERBOARDCTL_ADMIN_PASSWORD env var, or admin-password in config file)")
	}

	return cfg, nil
}
