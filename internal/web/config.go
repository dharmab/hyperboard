package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// config holds the configuration for the web frontend server.
type config struct {
	Port                  string
	AdminPassword         string
	SessionSecret         string
	APIURL                string
	LogLevel              string
	TagFilters            []tagFilter
	QuickTag              string
	InsecureSessionCookie bool
}

// bindConfig registers CLI flags and environment variable bindings.
func bindConfig(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("port", "8080", "Port to listen on")
	flags.String("admin-password", "", "Admin password")
	flags.String("session-secret", "", "Session secret key")
	flags.String("api-url", "", "Hyperboard API URL")
	flags.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	flags.String("tag-filters", "", "Tag filter buttons as JSON array")
	flags.String("quick-tag", "", "Tag name to toggle with the 'f' hotkey")
	flags.Bool("insecure-session-cookie", false, "Allow session cookies over HTTP (local development only)")

	viper.SetEnvPrefix("HYPERBOARD_WEB")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	_ = viper.BindPFlags(flags)
	_ = cmd.MarkFlagRequired("admin-password")
	_ = cmd.MarkFlagRequired("session-secret")
}

// loadConfig reads and validates configuration values from viper.
func loadConfig() (*config, error) {
	tagFilters, err := parseTagFilters(viper.GetString("tag-filters"))
	if err != nil {
		return nil, fmt.Errorf("parsing tag-filters: %w", err)
	}
	cfg := &config{
		Port:                  viper.GetString("port"),
		AdminPassword:         viper.GetString("admin-password"),
		SessionSecret:         viper.GetString("session-secret"),
		APIURL:                viper.GetString("api-url"),
		LogLevel:              viper.GetString("log-level"),
		TagFilters:            tagFilters,
		QuickTag:              viper.GetString("quick-tag"),
		InsecureSessionCookie: viper.GetBool("insecure-session-cookie"),
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *config) validate() error {
	if cfg.AdminPassword == "" {
		return errors.New("admin password is required")
	}
	if cfg.SessionSecret == "" {
		return errors.New("session secret is required")
	}
	return nil
}

// parseTagFilters parses a JSON string into a slice of tag filter definitions.
func parseTagFilters(jsonStr string) ([]tagFilter, error) {
	if jsonStr == "" {
		return nil, nil
	}
	var filters []tagFilter
	if err := json.Unmarshal([]byte(jsonStr), &filters); err != nil {
		return nil, fmt.Errorf("invalid JSON %q: %w", jsonStr, err)
	}
	return filters, nil
}
