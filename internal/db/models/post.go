package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Post represents the posts table.
type Post struct {
	ID           uuid.UUID
	MimeType     string
	ContentURL   string
	ThumbnailURL string
	Note         string
	HasAudio     bool
	Sha256       string
	Phash        sql.Null[int64]
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Tags is populated by the store's loadPostTags method when loading posts; nil by default.
	Tags TagSlice
}

// PostSlice is a slice of Post pointers.
type PostSlice []*Post
