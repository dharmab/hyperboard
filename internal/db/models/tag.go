package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
)

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
