// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
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

// TagAlias is an object representing the database table.
type TagAlias struct {
	ID        uuid.UUID `db:"id,pk" `
	TagID     uuid.UUID `db:"tag_id" `
	AliasName string    `db:"alias_name" `
	CreatedAt time.Time `db:"created_at" `

	R tagAliasR `db:"-" `
}

// TagAliasSlice is an alias for a slice of pointers to TagAlias.
// This should almost always be used instead of []*TagAlias.
type TagAliasSlice []*TagAlias

// TagAliases contains methods to work with the tag_aliases table
var TagAliases = psql.NewTablex[*TagAlias, TagAliasSlice, *TagAliasSetter]("", "tag_aliases", buildTagAliasColumns("tag_aliases"))

// TagAliasesQuery is a query on the tag_aliases table
type TagAliasesQuery = *psql.ViewQuery[*TagAlias, TagAliasSlice]

// tagAliasR is where relationships are stored.
type tagAliasR struct {
	Tag *Tag // tag_aliases.tag_aliases_tag_id_fkey
}

func buildTagAliasColumns(alias string) tagAliasColumns {
	return tagAliasColumns{
		ColumnsExpr: expr.NewColumnsExpr(
			"id", "tag_id", "alias_name", "created_at",
		).WithParent("tag_aliases"),
		tableAlias: alias,
		ID:         psql.Quote(alias, "id"),
		TagID:      psql.Quote(alias, "tag_id"),
		AliasName:  psql.Quote(alias, "alias_name"),
		CreatedAt:  psql.Quote(alias, "created_at"),
	}
}

type tagAliasColumns struct {
	expr.ColumnsExpr
	tableAlias string
	ID         psql.Expression
	TagID      psql.Expression
	AliasName  psql.Expression
	CreatedAt  psql.Expression
}

func (c tagAliasColumns) Alias() string {
	return c.tableAlias
}

func (tagAliasColumns) AliasedAs(alias string) tagAliasColumns {
	return buildTagAliasColumns(alias)
}

// TagAliasSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type TagAliasSetter struct {
	ID        *uuid.UUID `db:"id,pk" `
	TagID     *uuid.UUID `db:"tag_id" `
	AliasName *string    `db:"alias_name" `
	CreatedAt *time.Time `db:"created_at" `
}

func (s TagAliasSetter) SetColumns() []string {
	vals := make([]string, 0, 4)
	if s.ID != nil {
		vals = append(vals, "id")
	}
	if s.TagID != nil {
		vals = append(vals, "tag_id")
	}
	if s.AliasName != nil {
		vals = append(vals, "alias_name")
	}
	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}
	return vals
}

func (s TagAliasSetter) Overwrite(t *TagAlias) {
	if s.ID != nil {
		t.ID = func() uuid.UUID {
			if s.ID == nil {
				return *new(uuid.UUID)
			}
			return *s.ID
		}()
	}
	if s.TagID != nil {
		t.TagID = func() uuid.UUID {
			if s.TagID == nil {
				return *new(uuid.UUID)
			}
			return *s.TagID
		}()
	}
	if s.AliasName != nil {
		t.AliasName = func() string {
			if s.AliasName == nil {
				return *new(string)
			}
			return *s.AliasName
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
}

func (s *TagAliasSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return TagAliases.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 4)
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

		if s.TagID != nil {
			vals[1] = psql.Arg(func() uuid.UUID {
				if s.TagID == nil {
					return *new(uuid.UUID)
				}
				return *s.TagID
			}())
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.AliasName != nil {
			vals[2] = psql.Arg(func() string {
				if s.AliasName == nil {
					return *new(string)
				}
				return *s.AliasName
			}())
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[3] = psql.Arg(func() time.Time {
				if s.CreatedAt == nil {
					return *new(time.Time)
				}
				return *s.CreatedAt
			}())
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s TagAliasSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s TagAliasSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 4)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.TagID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "tag_id")...),
			psql.Arg(s.TagID),
		}})
	}

	if s.AliasName != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "alias_name")...),
			psql.Arg(s.AliasName),
		}})
	}

	if s.CreatedAt != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "created_at")...),
			psql.Arg(s.CreatedAt),
		}})
	}

	return exprs
}

// FindTagAlias retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindTagAlias(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*TagAlias, error) {
	if len(cols) == 0 {
		return TagAliases.Query(
			sm.Where(TagAliases.Columns.ID.EQ(psql.Arg(IDPK))),
		).One(ctx, exec)
	}

	return TagAliases.Query(
		sm.Where(TagAliases.Columns.ID.EQ(psql.Arg(IDPK))),
		sm.Columns(TagAliases.Columns.Only(cols...)),
	).One(ctx, exec)
}

// TagAliasExists checks the presence of a single record by primary key
func TagAliasExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return TagAliases.Query(
		sm.Where(TagAliases.Columns.ID.EQ(psql.Arg(IDPK))),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after TagAlias is retrieved from the database
func (o *TagAlias) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagAliases.AfterSelectHooks.RunHooks(ctx, exec, TagAliasSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = TagAliases.AfterInsertHooks.RunHooks(ctx, exec, TagAliasSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = TagAliases.AfterUpdateHooks.RunHooks(ctx, exec, TagAliasSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = TagAliases.AfterDeleteHooks.RunHooks(ctx, exec, TagAliasSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the TagAlias
func (o *TagAlias) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *TagAlias) pkEQ() dialect.Expression {
	return psql.Quote("tag_aliases", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the TagAlias
func (o *TagAlias) Update(ctx context.Context, exec bob.Executor, s *TagAliasSetter) error {
	v, err := TagAliases.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single TagAlias record with an executor
func (o *TagAlias) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := TagAliases.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the TagAlias using the executor
func (o *TagAlias) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := TagAliases.Query(
		sm.Where(TagAliases.Columns.ID.EQ(psql.Arg(o.ID))),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after TagAliasSlice is retrieved from the database
func (o TagAliasSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagAliases.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = TagAliases.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = TagAliases.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = TagAliases.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o TagAliasSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("tag_aliases", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
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
func (o TagAliasSlice) copyMatchingRows(from ...*TagAlias) {
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
func (o TagAliasSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagAliases.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagAlias:
				o.copyMatchingRows(retrieved)
			case []*TagAlias:
				o.copyMatchingRows(retrieved...)
			case TagAliasSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagAlias or a slice of TagAlias
				// then run the AfterUpdateHooks on the slice
				_, err = TagAliases.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o TagAliasSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagAliases.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagAlias:
				o.copyMatchingRows(retrieved)
			case []*TagAlias:
				o.copyMatchingRows(retrieved...)
			case TagAliasSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagAlias or a slice of TagAlias
				// then run the AfterDeleteHooks on the slice
				_, err = TagAliases.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o TagAliasSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals TagAliasSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagAliases.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o TagAliasSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagAliases.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o TagAliasSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := TagAliases.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

// Tag starts a query for related objects on tags
func (o *TagAlias) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.Where(Tags.Columns.ID.EQ(psql.Arg(o.TagID))),
	)...)
}

func (os TagAliasSlice) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	pkTagID := make(pgtypes.Array[uuid.UUID], 0, len(os))
	for _, o := range os {
		if o == nil {
			continue
		}
		pkTagID = append(pkTagID, o.TagID)
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkTagID), "uuid[]")),
	))

	return Tags.Query(append(mods,
		sm.Where(psql.Group(Tags.Columns.ID).OP("IN", PKArgExpr)),
	)...)
}

func attachTagAliasTag0(ctx context.Context, exec bob.Executor, count int, tagAlias0 *TagAlias, tag1 *Tag) (*TagAlias, error) {
	setter := &TagAliasSetter{
		TagID: func() *uuid.UUID { return &tag1.ID }(),
	}

	err := tagAlias0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachTagAliasTag0: %w", err)
	}

	return tagAlias0, nil
}

func (tagAlias0 *TagAlias) InsertTag(ctx context.Context, exec bob.Executor, related *TagSetter) error {
	var err error

	tag1, err := Tags.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachTagAliasTag0(ctx, exec, 1, tagAlias0, tag1)
	if err != nil {
		return err
	}

	tagAlias0.R.Tag = tag1

	tag1.R.TagAliases = append(tag1.R.TagAliases, tagAlias0)

	return nil
}

func (tagAlias0 *TagAlias) AttachTag(ctx context.Context, exec bob.Executor, tag1 *Tag) error {
	var err error

	_, err = attachTagAliasTag0(ctx, exec, 1, tagAlias0, tag1)
	if err != nil {
		return err
	}

	tagAlias0.R.Tag = tag1

	tag1.R.TagAliases = append(tag1.R.TagAliases, tagAlias0)

	return nil
}

type tagAliasWhere[Q psql.Filterable] struct {
	ID        psql.WhereMod[Q, uuid.UUID]
	TagID     psql.WhereMod[Q, uuid.UUID]
	AliasName psql.WhereMod[Q, string]
	CreatedAt psql.WhereMod[Q, time.Time]
}

func (tagAliasWhere[Q]) AliasedAs(alias string) tagAliasWhere[Q] {
	return buildTagAliasWhere[Q](buildTagAliasColumns(alias))
}

func buildTagAliasWhere[Q psql.Filterable](cols tagAliasColumns) tagAliasWhere[Q] {
	return tagAliasWhere[Q]{
		ID:        psql.Where[Q, uuid.UUID](cols.ID),
		TagID:     psql.Where[Q, uuid.UUID](cols.TagID),
		AliasName: psql.Where[Q, string](cols.AliasName),
		CreatedAt: psql.Where[Q, time.Time](cols.CreatedAt),
	}
}

func (o *TagAlias) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "Tag":
		rel, ok := retrieved.(*Tag)
		if !ok {
			return fmt.Errorf("tagAlias cannot load %T as %q", retrieved, name)
		}

		o.R.Tag = rel

		if rel != nil {
			rel.R.TagAliases = TagAliasSlice{o}
		}
		return nil
	default:
		return fmt.Errorf("tagAlias has no relationship %q", name)
	}
}

type tagAliasPreloader struct {
	Tag func(...psql.PreloadOption) psql.Preloader
}

func buildTagAliasPreloader() tagAliasPreloader {
	return tagAliasPreloader{
		Tag: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Tag, TagSlice](psql.PreloadRel{
				Name: "Tag",
				Sides: []psql.PreloadSide{
					{
						From:        TagAliases,
						To:          Tags,
						FromColumns: []string{"tag_id"},
						ToColumns:   []string{"id"},
					},
				},
			}, Tags.Columns.Names(), opts...)
		},
	}
}

type tagAliasThenLoader[Q orm.Loadable] struct {
	Tag func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildTagAliasThenLoader[Q orm.Loadable]() tagAliasThenLoader[Q] {
	type TagLoadInterface interface {
		LoadTag(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return tagAliasThenLoader[Q]{
		Tag: thenLoadBuilder[Q](
			"Tag",
			func(ctx context.Context, exec bob.Executor, retrieved TagLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadTag(ctx, exec, mods...)
			},
		),
	}
}

// LoadTag loads the tagAlias's Tag into the .R struct
func (o *TagAlias) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.Tag = nil

	related, err := o.Tag(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	related.R.TagAliases = TagAliasSlice{o}

	o.R.Tag = related
	return nil
}

// LoadTag loads the tagAlias's Tag into the .R struct
func (os TagAliasSlice) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	tags, err := os.Tag(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		if o == nil {
			continue
		}

		for _, rel := range tags {

			if !(o.TagID == rel.ID) {
				continue
			}

			rel.R.TagAliases = append(rel.R.TagAliases, o)

			o.R.Tag = rel
			break
		}
	}

	return nil
}

type tagAliasJoins[Q dialect.Joinable] struct {
	typ string
	Tag modAs[Q, tagColumns]
}

func (j tagAliasJoins[Q]) aliasedAs(alias string) tagAliasJoins[Q] {
	return buildTagAliasJoins[Q](buildTagAliasColumns(alias), j.typ)
}

func buildTagAliasJoins[Q dialect.Joinable](cols tagAliasColumns, typ string) tagAliasJoins[Q] {
	return tagAliasJoins[Q]{
		typ: typ,
		Tag: modAs[Q, tagColumns]{
			c: Tags.Columns,
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
