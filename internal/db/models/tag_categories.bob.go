// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"io"
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
)

// TagCategory is an object representing the database table.
type TagCategory struct {
	ID          uuid.UUID `db:"id,pk" `
	Name        string    `db:"name" `
	Description string    `db:"description" `
	Color       string    `db:"color" `
	CreatedAt   time.Time `db:"created_at" `
	UpdatedAt   time.Time `db:"updated_at" `

	R tagCategoryR `db:"-" `
}

// TagCategorySlice is an alias for a slice of pointers to TagCategory.
// This should almost always be used instead of []*TagCategory.
type TagCategorySlice []*TagCategory

// TagCategories contains methods to work with the tag_categories table
var TagCategories = psql.NewTablex[*TagCategory, TagCategorySlice, *TagCategorySetter]("", "tag_categories", buildTagCategoryColumns("tag_categories"))

// TagCategoriesQuery is a query on the tag_categories table
type TagCategoriesQuery = *psql.ViewQuery[*TagCategory, TagCategorySlice]

// tagCategoryR is where relationships are stored.
type tagCategoryR struct {
	Tags TagSlice // tags.tags_tag_category_id_fkey
}

func buildTagCategoryColumns(alias string) tagCategoryColumns {
	return tagCategoryColumns{
		ColumnsExpr: expr.NewColumnsExpr(
			"id", "name", "description", "color", "created_at", "updated_at",
		).WithParent("tag_categories"),
		tableAlias:  alias,
		ID:          psql.Quote(alias, "id"),
		Name:        psql.Quote(alias, "name"),
		Description: psql.Quote(alias, "description"),
		Color:       psql.Quote(alias, "color"),
		CreatedAt:   psql.Quote(alias, "created_at"),
		UpdatedAt:   psql.Quote(alias, "updated_at"),
	}
}

type tagCategoryColumns struct {
	expr.ColumnsExpr
	tableAlias  string
	ID          psql.Expression
	Name        psql.Expression
	Description psql.Expression
	Color       psql.Expression
	CreatedAt   psql.Expression
	UpdatedAt   psql.Expression
}

func (c tagCategoryColumns) Alias() string {
	return c.tableAlias
}

func (tagCategoryColumns) AliasedAs(alias string) tagCategoryColumns {
	return buildTagCategoryColumns(alias)
}

// TagCategorySetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type TagCategorySetter struct {
	ID          *uuid.UUID `db:"id,pk" `
	Name        *string    `db:"name" `
	Description *string    `db:"description" `
	Color       *string    `db:"color" `
	CreatedAt   *time.Time `db:"created_at" `
	UpdatedAt   *time.Time `db:"updated_at" `
}

func (s TagCategorySetter) SetColumns() []string {
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
	if s.Color != nil {
		vals = append(vals, "color")
	}
	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}
	if s.UpdatedAt != nil {
		vals = append(vals, "updated_at")
	}
	return vals
}

func (s TagCategorySetter) Overwrite(t *TagCategory) {
	if s.ID != nil {
		t.ID = func() uuid.UUID {
			if s.ID == nil {
				return *new(uuid.UUID)
			}
			return *s.ID
		}()
	}
	if s.Name != nil {
		t.Name = func() string {
			if s.Name == nil {
				return *new(string)
			}
			return *s.Name
		}()
	}
	if s.Description != nil {
		t.Description = func() string {
			if s.Description == nil {
				return *new(string)
			}
			return *s.Description
		}()
	}
	if s.Color != nil {
		t.Color = func() string {
			if s.Color == nil {
				return *new(string)
			}
			return *s.Color
		}()
	}
	if s.CreatedAt != nil {
		t.CreatedAt = func() time.Time {
			if s.CreatedAt == nil {
				return *new(time.Time)
			}
			return *s.CreatedAt
		}()
	}
	if s.UpdatedAt != nil {
		t.UpdatedAt = func() time.Time {
			if s.UpdatedAt == nil {
				return *new(time.Time)
			}
			return *s.UpdatedAt
		}()
	}
}

func (s *TagCategorySetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return TagCategories.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 6)
		if s.ID != nil {
			vals[0] = psql.Arg(func() uuid.UUID {
				if s.ID == nil {
					return *new(uuid.UUID)
				}
				return *s.ID
			}())
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.Name != nil {
			vals[1] = psql.Arg(func() string {
				if s.Name == nil {
					return *new(string)
				}
				return *s.Name
			}())
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.Description != nil {
			vals[2] = psql.Arg(func() string {
				if s.Description == nil {
					return *new(string)
				}
				return *s.Description
			}())
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.Color != nil {
			vals[3] = psql.Arg(func() string {
				if s.Color == nil {
					return *new(string)
				}
				return *s.Color
			}())
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[4] = psql.Arg(func() time.Time {
				if s.CreatedAt == nil {
					return *new(time.Time)
				}
				return *s.CreatedAt
			}())
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		if s.UpdatedAt != nil {
			vals[5] = psql.Arg(func() time.Time {
				if s.UpdatedAt == nil {
					return *new(time.Time)
				}
				return *s.UpdatedAt
			}())
		} else {
			vals[5] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s TagCategorySetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s TagCategorySetter) Expressions(prefix ...string) []bob.Expression {
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

	if s.Color != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "color")...),
			psql.Arg(s.Color),
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

// FindTagCategory retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindTagCategory(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*TagCategory, error) {
	if len(cols) == 0 {
		return TagCategories.Query(
			sm.Where(TagCategories.Columns.ID.EQ(psql.Arg(IDPK))),
		).One(ctx, exec)
	}

	return TagCategories.Query(
		sm.Where(TagCategories.Columns.ID.EQ(psql.Arg(IDPK))),
		sm.Columns(TagCategories.Columns.Only(cols...)),
	).One(ctx, exec)
}

// TagCategoryExists checks the presence of a single record by primary key
func TagCategoryExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return TagCategories.Query(
		sm.Where(TagCategories.Columns.ID.EQ(psql.Arg(IDPK))),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after TagCategory is retrieved from the database
func (o *TagCategory) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagCategories.AfterSelectHooks.RunHooks(ctx, exec, TagCategorySlice{o})
	case bob.QueryTypeInsert:
		ctx, err = TagCategories.AfterInsertHooks.RunHooks(ctx, exec, TagCategorySlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = TagCategories.AfterUpdateHooks.RunHooks(ctx, exec, TagCategorySlice{o})
	case bob.QueryTypeDelete:
		ctx, err = TagCategories.AfterDeleteHooks.RunHooks(ctx, exec, TagCategorySlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the TagCategory
func (o *TagCategory) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *TagCategory) pkEQ() dialect.Expression {
	return psql.Quote("tag_categories", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the TagCategory
func (o *TagCategory) Update(ctx context.Context, exec bob.Executor, s *TagCategorySetter) error {
	v, err := TagCategories.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single TagCategory record with an executor
func (o *TagCategory) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := TagCategories.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the TagCategory using the executor
func (o *TagCategory) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := TagCategories.Query(
		sm.Where(TagCategories.Columns.ID.EQ(psql.Arg(o.ID))),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after TagCategorySlice is retrieved from the database
func (o TagCategorySlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagCategories.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = TagCategories.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = TagCategories.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = TagCategories.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o TagCategorySlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("tag_categories", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
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
func (o TagCategorySlice) copyMatchingRows(from ...*TagCategory) {
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
func (o TagCategorySlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagCategories.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagCategory:
				o.copyMatchingRows(retrieved)
			case []*TagCategory:
				o.copyMatchingRows(retrieved...)
			case TagCategorySlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagCategory or a slice of TagCategory
				// then run the AfterUpdateHooks on the slice
				_, err = TagCategories.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o TagCategorySlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagCategories.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagCategory:
				o.copyMatchingRows(retrieved)
			case []*TagCategory:
				o.copyMatchingRows(retrieved...)
			case TagCategorySlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagCategory or a slice of TagCategory
				// then run the AfterDeleteHooks on the slice
				_, err = TagCategories.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o TagCategorySlice) UpdateAll(ctx context.Context, exec bob.Executor, vals TagCategorySetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagCategories.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o TagCategorySlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagCategories.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o TagCategorySlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := TagCategories.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

// Tags starts a query for related objects on tags
func (o *TagCategory) Tags(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.Where(Tags.Columns.TagCategoryID.EQ(psql.Arg(o.ID))),
	)...)
}

func (os TagCategorySlice) Tags(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	pkID := make(pgtypes.Array[uuid.UUID], 0, len(os))
	for _, o := range os {
		if o == nil {
			continue
		}
		pkID = append(pkID, o.ID)
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkID), "uuid[]")),
	))

	return Tags.Query(append(mods,
		sm.Where(psql.Group(Tags.Columns.TagCategoryID).OP("IN", PKArgExpr)),
	)...)
}

func insertTagCategoryTags0(ctx context.Context, exec bob.Executor, tags1 []*TagSetter, tagCategory0 *TagCategory) (TagSlice, error) {
	for i := range tags1 {
		tags1[i].TagCategoryID = func() *sql.Null[uuid.UUID] { v := sql.Null[uuid.UUID]{V: tagCategory0.ID, Valid: true}; return &v }()
	}

	ret, err := Tags.Insert(bob.ToMods(tags1...)).All(ctx, exec)
	if err != nil {
		return ret, fmt.Errorf("insertTagCategoryTags0: %w", err)
	}

	return ret, nil
}

func attachTagCategoryTags0(ctx context.Context, exec bob.Executor, count int, tags1 TagSlice, tagCategory0 *TagCategory) (TagSlice, error) {
	setter := &TagSetter{
		TagCategoryID: func() *sql.Null[uuid.UUID] { v := sql.Null[uuid.UUID]{V: tagCategory0.ID, Valid: true}; return &v }(),
	}

	err := tags1.UpdateAll(ctx, exec, *setter)
	if err != nil {
		return nil, fmt.Errorf("attachTagCategoryTags0: %w", err)
	}

	return tags1, nil
}

func (tagCategory0 *TagCategory) InsertTags(ctx context.Context, exec bob.Executor, related ...*TagSetter) error {
	if len(related) == 0 {
		return nil
	}

	var err error

	tags1, err := insertTagCategoryTags0(ctx, exec, related, tagCategory0)
	if err != nil {
		return err
	}

	tagCategory0.R.Tags = append(tagCategory0.R.Tags, tags1...)

	for _, rel := range tags1 {
		rel.R.TagCategory = tagCategory0
	}
	return nil
}

func (tagCategory0 *TagCategory) AttachTags(ctx context.Context, exec bob.Executor, related ...*Tag) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	tags1 := TagSlice(related)

	_, err = attachTagCategoryTags0(ctx, exec, len(related), tags1, tagCategory0)
	if err != nil {
		return err
	}

	tagCategory0.R.Tags = append(tagCategory0.R.Tags, tags1...)

	for _, rel := range related {
		rel.R.TagCategory = tagCategory0
	}

	return nil
}

type tagCategoryWhere[Q psql.Filterable] struct {
	ID          psql.WhereMod[Q, uuid.UUID]
	Name        psql.WhereMod[Q, string]
	Description psql.WhereMod[Q, string]
	Color       psql.WhereMod[Q, string]
	CreatedAt   psql.WhereMod[Q, time.Time]
	UpdatedAt   psql.WhereMod[Q, time.Time]
}

func (tagCategoryWhere[Q]) AliasedAs(alias string) tagCategoryWhere[Q] {
	return buildTagCategoryWhere[Q](buildTagCategoryColumns(alias))
}

func buildTagCategoryWhere[Q psql.Filterable](cols tagCategoryColumns) tagCategoryWhere[Q] {
	return tagCategoryWhere[Q]{
		ID:          psql.Where[Q, uuid.UUID](cols.ID),
		Name:        psql.Where[Q, string](cols.Name),
		Description: psql.Where[Q, string](cols.Description),
		Color:       psql.Where[Q, string](cols.Color),
		CreatedAt:   psql.Where[Q, time.Time](cols.CreatedAt),
		UpdatedAt:   psql.Where[Q, time.Time](cols.UpdatedAt),
	}
}

func (o *TagCategory) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Tags":
		rels, ok := retrieved.(TagSlice)
		if !ok {
			return fmt.Errorf("tagCategory cannot load %T as %q", retrieved, name)
		}

		o.R.Tags = rels

		for _, rel := range rels {
			if rel != nil {
				rel.R.TagCategory = o
			}
		}
		return nil
	default:
		return fmt.Errorf("tagCategory has no relationship %q", name)
	}
}

type tagCategoryPreloader struct{}

func buildTagCategoryPreloader() tagCategoryPreloader {
	return tagCategoryPreloader{}
}

type tagCategoryThenLoader[Q orm.Loadable] struct {
	Tags func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildTagCategoryThenLoader[Q orm.Loadable]() tagCategoryThenLoader[Q] {
	type TagsLoadInterface interface {
		LoadTags(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return tagCategoryThenLoader[Q]{
		Tags: thenLoadBuilder[Q](
			"Tags",
			func(ctx context.Context, exec bob.Executor, retrieved TagsLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTags(ctx, exec, mods...)
			},
		),
	}
}

// LoadTags loads the tagCategory's Tags into the .R struct
func (o *TagCategory) LoadTags(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
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
		rel.R.TagCategory = o
	}

	o.R.Tags = related
	return nil
}

// LoadTags loads the tagCategory's Tags into the .R struct
func (os TagCategorySlice) LoadTags(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	tags, err := os.Tags(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		if o == nil {
			continue
		}

		o.R.Tags = nil
	}

	for _, o := range os {
		if o == nil {
			continue
		}

		for _, rel := range tags {

			if !rel.TagCategoryID.Valid {
				continue
			}
			if !(rel.TagCategoryID.Valid && o.ID == rel.TagCategoryID.V) {
				continue
			}

			rel.R.TagCategory = o

			o.R.Tags = append(o.R.Tags, rel)
		}
	}

	return nil
}

type tagCategoryJoins[Q dialect.Joinable] struct {
	typ  string
	Tags modAs[Q, tagColumns]
}

func (j tagCategoryJoins[Q]) aliasedAs(alias string) tagCategoryJoins[Q] {
	return buildTagCategoryJoins[Q](buildTagCategoryColumns(alias), j.typ)
}

func buildTagCategoryJoins[Q dialect.Joinable](cols tagCategoryColumns, typ string) tagCategoryJoins[Q] {
	return tagCategoryJoins[Q]{
		typ: typ,
		Tags: modAs[Q, tagColumns]{
			c: Tags.Columns,
			f: func(to tagColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Tags.Name().As(to.Alias())).On(
						to.TagCategoryID.EQ(cols.ID),
					))
				}

				return mods
			},
		},
	}
}
