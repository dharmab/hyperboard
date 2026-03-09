// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/maphash"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/clause"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/orm"
)

var TableNames = struct {
	Notes         string
	Posts         string
	PostsTags     string
	TagAliases    string
	TagCategories string
	Tags          string
}{
	Notes:         "notes",
	Posts:         "posts",
	PostsTags:     "posts_tags",
	TagAliases:    "tag_aliases",
	TagCategories: "tag_categories",
	Tags:          "tags",
}

var ColumnNames = struct {
	Notes         noteColumnNames
	Posts         postColumnNames
	PostsTags     postsTagColumnNames
	TagAliases    tagAliasColumnNames
	TagCategories tagCategoryColumnNames
	Tags          tagColumnNames
}{
	Notes: noteColumnNames{
		ID:        "id",
		Title:     "title",
		Content:   "content",
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",
	},
	Posts: postColumnNames{
		ID:           "id",
		MimeType:     "mime_type",
		ContentURL:   "content_url",
		ThumbnailURL: "thumbnail_url",
		Note:         "note",
		HasAudio:     "has_audio",
		Sha256:       "sha256",
		Phash:        "phash",
		CreatedAt:    "created_at",
		UpdatedAt:    "updated_at",
	},
	PostsTags: postsTagColumnNames{
		PostID: "post_id",
		TagID:  "tag_id",
	},
	TagAliases: tagAliasColumnNames{
		ID:        "id",
		TagID:     "tag_id",
		AliasName: "alias_name",
		CreatedAt: "created_at",
	},
	TagCategories: tagCategoryColumnNames{
		ID:          "id",
		Name:        "name",
		Description: "description",
		Color:       "color",
		CreatedAt:   "created_at",
		UpdatedAt:   "updated_at",
	},
	Tags: tagColumnNames{
		ID:            "id",
		Name:          "name",
		Description:   "description",
		TagCategoryID: "tag_category_id",
		CreatedAt:     "created_at",
		UpdatedAt:     "updated_at",
	},
}

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
	TagCategories tagCategoryWhere[Q]
	Tags          tagWhere[Q]
} {
	return struct {
		Notes         noteWhere[Q]
		Posts         postWhere[Q]
		PostsTags     postsTagWhere[Q]
		TagAliases    tagAliasWhere[Q]
		TagCategories tagCategoryWhere[Q]
		Tags          tagWhere[Q]
	}{
		Notes:         buildNoteWhere[Q](NoteColumns),
		Posts:         buildPostWhere[Q](PostColumns),
		PostsTags:     buildPostsTagWhere[Q](PostsTagColumns),
		TagAliases:    buildTagAliasWhere[Q](TagAliasColumns),
		TagCategories: buildTagCategoryWhere[Q](TagCategoryColumns),
		Tags:          buildTagWhere[Q](TagColumns),
	}
}

var Preload = getPreloaders()

type preloaders struct {
	Post        postPreloader
	PostsTag    postsTagPreloader
	TagAlias    tagAliasPreloader
	TagCategory tagCategoryPreloader
	Tag         tagPreloader
}

func getPreloaders() preloaders {
	return preloaders{
		Post:        buildPostPreloader(),
		PostsTag:    buildPostsTagPreloader(),
		TagAlias:    buildTagAliasPreloader(),
		TagCategory: buildTagCategoryPreloader(),
		Tag:         buildTagPreloader(),
	}
}

var (
	SelectThenLoad = getThenLoaders[*dialect.SelectQuery]()
	InsertThenLoad = getThenLoaders[*dialect.InsertQuery]()
	UpdateThenLoad = getThenLoaders[*dialect.UpdateQuery]()
)

type thenLoaders[Q orm.Loadable] struct {
	Post        postThenLoader[Q]
	PostsTag    postsTagThenLoader[Q]
	TagAlias    tagAliasThenLoader[Q]
	TagCategory tagCategoryThenLoader[Q]
	Tag         tagThenLoader[Q]
}

func getThenLoaders[Q orm.Loadable]() thenLoaders[Q] {
	return thenLoaders[Q]{
		Post:        buildPostThenLoader[Q](),
		PostsTag:    buildPostsTagThenLoader[Q](),
		TagAlias:    buildTagAliasThenLoader[Q](),
		TagCategory: buildTagCategoryThenLoader[Q](),
		Tag:         buildTagThenLoader[Q](),
	}
}

func thenLoadBuilder[Q orm.Loadable, T any](name string, f func(context.Context, bob.Executor, T, ...bob.Mod[*dialect.SelectQuery]) error) func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q] {
	return func(queryMods ...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q] {
		return orm.Loader[Q](func(ctx context.Context, exec bob.Executor, retrieved any) error {
			loader, isLoader := retrieved.(T)
			if !isLoader {
				return fmt.Errorf("object %T cannot load %q", retrieved, name)
			}

			err := f(ctx, exec, loader, queryMods...)

			// Don't cause an issue due to missing relationships
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			return err
		})
	}
}

var (
	SelectJoins = getJoins[*dialect.SelectQuery]()
	UpdateJoins = getJoins[*dialect.UpdateQuery]()
	DeleteJoins = getJoins[*dialect.DeleteQuery]()
)

type joinSet[Q interface{ aliasedAs(string) Q }] struct {
	InnerJoin Q
	LeftJoin  Q
	RightJoin Q
}

func (j joinSet[Q]) AliasedAs(alias string) joinSet[Q] {
	return joinSet[Q]{
		InnerJoin: j.InnerJoin.aliasedAs(alias),
		LeftJoin:  j.LeftJoin.aliasedAs(alias),
		RightJoin: j.RightJoin.aliasedAs(alias),
	}
}

type joins[Q dialect.Joinable] struct {
	Posts         joinSet[postJoins[Q]]
	PostsTags     joinSet[postsTagJoins[Q]]
	TagAliases    joinSet[tagAliasJoins[Q]]
	TagCategories joinSet[tagCategoryJoins[Q]]
	Tags          joinSet[tagJoins[Q]]
}

func buildJoinSet[Q interface{ aliasedAs(string) Q }, C any, F func(C, string) Q](c C, f F) joinSet[Q] {
	return joinSet[Q]{
		InnerJoin: f(c, clause.InnerJoin),
		LeftJoin:  f(c, clause.LeftJoin),
		RightJoin: f(c, clause.RightJoin),
	}
}

func getJoins[Q dialect.Joinable]() joins[Q] {
	return joins[Q]{
		Posts:         buildJoinSet[postJoins[Q]](PostColumns, buildPostJoins),
		PostsTags:     buildJoinSet[postsTagJoins[Q]](PostsTagColumns, buildPostsTagJoins),
		TagAliases:    buildJoinSet[tagAliasJoins[Q]](TagAliasColumns, buildTagAliasJoins),
		TagCategories: buildJoinSet[tagCategoryJoins[Q]](TagCategoryColumns, buildTagCategoryJoins),
		Tags:          buildJoinSet[tagJoins[Q]](TagColumns, buildTagJoins),
	}
}

type modAs[Q any, C interface{ AliasedAs(string) C }] struct {
	c C
	f func(C) bob.Mod[Q]
}

func (m modAs[Q, C]) Apply(q Q) {
	m.f(m.c).Apply(q)
}

func (m modAs[Q, C]) AliasedAs(alias string) bob.Mod[Q] {
	m.c = m.c.AliasedAs(alias)
	return m
}

func randInt() int64 {
	out := int64(new(maphash.Hash).Sum64())

	if out < 0 {
		return -out % 10000
	}

	return out % 10000
}

// ErrUniqueConstraint captures all unique constraint errors by explicitly leaving `s` empty.
var ErrUniqueConstraint = &UniqueConstraintError{s: ""}

type UniqueConstraintError struct {
	// schema is the schema where the unique constraint is defined.
	schema string
	// table is the name of the table where the unique constraint is defined.
	table string
	// columns are the columns constituting the unique constraint.
	columns []string
	// s is a string uniquely identifying the constraint in the raw error message returned from the database.
	s string
}

func (e *UniqueConstraintError) Error() string {
	return e.s
}

func (e *UniqueConstraintError) Is(target error) bool {
	err, ok := target.(*pgconn.PgError)
	if !ok {
		return false
	}
	return err.Code == "23505" && (e.s == "" || err.ConstraintName == e.s)
}
