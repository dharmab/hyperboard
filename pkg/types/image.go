package types

import (
	"time"

	"github.com/google/uuid"
)

type ImageID uuid.UUID

type Image struct {
	ID            ImageID   `json:"id"`
	Version       uint      `json:"version"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	MIMEType      string    `json:"mimeType"`
	ContentURL    string    `json:"contentURL"`
	ThumbnailURL  string    `json:"thumbnailURL"`
	Tags          []TagName `json:"tags"`
	RelatedImages []ImageID `json:"relatedImages"`
	IsFavorite    bool      `json:"isFavorite"`
}

type TagName string

type Tag struct {
	Name     TagName      `json:"name"`
	Category CategoryName `json:"category"`
}

type CategoryName string

type TagCategory struct {
	Name CategoryName `json:"name"`
}
