// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
)

var (
	SelectWhere     = Where[*dialect.SelectQuery]()
	UpdateWhere     = Where[*dialect.UpdateQuery]()
	DeleteWhere     = Where[*dialect.DeleteQuery]()
	OnConflictWhere = Where[*clause.ConflictClause]() // Used in ON CONFLICT DO UPDATE
)

func Where[Q psql.Filterable]() struct {
	Notes         noteWhere[Q]
	Posts         postWhere[Q]
	PostsTags     postsTagWhere[Q]
	TagAliases    tagAliasWhere[Q]
	TagCascades   tagCascadeWhere[Q]
	TagCategories tagCategoryWhere[Q]
	Tags          tagWhere[Q]
} {
	return struct {
		Notes         noteWhere[Q]
		Posts         postWhere[Q]
		PostsTags     postsTagWhere[Q]
		TagAliases    tagAliasWhere[Q]
		TagCascades   tagCascadeWhere[Q]
		TagCategories tagCategoryWhere[Q]
		Tags          tagWhere[Q]
	}{
		Notes:         buildNoteWhere[Q](Notes.Columns),
		Posts:         buildPostWhere[Q](Posts.Columns),
		PostsTags:     buildPostsTagWhere[Q](PostsTags.Columns),
		TagAliases:    buildTagAliasWhere[Q](TagAliases.Columns),
		TagCascades:   buildTagCascadeWhere[Q](TagCascades.Columns),
		TagCategories: buildTagCategoryWhere[Q](TagCategories.Columns),
		Tags:          buildTagWhere[Q](Tags.Columns),
	}
}
