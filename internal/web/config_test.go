package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRequiresCredentials(t *testing.T) {
	t.Parallel()

	err := (&config{}).validate()
	require.EqualError(t, err, "admin password is required")

	err = (&config{AdminPassword: "hyperboard"}).validate()
	require.EqualError(t, err, "session secret is required")

	assert.NoError(t, (&config{AdminPassword: "hyperboard", SessionSecret: "session-secret"}).validate())
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
