package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Post represents a row in the posts table.
type Post struct {
	ID           uuid.UUID
	MimeType     string
	ContentURL   string
	ThumbnailURL string
	Note         string
	HasAudio     bool
	SHA256       string
	Phash        sql.Null[int64]
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Loaded relationships
	Tags TagSlice
}

type PostSlice []*Post
