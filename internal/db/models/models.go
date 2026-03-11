package models

import (
	"database/sql"
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

// TagCategory represents a row in the tag_categories table.
type TagCategory struct {
	ID          uuid.UUID
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TagCategorySlice []*TagCategory

// Tag represents a row in the tags table.
type Tag struct {
	ID            uuid.UUID
	Name          string
	Description   string
	TagCategoryID sql.Null[uuid.UUID]
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Loaded relationships
	TagCategory *TagCategory
	Tags        []*Tag
}

type TagSlice []*Tag

// Post represents a row in the posts table.
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

	// Loaded relationships
	Tags TagSlice
}

type PostSlice []*Post
