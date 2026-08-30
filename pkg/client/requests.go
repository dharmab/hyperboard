package client

import "github.com/dharmab/hyperboard/pkg/types"

// NewPostUpdateRequest extracts mutable post metadata for an update request.
func NewPostUpdateRequest(post types.Post) PostUpdateRequest {
	return PostUpdateRequest{
		ID:   post.ID,
		Note: post.Note,
		Tags: post.Tags,
	}
}

// NewTagUpdateRequest extracts mutable tag fields for a create or update request.
func NewTagUpdateRequest(tag types.Tag) TagUpdateRequest {
	request := TagUpdateRequest{
		Name:        tag.Name,
		Category:    tag.Category,
		Description: tag.Description,
	}
	if tag.Aliases != nil {
		request.Aliases = *tag.Aliases
	}
	if tag.CascadingTags != nil {
		request.CascadingTags = *tag.CascadingTags
	}
	return request
}

// NewTagCategoryUpdateRequest extracts mutable tag-category fields for a create or update request.
func NewTagCategoryUpdateRequest(category types.TagCategory) TagCategoryUpdateRequest {
	return TagCategoryUpdateRequest{
		Name:        category.Name,
		Description: category.Description,
		Color:       category.Color,
	}
}
