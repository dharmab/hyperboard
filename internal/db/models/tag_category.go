package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

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
