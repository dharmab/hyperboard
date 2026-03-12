package web

import (
	"html/template"

	"github.com/dharmab/hyperboard/pkg/types"
)

// tagFilter defines a labeled set of tag names for filter buttons in the UI.
type tagFilter struct {
	Label string   `json:"label"`
	Tags  []string `json:"tags"`
}

// postsData holds template data for the posts listing page.
type postsData struct {
	Posts      []types.Post
	NextCursor string
	Search     string
	TagFilters []tagFilter
	Error      string
}

// postData holds template data for the single post view page.
type postData struct {
	Post         types.Post
	IsVideo      bool
	FileSize     int64
	SimilarPosts []types.Post
	Error        string
}

// tagsData holds template data for the tags listing page.
type tagsData struct {
	Tags           []types.Tag
	CategoryColors map[string]string
	Error          string
}

// tagEditData holds template data for the tag edit form.
type tagEditData struct {
	Tag           types.Tag
	Categories    []types.TagCategory
	Aliases       []string
	CascadingTags []string
	CurrentName   string
	IsNew         bool
	Error         string
}

// tagCategoriesData holds template data for the tag categories listing page.
type tagCategoriesData struct {
	Categories []types.TagCategory
	TagCounts  map[string]int
	Error      string
}

// tagCategoryEditData holds template data for the tag category edit form.
type tagCategoryEditData struct {
	Category    types.TagCategory
	CurrentName string
	IsNew       bool
	Error       string
}

// notesData holds template data for the notes listing page.
type notesData struct {
	Notes []types.Note
	Error string
}

// noteData holds template data for the single note view page.
type noteData struct {
	Note            types.Note
	RenderedContent template.HTML
	IsNew           bool
	Error           string
}
