package cli

import (
	"fmt"

	"github.com/google/uuid"
)

func ParseID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return id, fmt.Errorf("invalid ID %q: %w", s, err)
	}
	return id, nil
}
