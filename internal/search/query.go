package search

import "github.com/dharmab/hyperboard/pkg/types"

// Query represents parsed search parameters including tag filters, sort order, and type filters.
type Query struct {
	IncludedTags []types.TagName
	ExcludedTags []types.TagName
	Sort         Sort
	Tagged       *bool // Filter by tag presence: true = has tags, false = no tags, nil = no filter
	TypeImage    bool  // Filter to image posts (type:image)
	TypeVideo    bool  // Filter to video posts (type:video)
	TypeAudio    bool  // Filter to posts with audio (type:audio)
}
