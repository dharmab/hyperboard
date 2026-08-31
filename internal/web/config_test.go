package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidation(t *testing.T) {
	t.Parallel()

	validConfig := config{
		AdminPassword: "hyperboard",
		SessionSecret: "session-secret",
		APIURL:        "https://api.example.com",
	}

	tests := []struct {
		name      string
		mutate    func(*config)
		errString string
	}{
		{
			name: "valid HTTP URL",
			mutate: func(cfg *config) {
				cfg.APIURL = "http://localhost:8081/api"
			},
		},
		{name: "valid HTTPS URL"},
		{
			name: "missing admin password",
			mutate: func(cfg *config) {
				cfg.AdminPassword = ""
			},
			errString: "admin password is required",
		},
		{
			name: "missing session secret",
			mutate: func(cfg *config) {
				cfg.SessionSecret = ""
			},
			errString: "session secret is required",
		},
		{
			name: "missing API URL",
			mutate: func(cfg *config) {
				cfg.APIURL = ""
			},
			errString: "API URL is required",
		},
		{
			name: "relative API URL",
			mutate: func(cfg *config) {
				cfg.APIURL = "/api"
			},
			errString: "API URL must be an absolute HTTP or HTTPS URL with a host",
		},
		{
			name: "HTTP URL without host",
			mutate: func(cfg *config) {
				cfg.APIURL = "http:/api"
			},
			errString: "API URL must be an absolute HTTP or HTTPS URL with a host",
		},
		{
			name: "API URL with port but no hostname",
			mutate: func(cfg *config) {
				cfg.APIURL = "http://:8081"
			},
			errString: "API URL must be an absolute HTTP or HTTPS URL with a host",
		},
		{
			name: "unsupported scheme",
			mutate: func(cfg *config) {
				cfg.APIURL = "ftp://api.example.com"
			},
			errString: "API URL must be an absolute HTTP or HTTPS URL with a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}

			err := cfg.validate()
			if tt.errString == "" {
				assert.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.errString)
		})
	}
}

func TestParseTagFilters(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON", func(t *testing.T) {
		t.Parallel()
		input := `[{"label":"Colors","tags":["blue","green"]},{"label":"Favorites","tags":["favorite","-archived"]}]`
		filters, err := parseTagFilters(input)
		require.NoError(t, err)
		require.Len(t, filters, 2)
		colorFilter := filters[0]
		favoritesFilter := filters[1]
		assert.Equal(t, "Colors", colorFilter.Label)
		assert.Equal(t, []string{"blue", "green"}, colorFilter.Tags)
		assert.Equal(t, "Favorites", favoritesFilter.Label)
		assert.Equal(t, []string{"favorite", "-archived"}, favoritesFilter.Tags)
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		filters, err := parseTagFilters("")
		require.NoError(t, err)
		assert.Nil(t, filters)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		t.Parallel()
		_, err := parseTagFilters("{bad json")
		require.Error(t, err)
	})
}
