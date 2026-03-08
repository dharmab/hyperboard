// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
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

// Tag is an object representing the database table.
type Tag struct {
	ID            uuid.UUID           `db:"id,pk" `
	Name          string              `db:"name" `
	Description   string              `db:"description" `
	TagCategoryID sql.Null[uuid.UUID] `db:"tag_category_id" `
	CreatedAt     time.Time           `db:"created_at" `
	UpdatedAt     time.Time           `db:"updated_at" `

	R tagR `db:"-" `
}

// TagSlice is an alias for a slice of pointers to Tag.
// This should almost always be used instead of []*Tag.
type TagSlice []*Tag

// Tags contains methods to work with the tags table
var Tags = psql.NewTablex[*Tag, TagSlice, *TagSetter]("", "tags")

// TagsQuery is a query on the tags table
type TagsQuery = *psql.ViewQuery[*Tag, TagSlice]

// tagR is where relationships are stored.
type tagR struct {
	Posts       PostSlice    // posts_tags.posts_tags_post_id_fkeyposts_tags.posts_tags_tag_id_fkey
	TagCategory *TagCategory // tags.tags_tag_category_id_fkey
}

type tagColumnNames struct {
	ID            string
	Name          string
	Description   string
	TagCategoryID string
	CreatedAt     string
	UpdatedAt     string
}

var TagColumns = buildTagColumns("tags")

type tagColumns struct {
	tableAlias    string
	ID            psql.Expression
	Name          psql.Expression
	Description   psql.Expression
	TagCategoryID psql.Expression
	CreatedAt     psql.Expression
	UpdatedAt     psql.Expression
}

func (c tagColumns) Alias() string {
	return c.tableAlias
}

func (tagColumns) AliasedAs(alias string) tagColumns {
	return buildTagColumns(alias)
}

func buildTagColumns(alias string) tagColumns {
	return tagColumns{
		tableAlias:    alias,
		ID:            psql.Quote(alias, "id"),
		Name:          psql.Quote(alias, "name"),
		Description:   psql.Quote(alias, "description"),
		TagCategoryID: psql.Quote(alias, "tag_category_id"),
		CreatedAt:     psql.Quote(alias, "created_at"),
		UpdatedAt:     psql.Quote(alias, "updated_at"),
	}
}

type tagWhere[Q psql.Filterable] struct {
	ID            psql.WhereMod[Q, uuid.UUID]
	Name          psql.WhereMod[Q, string]
	Description   psql.WhereMod[Q, string]
	TagCategoryID psql.WhereNullMod[Q, uuid.UUID]
	CreatedAt     psql.WhereMod[Q, time.Time]
	UpdatedAt     psql.WhereMod[Q, time.Time]
}

func (tagWhere[Q]) AliasedAs(alias string) tagWhere[Q] {
	return buildTagWhere[Q](buildTagColumns(alias))
}

func buildTagWhere[Q psql.Filterable](cols tagColumns) tagWhere[Q] {
	return tagWhere[Q]{
		ID:            psql.Where[Q, uuid.UUID](cols.ID),
		Name:          psql.Where[Q, string](cols.Name),
		Description:   psql.Where[Q, string](cols.Description),
		TagCategoryID: psql.WhereNull[Q, uuid.UUID](cols.TagCategoryID),
		CreatedAt:     psql.Where[Q, time.Time](cols.CreatedAt),
		UpdatedAt:     psql.Where[Q, time.Time](cols.UpdatedAt),
	}
}

var TagErrors = &tagErrors{
	ErrUniqueTagsPkey: &UniqueConstraintError{
		schema:  "",
		table:   "tags",
		columns: []string{"id"},
		s:       "tags_pkey",
	},

	ErrUniqueTagsNameKey: &UniqueConstraintError{
		schema:  "",
		table:   "tags",
		columns: []string{"name"},
		s:       "tags_name_key",
	},
}

type tagErrors struct {
	ErrUniqueTagsPkey *UniqueConstraintError

	ErrUniqueTagsNameKey *UniqueConstraintError
}

// TagSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type TagSetter struct {
	ID            *uuid.UUID           `db:"id,pk" `
	Name          *string              `db:"name" `
	Description   *string              `db:"description" `
	TagCategoryID *sql.Null[uuid.UUID] `db:"tag_category_id" `
	CreatedAt     *time.Time           `db:"created_at" `
	UpdatedAt     *time.Time           `db:"updated_at" `
}

func (s TagSetter) SetColumns() []string {
	vals := make([]string, 0, 6)
	if s.ID != nil {
		vals = append(vals, "id")
	}

	if s.Name != nil {
		vals = append(vals, "name")
	}

	if s.Description != nil {
		vals = append(vals, "description")
	}

	if s.TagCategoryID != nil {
		vals = append(vals, "tag_category_id")
	}

	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}

	if s.UpdatedAt != nil {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s TagSetter) Overwrite(t *Tag) {
	if s.ID != nil {
		t.ID = *s.ID
	}
	if s.Name != nil {
		t.Name = *s.Name
	}
	if s.Description != nil {
		t.Description = *s.Description
	}
	if s.TagCategoryID != nil {
		t.TagCategoryID = *s.TagCategoryID
	}
	if s.CreatedAt != nil {
		t.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt != nil {
		t.UpdatedAt = *s.UpdatedAt
	}
}

func (s *TagSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Tags.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 6)
		if s.ID != nil {
			vals[0] = psql.Arg(*s.ID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.Name != nil {
			vals[1] = psql.Arg(*s.Name)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.Description != nil {
			vals[2] = psql.Arg(*s.Description)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.TagCategoryID != nil {
			vals[3] = psql.Arg(*s.TagCategoryID)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[4] = psql.Arg(*s.CreatedAt)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.UpdatedAt != nil {
			vals[5] = psql.Arg(*s.UpdatedAt)
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s TagSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s TagSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 6)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.Name != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "name")...),
			psql.Arg(s.Name),
		}})
	}

	if s.Description != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "description")...),
			psql.Arg(s.Description),
		}})
	}

	if s.TagCategoryID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "tag_category_id")...),
			psql.Arg(s.TagCategoryID),
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

// FindTag retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindTag(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*Tag, error) {
	if len(cols) == 0 {
		return Tags.Query(
			SelectWhere.Tags.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return Tags.Query(
		SelectWhere.Tags.ID.EQ(IDPK),
		sm.Columns(Tags.Columns().Only(cols...)),
	).One(ctx, exec)
}

// TagExists checks the presence of a single record by primary key
func TagExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return Tags.Query(
		SelectWhere.Tags.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after Tag is retrieved from the database
func (o *Tag) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Tags.AfterSelectHooks.RunHooks(ctx, exec, TagSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Tags.AfterInsertHooks.RunHooks(ctx, exec, TagSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Tags.AfterUpdateHooks.RunHooks(ctx, exec, TagSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Tags.AfterDeleteHooks.RunHooks(ctx, exec, TagSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the Tag
func (o *Tag) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *Tag) pkEQ() dialect.Expression {
	return psql.Quote("tags", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the Tag
func (o *Tag) Update(ctx context.Context, exec bob.Executor, s *TagSetter) error {
	v, err := Tags.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single Tag record with an executor
func (o *Tag) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Tags.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the Tag using the executor
func (o *Tag) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Tags.Query(
		SelectWhere.Tags.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after TagSlice is retrieved from the database
func (o TagSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Tags.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Tags.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Tags.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Tags.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o TagSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("tags", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
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
func (o TagSlice) copyMatchingRows(from ...*Tag) {
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
func (o TagSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Tags.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Tag:
				o.copyMatchingRows(retrieved)
			case []*Tag:
				o.copyMatchingRows(retrieved...)
			case TagSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Tag or a slice of Tag
				// then run the AfterUpdateHooks on the slice
				_, err = Tags.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o TagSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Tags.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Tag:
				o.copyMatchingRows(retrieved)
			case []*Tag:
				o.copyMatchingRows(retrieved...)
			case TagSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Tag or a slice of Tag
				// then run the AfterDeleteHooks on the slice
				_, err = Tags.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o TagSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals TagSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Tags.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o TagSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Tags.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o TagSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Tags.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

type tagJoins[Q dialect.Joinable] struct {
	typ         string
	Posts       modAs[Q, postColumns]
	TagCategory modAs[Q, tagCategoryColumns]
}

func (j tagJoins[Q]) aliasedAs(alias string) tagJoins[Q] {
	return buildTagJoins[Q](buildTagColumns(alias), j.typ)
}

func buildTagJoins[Q dialect.Joinable](cols tagColumns, typ string) tagJoins[Q] {
	return tagJoins[Q]{
		typ: typ,
		Posts: modAs[Q, postColumns]{
			c: PostColumns,
			f: func(to postColumns) bob.Mod[Q] {
				random := strconv.FormatInt(randInt(), 10)
				mods := make(mods.QueryMods[Q], 0, 2)

				{
					to := PostsTagColumns.AliasedAs(PostsTagColumns.Alias() + random)
					mods = append(mods, dialect.Join[Q](typ, PostsTags.Name().As(to.Alias())).On(
						to.TagID.EQ(cols.ID),
					))
				}
				{
					cols := PostsTagColumns.AliasedAs(PostsTagColumns.Alias() + random)
					mods = append(mods, dialect.Join[Q](typ, Posts.Name().As(to.Alias())).On(
						to.ID.EQ(cols.PostID),
					))
				}

				return mods
			},
		},
		TagCategory: modAs[Q, tagCategoryColumns]{
			c: TagCategoryColumns,
			f: func(to tagCategoryColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, TagCategories.Name().As(to.Alias())).On(
						to.ID.EQ(cols.TagCategoryID),
					))
				}

				return mods
			},
		},
	}
}

// Posts starts a query for related objects on posts
func (o *Tag) Posts(mods ...bob.Mod[*dialect.SelectQuery]) PostsQuery {
	return Posts.Query(append(mods,
		sm.InnerJoin(PostsTags.NameAs()).On(
			PostColumns.ID.EQ(PostsTagColumns.PostID)),
		sm.Where(PostsTagColumns.TagID.EQ(psql.Arg(o.ID))),
	)...)
}

func (os TagSlice) Posts(mods ...bob.Mod[*dialect.SelectQuery]) PostsQuery {
	pkID := make(pgtypes.Array[uuid.UUID], len(os))
	for i, o := range os {
		pkID[i] = o.ID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkID), "uuid[]")),
	))

	return Posts.Query(append(mods,
		sm.InnerJoin(PostsTags.NameAs()).On(
			PostColumns.ID.EQ(PostsTagColumns.PostID),
		),
		sm.Where(psql.Group(PostsTagColumns.TagID).OP("IN", PKArgExpr)),
	)...)
}

// TagCategory starts a query for related objects on tag_categories
func (o *Tag) TagCategory(mods ...bob.Mod[*dialect.SelectQuery]) TagCategoriesQuery {
	return TagCategories.Query(append(mods,
		sm.Where(TagCategoryColumns.ID.EQ(psql.Arg(o.TagCategoryID))),
	)...)
}

func (os TagSlice) TagCategory(mods ...bob.Mod[*dialect.SelectQuery]) TagCategoriesQuery {
	pkTagCategoryID := make(pgtypes.Array[sql.Null[uuid.UUID]], len(os))
	for i, o := range os {
		pkTagCategoryID[i] = o.TagCategoryID
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkTagCategoryID), "uuid[]")),
	))

	return TagCategories.Query(append(mods,
		sm.Where(psql.Group(TagCategoryColumns.ID).OP("IN", PKArgExpr)),
	)...)
}

func (o *Tag) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Posts":
		rels, ok := retrieved.(PostSlice)
		if !ok {
			return fmt.Errorf("tag cannot load %T as %q", retrieved, name)
		}

		o.R.Posts = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.Tags = TagSlice{o}
			}
		}
		return nil
	case "TagCategory":
		rel, ok := retrieved.(*TagCategory)
		if !ok {
			return fmt.Errorf("tag cannot load %T as %q", retrieved, name)
		}

		o.R.TagCategory = rel

		if rel != nil {
			rel.R.Tags = TagSlice{o}
		}
		return nil
	default:
		return fmt.Errorf("tag has no relationship %q", name)
	}
}

type tagPreloader struct {
	TagCategory func(...psql.PreloadOption) psql.Preloader
}

func buildTagPreloader() tagPreloader {
	return tagPreloader{
		TagCategory: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*TagCategory, TagCategorySlice](orm.Relationship{
				Name: "TagCategory",
				Sides: []orm.RelSide{
					{
						From: TableNames.Tags,
						To:   TableNames.TagCategories,
						FromColumns: []string{
							ColumnNames.Tags.TagCategoryID,
						},
						ToColumns: []string{
							ColumnNames.TagCategories.ID,
						},
					},
				},
			}, TagCategories.Columns().Names(), opts...)
		},
	}
}

type tagThenLoader[Q orm.Loadable] struct {
	Posts       func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
	TagCategory func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildTagThenLoader[Q orm.Loadable]() tagThenLoader[Q] {
	type PostsLoadInterface interface {
		LoadPosts(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}
	type TagCategoryLoadInterface interface {
		LoadTagCategory(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return tagThenLoader[Q]{
		Posts: thenLoadBuilder[Q](
			"Posts",
			func(ctx context.Context, exec bob.Executor, retrieved PostsLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadPosts(ctx, exec, mods...)
			},
		),
		TagCategory: thenLoadBuilder[Q](
			"TagCategory",
			func(ctx context.Context, exec bob.Executor, retrieved TagCategoryLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTagCategory(ctx, exec, mods...)
			},
		),
	}
}

// LoadPosts loads the tag's Posts into the .R struct
func (o *Tag) LoadPosts(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Posts = nil

	related, err := o.Posts(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, rel := range related {
		rel.R.Tags = TagSlice{o}
	}

	o.R.Posts = related
	return nil
}

// LoadPosts loads the tag's Posts into the .R struct
func (os TagSlice) LoadPosts(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	// since we are changing the columns, we need to check if the original columns were set or add the defaults
	sq := dialect.SelectQuery{}
	for _, mod := range mods {
		mod.Apply(&sq)
	}

	if len(sq.SelectList.Columns) == 0 {
		mods = append(mods, sm.Columns(Posts.Columns()))
	}

	q := os.Posts(append(
		mods,
		sm.Columns(PostsTagColumns.TagID.As("related_tags.ID")),
	)...)

	IDSlice := []uuid.UUID{}

	mapper := scan.Mod(scan.StructMapper[*Post](), func(ctx context.Context, cols []string) (scan.BeforeFunc, func(any, any) error) {
		return func(row *scan.Row) (any, error) {
				IDSlice = append(IDSlice, *new(uuid.UUID))
				row.ScheduleScan("related_tags.ID", &IDSlice[len(IDSlice)-1])

				return nil, nil
			},
			func(any, any) error {
				return nil
			}
	})

	posts, err := bob.Allx[*Post, PostSlice](ctx, exec, q, mapper)
	if err != nil {
		return err
	}

	for _, o := range os {
		o.R.Posts = nil
	}

	for _, o := range os {
		for i, rel := range posts {
			if o.ID != IDSlice[i] {
				continue
			}

			rel.R.Tags = append(rel.R.Tags, o)

			o.R.Posts = append(o.R.Posts, rel)
		}
	}

	return nil
}

// LoadTagCategory loads the tag's TagCategory into the .R struct
func (o *Tag) LoadTagCategory(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.TagCategory = nil

	related, err := o.TagCategory(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	related.R.Tags = TagSlice{o}

	o.R.TagCategory = related
	return nil
}

// LoadTagCategory loads the tag's TagCategory into the .R struct
func (os TagSlice) LoadTagCategory(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	tagCategories, err := os.TagCategory(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		for _, rel := range tagCategories {
			if o.TagCategoryID.V != rel.ID {
				continue
			}

			rel.R.Tags = append(rel.R.Tags, o)

			o.R.TagCategory = rel
			break
		}
	}

	return nil
}

func attachTagPosts0(ctx context.Context, exec bob.Executor, count int, tag0 *Tag, posts2 PostSlice) (PostsTagSlice, error) {
	setters := make([]*PostsTagSetter, count)
	for i := 0; i < count; i++ {
		setters[i] = &PostsTagSetter{
			TagID:  &tag0.ID,
			PostID: &posts2[i].ID,
		}
	}

	postsTags1, err := PostsTags.Insert(bob.ToMods(setters...)).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("attachTagPosts0: %w", err)
	}

	return postsTags1, nil
}

func (tag0 *Tag) InsertPosts(ctx context.Context, exec bob.Executor, related ...*PostSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	inserted, err := Posts.Insert(bob.ToMods(related...)).All(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}
	posts2 := PostSlice(inserted)

	_, err = attachTagPosts0(ctx, exec, len(related), tag0, posts2)
	if err != nil {
		return err
	}

	tag0.R.Posts = append(tag0.R.Posts, posts2...)

	for _, rel := range posts2 {
		rel.R.Tags = append(rel.R.Tags, tag0)
	}
	return nil
}

func (tag0 *Tag) AttachPosts(ctx context.Context, exec bob.Executor, related ...*Post) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	posts2 := PostSlice(related)

	_, err = attachTagPosts0(ctx, exec, len(related), tag0, posts2)
	if err != nil {
		return err
	}

	tag0.R.Posts = append(tag0.R.Posts, posts2...)

	for _, rel := range related {
		rel.R.Tags = append(rel.R.Tags, tag0)
	}

	return nil
}

func attachTagTagCategory0(ctx context.Context, exec bob.Executor, count int, tag0 *Tag, tagCategory1 *TagCategory) (*Tag, error) {
	setter := &TagSetter{
		TagCategoryID: func() *sql.Null[uuid.UUID] {
			v := sql.Null[uuid.UUID]{V: tagCategory1.ID, Valid: true}
			return &v
		}(),
	}

	err := tag0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachTagTagCategory0: %w", err)
	}

	return tag0, nil
}

func (tag0 *Tag) InsertTagCategory(ctx context.Context, exec bob.Executor, related *TagCategorySetter) error {
	tagCategory1, err := TagCategories.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachTagTagCategory0(ctx, exec, 1, tag0, tagCategory1)
	if err != nil {
		return err
	}

	tag0.R.TagCategory = tagCategory1

	tagCategory1.R.Tags = append(tagCategory1.R.Tags, tag0)

	return nil
}

func (tag0 *Tag) AttachTagCategory(ctx context.Context, exec bob.Executor, tagCategory1 *TagCategory) error {
	var err error

	_, err = attachTagTagCategory0(ctx, exec, 1, tag0, tagCategory1)
	if err != nil {
		return err
	}

	tag0.R.TagCategory = tagCategory1

	tagCategory1.R.Tags = append(tagCategory1.R.Tags, tag0)

	return nil
}
