package cli

import (
	"testing"

	"github.com/google/uuid"
)

func TestParseID(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		expected := uuid.New()
		got, err := ParseID(expected.String())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != expected {
			t.Errorf("got %v, want %v", got, expected)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseID("not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid UUID")
		}
	})
}
