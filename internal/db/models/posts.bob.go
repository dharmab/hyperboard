// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
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
	"github.com/stephenafamo/scan"
)

// Post is an object representing the database table.
type Post struct {
	ID           uuid.UUID `db:"id,pk" `
	MimeType     string    `db:"mime_type" `
	ContentURL   string    `db:"content_url" `
	ThumbnailURL string    `db:"thumbnail_url" `
	Note         string    `db:"note" `
	CreatedAt    time.Time `db:"created_at" `
	UpdatedAt    time.Time `db:"updated_at" `

	R postR `db:"-" `
}

// PostSlice is an alias for a slice of pointers to Post.
// This should almost always be used instead of []*Post.
type PostSlice []*Post

// Posts contains methods to work with the posts table
var Posts = psql.NewTablex[*Post, PostSlice, *PostSetter]("", "posts")

// PostsQuery is a query on the posts table
type PostsQuery = *psql.ViewQuery[*Post, PostSlice]

// postR is where relationships are stored.
type postR struct {
	Tags TagSlice // posts_tags.posts_tags_post_id_fkeyposts_tags.posts_tags_tag_id_fkey
}

type postColumnNames struct {
	ID           string
	MimeType     string
	ContentURL   string
	ThumbnailURL string
	Note         string
	CreatedAt    string
	UpdatedAt    string
}

var PostColumns = buildPostColumns("posts")

type postColumns struct {
	tableAlias   string
	ID           psql.Expression
	MimeType     psql.Expression
	ContentURL   psql.Expression
	ThumbnailURL psql.Expression
	Note         psql.Expression
	CreatedAt    psql.Expression
	UpdatedAt    psql.Expression
}

func (c postColumns) Alias() string {
	return c.tableAlias
}

func (postColumns) AliasedAs(alias string) postColumns {
	return buildPostColumns(alias)
}

func buildPostColumns(alias string) postColumns {
	return postColumns{
		tableAlias:   alias,
		ID:           psql.Quote(alias, "id"),
		MimeType:     psql.Quote(alias, "mime_type"),
		ContentURL:   psql.Quote(alias, "content_url"),
		ThumbnailURL: psql.Quote(alias, "thumbnail_url"),
		Note:         psql.Quote(alias, "note"),
		CreatedAt:    psql.Quote(alias, "created_at"),
		UpdatedAt:    psql.Quote(alias, "updated_at"),
	}
}

type postWhere[Q psql.Filterable] struct {
	ID           psql.WhereMod[Q, uuid.UUID]
	MimeType     psql.WhereMod[Q, string]
	ContentURL   psql.WhereMod[Q, string]
	ThumbnailURL psql.WhereMod[Q, string]
	Note         psql.WhereMod[Q, string]
	CreatedAt    psql.WhereMod[Q, time.Time]
	UpdatedAt    psql.WhereMod[Q, time.Time]
}

func (postWhere[Q]) AliasedAs(alias string) postWhere[Q] {
	return buildPostWhere[Q](buildPostColumns(alias))
}

func buildPostWhere[Q psql.Filterable](cols postColumns) postWhere[Q] {
	return postWhere[Q]{
		ID:           psql.Where[Q, uuid.UUID](cols.ID),
		MimeType:     psql.Where[Q, string](cols.MimeType),
		ContentURL:   psql.Where[Q, string](cols.ContentURL),
		ThumbnailURL: psql.Where[Q, string](cols.ThumbnailURL),
		Note:         psql.Where[Q, string](cols.Note),
		CreatedAt:    psql.Where[Q, time.Time](cols.CreatedAt),
		UpdatedAt:    psql.Where[Q, time.Time](cols.UpdatedAt),
	}
}

var PostErrors = &postErrors{
	ErrUniquePostsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "posts",
		columns: []string{"id"},
		s:       "posts_pkey",
	},
}

type postErrors struct {
	ErrUniquePostsPkey *UniqueConstraintError
}

// PostSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type PostSetter struct {
	ID           *uuid.UUID `db:"id,pk" `
	MimeType     *string    `db:"mime_type" `
	ContentURL   *string    `db:"content_url" `
	ThumbnailURL *string    `db:"thumbnail_url" `
	Note         *string    `db:"note" `
	CreatedAt    *time.Time `db:"created_at" `
	UpdatedAt    *time.Time `db:"updated_at" `
}

func (s PostSetter) SetColumns() []string {
	vals := make([]string, 0, 7)
	if s.ID != nil {
		vals = append(vals, "id")
	}

	if s.MimeType != nil {
		vals = append(vals, "mime_type")
	}

	if s.ContentURL != nil {
		vals = append(vals, "content_url")
	}

	if s.ThumbnailURL != nil {
		vals = append(vals, "thumbnail_url")
	}

	if s.Note != nil {
		vals = append(vals, "note")
	}

	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}

	if s.UpdatedAt != nil {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s PostSetter) Overwrite(t *Post) {
	if s.ID != nil {
		t.ID = *s.ID
	}
	if s.MimeType != nil {
		t.MimeType = *s.MimeType
	}
	if s.ContentURL != nil {
		t.ContentURL = *s.ContentURL
	}
	if s.ThumbnailURL != nil {
		t.ThumbnailURL = *s.ThumbnailURL
	}
	if s.Note != nil {
		t.Note = *s.Note
	}
	if s.CreatedAt != nil {
		t.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt != nil {
		t.UpdatedAt = *s.UpdatedAt
	}
}

func (s *PostSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Posts.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 7)
		if s.ID != nil {
			vals[0] = psql.Arg(*s.ID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.MimeType != nil {
			vals[1] = psql.Arg(*s.MimeType)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.ContentURL != nil {
			vals[2] = psql.Arg(*s.ContentURL)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.ThumbnailURL != nil {
			vals[3] = psql.Arg(*s.ThumbnailURL)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.Note != nil {
			vals[4] = psql.Arg(*s.Note)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[5] = psql.Arg(*s.CreatedAt)
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		if s.UpdatedAt != nil {
			vals[6] = psql.Arg(*s.UpdatedAt)
		} else {
			vals[6] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s PostSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s PostSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 7)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.MimeType != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "mime_type")...),
			psql.Arg(s.MimeType),
		}})
	}

	if s.ContentURL != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "content_url")...),
			psql.Arg(s.ContentURL),
		}})
	}

	if s.ThumbnailURL != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "thumbnail_url")...),
			psql.Arg(s.ThumbnailURL),
		}})
	}

	if s.Note != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "note")...),
			psql.Arg(s.Note),
		}})
	}

	if s.CreatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "created_at")...),
			psql.Arg(s.CreatedAt),
		}})
	}

	if s.UpdatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "updated_at")...),
			psql.Arg(s.UpdatedAt),
		}})
	}

	return exprs
}

// FindPost retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindPost(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*Post, error) {
	if len(cols) == 0 {
		return Posts.Query(
			SelectWhere.Posts.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return Posts.Query(
		SelectWhere.Posts.ID.EQ(IDPK),
		sm.Columns(Posts.Columns().Only(cols...)),
	).One(ctx, exec)
}

// PostExists checks the presence of a single record by primary key
func PostExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return Posts.Query(
		SelectWhere.Posts.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after Post is retrieved from the database
func (o *Post) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Posts.AfterSelectHooks.RunHooks(ctx, exec, PostSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Posts.AfterInsertHooks.RunHooks(ctx, exec, PostSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Posts.AfterUpdateHooks.RunHooks(ctx, exec, PostSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Posts.AfterDeleteHooks.RunHooks(ctx, exec, PostSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the Post
func (o *Post) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *Post) pkEQ() dialect.Expression {
	return psql.Quote("posts", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the Post
func (o *Post) Update(ctx context.Context, exec bob.Executor, s *PostSetter) error {
	v, err := Posts.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single Post record with an executor
func (o *Post) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Posts.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the Post using the executor
func (o *Post) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Posts.Query(
		SelectWhere.Posts.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after PostSlice is retrieved from the database
func (o PostSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Posts.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Posts.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Posts.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Posts.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o PostSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("posts", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
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
func (o PostSlice) copyMatchingRows(from ...*Post) {
	for i, old := range o {
		for _, new := range from {
			if new.ID != old.ID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o PostSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Posts.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Post:
				o.copyMatchingRows(retrieved)
			case []*Post:
				o.copyMatchingRows(retrieved...)
			case PostSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Post or a slice of Post
				// then run the AfterUpdateHooks on the slice
				_, err = Posts.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o PostSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Posts.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Post:
				o.copyMatchingRows(retrieved)
			case []*Post:
				o.copyMatchingRows(retrieved...)
			case PostSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Post or a slice of Post
				// then run the AfterDeleteHooks on the slice
				_, err = Posts.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o PostSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals PostSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Posts.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o PostSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Posts.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o PostSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Posts.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type postJoins[Q dialect.Joinable] struct {
	typ  string
	Tags modAs[Q, tagColumns]
}

func (j postJoins[Q]) aliasedAs(alias string) postJoins[Q] {
	return buildPostJoins[Q](buildPostColumns(alias), j.typ)
}

func buildPostJoins[Q dialect.Joinable](cols postColumns, typ string) postJoins[Q] {
	return postJoins[Q]{
		typ: typ,
		Tags: modAs[Q, tagColumns]{
			c: TagColumns,
			f: func(to tagColumns) bob.Mod[Q] {
				random := strconv.FormatInt(randInt(), 10)
				mods := make(mods.QueryMods[Q], 0, 2)

				{
					to := PostsTagColumns.AliasedAs(PostsTagColumns.Alias() + random)
					mods = append(mods, dialect.Join[Q](typ, PostsTags.Name().As(to.Alias())).On(
						to.PostID.EQ(cols.ID),
					))
				}
				{
					cols := PostsTagColumns.AliasedAs(PostsTagColumns.Alias() + random)
					mods = append(mods, dialect.Join[Q](typ, Tags.Name().As(to.Alias())).On(
						to.ID.EQ(cols.TagID),
					))
				}

				return mods
			},
		},
	}
}

// Tags starts a query for related objects on tags
func (o *Post) Tags(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.InnerJoin(PostsTags.NameAs()).On(
			TagColumns.ID.EQ(PostsTagColumns.TagID)),
		sm.Where(PostsTagColumns.PostID.EQ(psql.Arg(o.ID))),
	)...)
}

func (os PostSlice) Tags(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	pkID := make(pgtypes.Array[uuid.UUID], len(os))
	for i, o := range os {
		pkID[i] = o.ID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkID), "uuid[]")),
	))

	return Tags.Query(append(mods,
		sm.InnerJoin(PostsTags.NameAs()).On(
			TagColumns.ID.EQ(PostsTagColumns.TagID),
		),
		sm.Where(psql.Group(PostsTagColumns.PostID).OP("IN", PKArgExpr)),
	)...)
}

func (o *Post) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Tags":
		rels, ok := retrieved.(TagSlice)
		if !ok {
			return fmt.Errorf("post cannot load %T as %q", retrieved, name)
		}

		o.R.Tags = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.Posts = PostSlice{o}
			}
		}
		return nil
	default:
		return fmt.Errorf("post has no relationship %q", name)
	}
}

type postPreloader struct{}

func buildPostPreloader() postPreloader {
	return postPreloader{}
}

type postThenLoader[Q orm.Loadable] struct {
	Tags func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildPostThenLoader[Q orm.Loadable]() postThenLoader[Q] {
	type TagsLoadInterface interface {
		LoadTags(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return postThenLoader[Q]{
		Tags: thenLoadBuilder[Q](
			"Tags",
			func(ctx context.Context, exec bob.Executor, retrieved TagsLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTags(ctx, exec, mods...)
			},
		),
	}
}

// LoadTags loads the post's Tags into the .R struct
func (o *Post) LoadTags(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Tags = nil

	related, err := o.Tags(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.Posts = PostSlice{o}
	}

	o.R.Tags = related
	return nil
}

// LoadTags loads the post's Tags into the .R struct
func (os PostSlice) LoadTags(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	// since we are changing the columns, we need to check if the original columns were set or add the defaults
	sq := dialect.SelectQuery{}
	for _, mod := range mods {
		mod.Apply(&sq)
	}

	if len(sq.SelectList.Columns) == 0 {
		mods = append(mods, sm.Columns(Tags.Columns()))
	}

	q := os.Tags(append(
		mods,
		sm.Columns(PostsTagColumns.PostID.As("related_posts.ID")),
	)...)

	IDSlice := []uuid.UUID{}

	mapper := scan.Mod(scan.StructMapper[*Tag](), func(ctx context.Context, cols []string) (scan.BeforeFunc, func(any, any) error) {
		return func(row *scan.Row) (any, error) {
				IDSlice = append(IDSlice, *new(uuid.UUID))
				row.ScheduleScan("related_posts.ID", &IDSlice[len(IDSlice)-1])

				return nil, nil
			},
			func(any, any) error {
				return nil
			}
	})

	tags, err := bob.Allx[*Tag, TagSlice](ctx, exec, q, mapper)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.Tags = nil
	}

	for _, o := range os {
		for i, rel := range tags {
			if o.ID != IDSlice[i] {
				continue
			}

			rel.R.Posts = append(rel.R.Posts, o)

			o.R.Tags = append(o.R.Tags, rel)
		}
	}

	return nil
}

func attachPostTags0(ctx context.Context, exec bob.Executor, count int, post0 *Post, tags2 TagSlice) (PostsTagSlice, error) {
	setters := make([]*PostsTagSetter, count)
	for i := 0; i < count; i++ {
		setters[i] = &PostsTagSetter{
			PostID: &post0.ID,
			TagID:  &tags2[i].ID,
		}
	}

	postsTags1, err := PostsTags.Insert(bob.ToMods(setters...)).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("attachPostTags0: %w", err)
	}

	return postsTags1, nil
}

func (post0 *Post) InsertTags(ctx context.Context, exec bob.Executor, related ...*TagSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	inserted, err := Tags.Insert(bob.ToMods(related...)).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}
	tags2 := TagSlice(inserted)

	_, err = attachPostTags0(ctx, exec, len(related), post0, tags2)
	if err != nil {
		return err
	}

	post0.R.Tags = append(post0.R.Tags, tags2...)

	for _, rel := range tags2 {
		rel.R.Posts = append(rel.R.Posts, post0)
	}
	return nil
}

func (post0 *Post) AttachTags(ctx context.Context, exec bob.Executor, related ...*Tag) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	tags2 := TagSlice(related)

	_, err = attachPostTags0(ctx, exec, len(related), post0, tags2)
	if err != nil {
		return err
	}

	post0.R.Tags = append(post0.R.Tags, tags2...)

	for _, rel := range related {
		rel.R.Posts = append(rel.R.Posts, post0)
	}

	return nil
}
