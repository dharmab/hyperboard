package web

import (
	"html/template"

	"github.com/dharmab/hyperboard/pkg/types"
)

type tagFilter struct {
	Label string   `json:"label"`
	Tags  []string `json:"tags"`
}

type postsData struct {
	Posts      []types.Post
	NextCursor string
	Search     string
	TagFilters []tagFilter
	Error      string
}

type postData struct {
	Post         types.Post
	IsVideo      bool
	FileSize     int64
	SimilarPosts []types.Post
	Error        string
}

type tagsData struct {
	Tags           []types.Tag
	CategoryColors map[string]string
	Error          string
}

type tagEditData struct {
	Tag           types.Tag
	Categories    []types.TagCategory
	Aliases       []string
	CascadingTags []string
	CurrentName   string
	IsNew         bool
	Error         string
}

type tagCategoriesData struct {
	Categories []types.TagCategory
	TagCounts  map[string]int
	Error      string
}

type tagCategoryEditData struct {
	Category    types.TagCategory
	CurrentName string
	IsNew       bool
	Error       string
}

type notesData struct {
	Notes []types.Note
	Error string
}

type noteData struct {
	Note            types.Note
	RenderedContent template.HTML
	IsNew           bool
	Error           string
}
