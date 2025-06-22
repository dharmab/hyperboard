// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"fmt"
	"io"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/dialect/psql/um"
	"github.com/stephenafamo/bob/expr"
	"github.com/stephenafamo/bob/mods"
	"github.com/stephenafamo/bob/orm"
	"github.com/stephenafamo/bob/types/pgtypes"
)

// PostsTag is an object representing the database table.
type PostsTag struct {
	PostID int32 `db:"post_id,pk" `
	TagID  int32 `db:"tag_id,pk" `

	R postsTagR `db:"-" `
}

// PostsTagSlice is an alias for a slice of pointers to PostsTag.
// This should almost always be used instead of []*PostsTag.
type PostsTagSlice []*PostsTag

// PostsTags contains methods to work with the posts_tags table
var PostsTags = psql.NewTablex[*PostsTag, PostsTagSlice, *PostsTagSetter]("", "posts_tags")

// PostsTagsQuery is a query on the posts_tags table
type PostsTagsQuery = *psql.ViewQuery[*PostsTag, PostsTagSlice]

// postsTagR is where relationships are stored.
type postsTagR struct {
	Post *Post // posts_tags.posts_tags_post_id_fkey
	Tag  *Tag  // posts_tags.posts_tags_tag_id_fkey
}

type postsTagColumnNames struct {
	PostID string
	TagID  string
}

var PostsTagColumns = buildPostsTagColumns("posts_tags")

type postsTagColumns struct {
	tableAlias string
	PostID     psql.Expression
	TagID      psql.Expression
}

func (c postsTagColumns) Alias() string {
	return c.tableAlias
}

func (postsTagColumns) AliasedAs(alias string) postsTagColumns {
	return buildPostsTagColumns(alias)
}

func buildPostsTagColumns(alias string) postsTagColumns {
	return postsTagColumns{
		tableAlias: alias,
		PostID:     psql.Quote(alias, "post_id"),
		TagID:      psql.Quote(alias, "tag_id"),
	}
}

type postsTagWhere[Q psql.Filterable] struct {
	PostID psql.WhereMod[Q, int32]
	TagID  psql.WhereMod[Q, int32]
}

func (postsTagWhere[Q]) AliasedAs(alias string) postsTagWhere[Q] {
	return buildPostsTagWhere[Q](buildPostsTagColumns(alias))
}

func buildPostsTagWhere[Q psql.Filterable](cols postsTagColumns) postsTagWhere[Q] {
	return postsTagWhere[Q]{
		PostID: psql.Where[Q, int32](cols.PostID),
		TagID:  psql.Where[Q, int32](cols.TagID),
	}
}

var PostsTagErrors = &postsTagErrors{
	ErrUniquePostsTagsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "posts_tags",
		columns: []string{"post_id", "tag_id"},
		s:       "posts_tags_pkey",
	},
}

type postsTagErrors struct {
	ErrUniquePostsTagsPkey *UniqueConstraintError
}

// PostsTagSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type PostsTagSetter struct {
	PostID *int32 `db:"post_id,pk" `
	TagID  *int32 `db:"tag_id,pk" `
}

func (s PostsTagSetter) SetColumns() []string {
	vals := make([]string, 0, 2)
	if s.PostID != nil {
		vals = append(vals, "post_id")
	}

	if s.TagID != nil {
		vals = append(vals, "tag_id")
	}

	return vals
}

func (s PostsTagSetter) Overwrite(t *PostsTag) {
	if s.PostID != nil {
		t.PostID = *s.PostID
	}
	if s.TagID != nil {
		t.TagID = *s.TagID
	}
}

func (s *PostsTagSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return PostsTags.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 2)
		if s.PostID != nil {
			vals[0] = psql.Arg(*s.PostID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.TagID != nil {
			vals[1] = psql.Arg(*s.TagID)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s PostsTagSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s PostsTagSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 2)

	if s.PostID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "post_id")...),
			psql.Arg(s.PostID),
		}})
	}

	if s.TagID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "tag_id")...),
			psql.Arg(s.TagID),
		}})
	}

	return exprs
}

// FindPostsTag retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindPostsTag(ctx context.Context, exec bob.Executor, PostIDPK int32, TagIDPK int32, cols ...string) (*PostsTag, error) {
	if len(cols) == 0 {
		return PostsTags.Query(
			SelectWhere.PostsTags.PostID.EQ(PostIDPK),
			SelectWhere.PostsTags.TagID.EQ(TagIDPK),
		).One(ctx, exec)
	}

	return PostsTags.Query(
		SelectWhere.PostsTags.PostID.EQ(PostIDPK),
		SelectWhere.PostsTags.TagID.EQ(TagIDPK),
		sm.Columns(PostsTags.Columns().Only(cols...)),
	).One(ctx, exec)
}

// PostsTagExists checks the presence of a single record by primary key
func PostsTagExists(ctx context.Context, exec bob.Executor, PostIDPK int32, TagIDPK int32) (bool, error) {
	return PostsTags.Query(
		SelectWhere.PostsTags.PostID.EQ(PostIDPK),
		SelectWhere.PostsTags.TagID.EQ(TagIDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after PostsTag is retrieved from the database
func (o *PostsTag) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = PostsTags.AfterSelectHooks.RunHooks(ctx, exec, PostsTagSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = PostsTags.AfterInsertHooks.RunHooks(ctx, exec, PostsTagSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = PostsTags.AfterUpdateHooks.RunHooks(ctx, exec, PostsTagSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = PostsTags.AfterDeleteHooks.RunHooks(ctx, exec, PostsTagSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the PostsTag
func (o *PostsTag) primaryKeyVals() bob.Expression {
	return psql.ArgGroup(
		o.PostID,
		o.TagID,
	)
}

func (o *PostsTag) pkEQ() dialect.Expression {
	return psql.Group(psql.Quote("posts_tags", "post_id"), psql.Quote("posts_tags", "tag_id")).EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the PostsTag
func (o *PostsTag) Update(ctx context.Context, exec bob.Executor, s *PostsTagSetter) error {
	v, err := PostsTags.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single PostsTag record with an executor
func (o *PostsTag) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := PostsTags.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the PostsTag using the executor
func (o *PostsTag) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := PostsTags.Query(
		SelectWhere.PostsTags.PostID.EQ(o.PostID),
		SelectWhere.PostsTags.TagID.EQ(o.TagID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after PostsTagSlice is retrieved from the database
func (o PostsTagSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = PostsTags.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = PostsTags.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = PostsTags.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = PostsTags.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o PostsTagSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Group(psql.Quote("posts_tags", "post_id"), psql.Quote("posts_tags", "tag_id")).In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		pkPairs := make([]bob.Expression, len(o))
		for i, row := range o {
			pkPairs[i] = row.primaryKeyVals()
		}
		return bob.ExpressSlice(ctx, w, d, start, pkPairs, "", ", ", "")
	}))
}

// copyMatchingRows finds models in the given slice that have the same primary key
// then it first copies the existing relationships from the old model to the new model
// and then replaces the old model in the slice with the new model
func (o PostsTagSlice) copyMatchingRows(from ...*PostsTag) {
	for i, old := range o {
		for _, new := range from {
			if new.PostID != old.PostID {
				continue
			}
			if new.TagID != old.TagID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o PostsTagSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return PostsTags.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *PostsTag:
				o.copyMatchingRows(retrieved)
			case []*PostsTag:
				o.copyMatchingRows(retrieved...)
			case PostsTagSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a PostsTag or a slice of PostsTag
				// then run the AfterUpdateHooks on the slice
				_, err = PostsTags.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o PostsTagSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return PostsTags.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *PostsTag:
				o.copyMatchingRows(retrieved)
			case []*PostsTag:
				o.copyMatchingRows(retrieved...)
			case PostsTagSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a PostsTag or a slice of PostsTag
				// then run the AfterDeleteHooks on the slice
				_, err = PostsTags.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o PostsTagSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals PostsTagSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := PostsTags.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o PostsTagSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := PostsTags.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o PostsTagSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := PostsTags.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type postsTagJoins[Q dialect.Joinable] struct {
	typ  string
	Post modAs[Q, postColumns]
	Tag  modAs[Q, tagColumns]
}

func (j postsTagJoins[Q]) aliasedAs(alias string) postsTagJoins[Q] {
	return buildPostsTagJoins[Q](buildPostsTagColumns(alias), j.typ)
}

func buildPostsTagJoins[Q dialect.Joinable](cols postsTagColumns, typ string) postsTagJoins[Q] {
	return postsTagJoins[Q]{
		typ: typ,
		Post: modAs[Q, postColumns]{
			c: PostColumns,
			f: func(to postColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Posts.Name().As(to.Alias())).On(
						to.ID.EQ(cols.PostID),
					))
				}

				return mods
			},
		},
		Tag: modAs[Q, tagColumns]{
			c: TagColumns,
			f: func(to tagColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Tags.Name().As(to.Alias())).On(
						to.ID.EQ(cols.TagID),
					))
				}

				return mods
			},
		},
	}
}

// Post starts a query for related objects on posts
func (o *PostsTag) Post(mods ...bob.Mod[*dialect.SelectQuery]) PostsQuery {
	return Posts.Query(append(mods,
		sm.Where(PostColumns.ID.EQ(psql.Arg(o.PostID))),
	)...)
}

func (os PostsTagSlice) Post(mods ...bob.Mod[*dialect.SelectQuery]) PostsQuery {
	pkPostID := make(pgtypes.Array[int32], len(os))
	for i, o := range os {
		pkPostID[i] = o.PostID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkPostID), "integer[]")),
	))

	return Posts.Query(append(mods,
		sm.Where(psql.Group(PostColumns.ID).OP("IN", PKArgExpr)),
	)...)
}

// Tag starts a query for related objects on tags
func (o *PostsTag) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.Where(TagColumns.ID.EQ(psql.Arg(o.TagID))),
	)...)
}

func (os PostsTagSlice) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	pkTagID := make(pgtypes.Array[int32], len(os))
	for i, o := range os {
		pkTagID[i] = o.TagID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkTagID), "integer[]")),
	))

	return Tags.Query(append(mods,
		sm.Where(psql.Group(TagColumns.ID).OP("IN", PKArgExpr)),
	)...)
}

func (o *PostsTag) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Post":
		rel, ok := retrieved.(*Post)
		if !ok {
			return fmt.Errorf("postsTag cannot load %T as %q", retrieved, name)
		}

		o.R.Post = rel

		return nil
	case "Tag":
		rel, ok := retrieved.(*Tag)
		if !ok {
			return fmt.Errorf("postsTag cannot load %T as %q", retrieved, name)
		}

		o.R.Tag = rel

		return nil
	default:
		return fmt.Errorf("postsTag has no relationship %q", name)
	}
}

type postsTagPreloader struct {
	Post func(...psql.PreloadOption) psql.Preloader
	Tag  func(...psql.PreloadOption) psql.Preloader
}

func buildPostsTagPreloader() postsTagPreloader {
	return postsTagPreloader{
		Post: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Post, PostSlice](orm.Relationship{
				Name: "Post",
				Sides: []orm.RelSide{
					{
						From: TableNames.PostsTags,
						To:   TableNames.Posts,
						FromColumns: []string{
							ColumnNames.PostsTags.PostID,
						},
						ToColumns: []string{
							ColumnNames.Posts.ID,
						},
					},
				},
			}, Posts.Columns().Names(), opts...)
		},
		Tag: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Tag, TagSlice](orm.Relationship{
				Name: "Tag",
				Sides: []orm.RelSide{
					{
						From: TableNames.PostsTags,
						To:   TableNames.Tags,
						FromColumns: []string{
							ColumnNames.PostsTags.TagID,
						},
						ToColumns: []string{
							ColumnNames.Tags.ID,
						},
					},
				},
			}, Tags.Columns().Names(), opts...)
		},
	}
}

type postsTagThenLoader[Q orm.Loadable] struct {
	Post func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
	Tag  func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildPostsTagThenLoader[Q orm.Loadable]() postsTagThenLoader[Q] {
	type PostLoadInterface interface {
		LoadPost(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}
	type TagLoadInterface interface {
		LoadTag(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return postsTagThenLoader[Q]{
		Post: thenLoadBuilder[Q](
			"Post",
			func(ctx context.Context, exec bob.Executor, retrieved PostLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadPost(ctx, exec, mods...)
			},
		),
		Tag: thenLoadBuilder[Q](
			"Tag",
			func(ctx context.Context, exec bob.Executor, retrieved TagLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTag(ctx, exec, mods...)
			},
		),
	}
}

// LoadPost loads the postsTag's Post into the .R struct
func (o *PostsTag) LoadPost(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Post = nil

	related, err := o.Post(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R.Post = related
	return nil
}

// LoadPost loads the postsTag's Post into the .R struct
func (os PostsTagSlice) LoadPost(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	posts, err := os.Post(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range posts {
			if o.PostID != rel.ID {
				continue
			}

			o.R.Post = rel
			break
		}
	}

	return nil
}

// LoadTag loads the postsTag's Tag into the .R struct
func (o *PostsTag) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Tag = nil

	related, err := o.Tag(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R.Tag = related
	return nil
}

// LoadTag loads the postsTag's Tag into the .R struct
func (os PostsTagSlice) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	tags, err := os.Tag(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range tags {
			if o.TagID != rel.ID {
				continue
			}

			o.R.Tag = rel
			break
		}
	}

	return nil
}

func attachPostsTagPost0(ctx context.Context, exec bob.Executor, count int, postsTag0 *PostsTag, post1 *Post) (*PostsTag, error) {
	setter := &PostsTagSetter{
		PostID: &post1.ID,
	}

	err := postsTag0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachPostsTagPost0: %w", err)
	}

	return postsTag0, nil
}

func (postsTag0 *PostsTag) InsertPost(ctx context.Context, exec bob.Executor, related *PostSetter) error {
	post1, err := Posts.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachPostsTagPost0(ctx, exec, 1, postsTag0, post1)
	if err != nil {
		return err
	}

	postsTag0.R.Post = post1

	return nil
}

func (postsTag0 *PostsTag) AttachPost(ctx context.Context, exec bob.Executor, post1 *Post) error {
	var err error

	_, err = attachPostsTagPost0(ctx, exec, 1, postsTag0, post1)
	if err != nil {
		return err
	}

	postsTag0.R.Post = post1

	return nil
}

func attachPostsTagTag0(ctx context.Context, exec bob.Executor, count int, postsTag0 *PostsTag, tag1 *Tag) (*PostsTag, error) {
	setter := &PostsTagSetter{
		TagID: &tag1.ID,
	}

	err := postsTag0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachPostsTagTag0: %w", err)
	}

	return postsTag0, nil
}

func (postsTag0 *PostsTag) InsertTag(ctx context.Context, exec bob.Executor, related *TagSetter) error {
	tag1, err := Tags.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachPostsTagTag0(ctx, exec, 1, postsTag0, tag1)
	if err != nil {
		return err
	}

	postsTag0.R.Tag = tag1

	return nil
}

func (postsTag0 *PostsTag) AttachTag(ctx context.Context, exec bob.Executor, tag1 *Tag) error {
	var err error

	_, err = attachPostsTagTag0(ctx, exec, 1, postsTag0, tag1)
	if err != nil {
		return err
	}

	postsTag0.R.Tag = tag1

	return nil
}
