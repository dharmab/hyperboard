package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Note represents a row in the notes table.
type Note struct {
	ID        uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteSlice []*Note
