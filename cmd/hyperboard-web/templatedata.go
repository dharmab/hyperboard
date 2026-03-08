package main

import "github.com/dharmab/hyperboard/pkg/types"

type GalleryData struct {
	Posts      []types.Post
	NextCursor string
	Search     string
}

type PostData struct {
	Post     types.Post
	IsVideo  bool
	FileSize int64
}

type TagsData struct {
	Tags           []types.Tag
	CategoryColors map[string]string
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
}

type TagCategoryEditData struct {
	Category    types.TagCategory
	CurrentName string
	IsNew       bool
	Error       string
}

type NotesData struct {
	Notes []types.Note
}

type NoteData struct {
	Note            types.Note
	RenderedContent string // HTML rendered from markdown
}
