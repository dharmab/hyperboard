package main

import (
	"html/template"

	"github.com/dharmab/hyperboard/pkg/types"
)

type PostsData struct {
	Posts      []types.Post
	NextCursor string
	Search     string
	Error      string
}

type PostData struct {
	Post     types.Post
	IsVideo  bool
	FileSize int64
	Error    string
}

type TagsData struct {
	Tags           []types.Tag
	CategoryColors map[string]string
	Error          string
}

type TagEditData struct {
	Tag         types.Tag
	Categories  []types.TagCategory
	Aliases     []string
	CurrentName string
	IsNew       bool
	Error       string
}

type TagCategoriesData struct {
	Categories []types.TagCategory
	TagCounts  map[string]int
	Error      string
}

type TagCategoryEditData struct {
	Category    types.TagCategory
	CurrentName string
	IsNew       bool
	Error       string
}

type NotesData struct {
	Notes []types.Note
	Error string
}

type NoteData struct {
	Note            types.Note
	RenderedContent template.HTML
	IsNew           bool
	Error           string
}
