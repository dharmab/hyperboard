package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

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
