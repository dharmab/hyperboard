package cli

import (
	"testing"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseID(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		expected := uuid.New()
		got, err := ParseID(expected.String())
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseID("not-a-uuid")
		assert.Error(t, err)
	})
}
