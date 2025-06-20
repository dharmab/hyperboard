package types

import (
	"github.com/google/uuid"
)

type ImageID uuid.UUID

type Image struct {
	ID           ImageID   `json:"id"`
	MIMEType     string    `json:"mimeType"`
	ContentURL   string    `json:"contentURL"`
	ThumbnailURL string    `json:"thumbnailURL"`
	Tags         []TagName `json:"tags"`
}
