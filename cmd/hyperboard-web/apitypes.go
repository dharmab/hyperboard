package main

import "github.com/dharmab/hyperboard/internal/types"

type postsResponse struct {
	Items  *[]types.Post `json:"items"`
	Cursor *string       `json:"cursor"`
}

type tagsResponse struct {
	Items  *[]types.Tag `json:"items"`
	Cursor *string      `json:"cursor"`
}

type tagCategoriesResponse struct {
	Items  *[]types.TagCategory `json:"items"`
	Cursor *string              `json:"cursor"`
}

type notesResponse struct {
	Items  *[]types.Note `json:"items"`
	Cursor *string       `json:"cursor"`
}
