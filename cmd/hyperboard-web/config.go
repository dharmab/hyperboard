package main

import (
	"encoding/json"
	"strings"

	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	AdminPassword string
	SessionSecret string
	APIURL        string
	LogLevel      string
	TagFilters    []TagFilter
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

func loadConfig() (*Config, error) {
	tagFilters, err := parseTagFilters(viper.GetString("tag-filters"))
	if err != nil {
		return nil, fmt.Errorf("parsing tag-filters: %w", err)
	}
	return &Config{
		Port:          viper.GetString("port"),
		AdminPassword: viper.GetString("admin-password"),
		SessionSecret: viper.GetString("session-secret"),
		APIURL:        viper.GetString("api-url"),
		LogLevel:      viper.GetString("log-level"),
		TagFilters:    tagFilters,
	}, nil
}

func parseTagFilters(jsonStr string) ([]TagFilter, error) {
	if jsonStr == "" {
		return nil, nil
	}
	var filters []TagFilter
	if err := json.Unmarshal([]byte(jsonStr), &filters); err != nil {
		return nil, fmt.Errorf("invalid JSON %q: %w", jsonStr, err)
	}
	return filters, nil
}
