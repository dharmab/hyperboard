package models

import (
	"database/sql"
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

// TagCategory represents the tag_categories table.
type TagCategory struct {
	ID          uuid.UUID
	Name        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TagCategorySlice is a slice of TagCategory pointers.
type TagCategorySlice []*TagCategory

// Tag represents the tags table.
type Tag struct {
	ID            uuid.UUID
	Name          string
	Description   string
	TagCategoryID sql.Null[uuid.UUID]
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Category is populated by the store's loadTagCategories method when loading tags; nil by default.
	Category *TagCategory
}

// TagSlice is a slice of Tag pointers.
type TagSlice []*Tag

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
