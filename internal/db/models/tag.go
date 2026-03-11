package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
)

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
