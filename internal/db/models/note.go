package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// Note represents the notes table.
type Note struct {
	ID        uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NoteSlice is a slice of Note pointers.
type NoteSlice []*Note
