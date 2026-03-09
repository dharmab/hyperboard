package main

import (
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	AdminPassword string
	SessionSecret string
	APIURL        string
	LogLevel      string
	TagFilters    string
}

func bindConfig(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("port", "8080", "Port to listen on")
	flags.String("admin-password", "", "Admin password")
	flags.String("session-secret", "", "Session secret key")
	flags.String("api-url", "", "Hyperboard API URL")
	flags.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	flags.String("tag-filters", "", "Tag filter buttons as JSON array")

	viper.SetEnvPrefix("HYPERBOARD_WEB")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	_ = viper.BindPFlags(flags)
}

func loadConfig() *Config {
	return &Config{
		Port:          viper.GetString("port"),
		AdminPassword: viper.GetString("admin-password"),
		SessionSecret: viper.GetString("session-secret"),
		APIURL:        viper.GetString("api-url"),
		LogLevel:      viper.GetString("log-level"),
		TagFilters:    viper.GetString("tag-filters"),
	}
}

func parseTagFilters(jsonStr string) []TagFilter {
	if jsonStr == "" {
		return nil
	}
	var filters []TagFilter
	if err := json.Unmarshal([]byte(jsonStr), &filters); err != nil {
		log.Warn().Err(err).Msg("Failed to parse tag-filters JSON")
		return nil
	}
	return filters
}
