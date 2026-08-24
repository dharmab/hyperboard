package cli

import (
	"fmt"

	"uuid"
)

// ParseID parses a string as a UUID, returning an error with context on failure.
func ParseID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return id, fmt.Errorf("invalid ID %q: %w", s, err)
	}
	return id, nil
}
