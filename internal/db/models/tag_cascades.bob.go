// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"fmt"
	"io"

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

// TagCascade is an object representing the database table.
type TagCascade struct {
	TagID         uuid.UUID `db:"tag_id,pk" `
	CascadedTagID uuid.UUID `db:"cascaded_tag_id,pk" `

	R tagCascadeR `db:"-" `
}

// TagCascadeSlice is an alias for a slice of pointers to TagCascade.
// This should almost always be used instead of []*TagCascade.
type TagCascadeSlice []*TagCascade

// TagCascades contains methods to work with the tag_cascades table
var TagCascades = psql.NewTablex[*TagCascade, TagCascadeSlice, *TagCascadeSetter]("", "tag_cascades", buildTagCascadeColumns("tag_cascades"))

// TagCascadesQuery is a query on the tag_cascades table
type TagCascadesQuery = *psql.ViewQuery[*TagCascade, TagCascadeSlice]

// tagCascadeR is where relationships are stored.
type tagCascadeR struct {
	CascadedTagTag *Tag // tag_cascades.tag_cascades_cascaded_tag_id_fkey
	Tag            *Tag // tag_cascades.tag_cascades_tag_id_fkey
}

func buildTagCascadeColumns(alias string) tagCascadeColumns {
	return tagCascadeColumns{
		ColumnsExpr: expr.NewColumnsExpr(
			"tag_id", "cascaded_tag_id",
		).WithParent("tag_cascades"),
		tableAlias:    alias,
		TagID:         psql.Quote(alias, "tag_id"),
		CascadedTagID: psql.Quote(alias, "cascaded_tag_id"),
	}
}

type tagCascadeColumns struct {
	expr.ColumnsExpr
	tableAlias    string
	TagID         psql.Expression
	CascadedTagID psql.Expression
}

func (c tagCascadeColumns) Alias() string {
	return c.tableAlias
}

func (tagCascadeColumns) AliasedAs(alias string) tagCascadeColumns {
	return buildTagCascadeColumns(alias)
}

// TagCascadeSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type TagCascadeSetter struct {
	TagID         *uuid.UUID `db:"tag_id,pk" `
	CascadedTagID *uuid.UUID `db:"cascaded_tag_id,pk" `
}

func (s TagCascadeSetter) SetColumns() []string {
	vals := make([]string, 0, 2)
	if s.TagID != nil {
		vals = append(vals, "tag_id")
	}
	if s.CascadedTagID != nil {
		vals = append(vals, "cascaded_tag_id")
	}
	return vals
}

func (s TagCascadeSetter) Overwrite(t *TagCascade) {
	if s.TagID != nil {
		t.TagID = func() uuid.UUID {
			if s.TagID == nil {
				return *new(uuid.UUID)
			}
			return *s.TagID
		}()
	}
	if s.CascadedTagID != nil {
		t.CascadedTagID = func() uuid.UUID {
			if s.CascadedTagID == nil {
				return *new(uuid.UUID)
			}
			return *s.CascadedTagID
		}()
	}
}

func (s *TagCascadeSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return TagCascades.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 2)
		if s.TagID != nil {
			vals[0] = psql.Arg(func() uuid.UUID {
				if s.TagID == nil {
					return *new(uuid.UUID)
				}
				return *s.TagID
			}())
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.CascadedTagID != nil {
			vals[1] = psql.Arg(func() uuid.UUID {
				if s.CascadedTagID == nil {
					return *new(uuid.UUID)
				}
				return *s.CascadedTagID
			}())
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s TagCascadeSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s TagCascadeSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 2)

	if s.TagID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "tag_id")...),
			psql.Arg(s.TagID),
		}})
	}

	if s.CascadedTagID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "cascaded_tag_id")...),
			psql.Arg(s.CascadedTagID),
		}})
	}

	return exprs
}

// FindTagCascade retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindTagCascade(ctx context.Context, exec bob.Executor, TagIDPK uuid.UUID, CascadedTagIDPK uuid.UUID, cols ...string) (*TagCascade, error) {
	if len(cols) == 0 {
		return TagCascades.Query(
			sm.Where(TagCascades.Columns.TagID.EQ(psql.Arg(TagIDPK))),
			sm.Where(TagCascades.Columns.CascadedTagID.EQ(psql.Arg(CascadedTagIDPK))),
		).One(ctx, exec)
	}

	return TagCascades.Query(
		sm.Where(TagCascades.Columns.TagID.EQ(psql.Arg(TagIDPK))),
		sm.Where(TagCascades.Columns.CascadedTagID.EQ(psql.Arg(CascadedTagIDPK))),
		sm.Columns(TagCascades.Columns.Only(cols...)),
	).One(ctx, exec)
}

// TagCascadeExists checks the presence of a single record by primary key
func TagCascadeExists(ctx context.Context, exec bob.Executor, TagIDPK uuid.UUID, CascadedTagIDPK uuid.UUID) (bool, error) {
	return TagCascades.Query(
		sm.Where(TagCascades.Columns.TagID.EQ(psql.Arg(TagIDPK))),
		sm.Where(TagCascades.Columns.CascadedTagID.EQ(psql.Arg(CascadedTagIDPK))),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after TagCascade is retrieved from the database
func (o *TagCascade) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagCascades.AfterSelectHooks.RunHooks(ctx, exec, TagCascadeSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = TagCascades.AfterInsertHooks.RunHooks(ctx, exec, TagCascadeSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = TagCascades.AfterUpdateHooks.RunHooks(ctx, exec, TagCascadeSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = TagCascades.AfterDeleteHooks.RunHooks(ctx, exec, TagCascadeSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the TagCascade
func (o *TagCascade) primaryKeyVals() bob.Expression {
	return psql.ArgGroup(
		o.TagID,
		o.CascadedTagID,
	)
}

func (o *TagCascade) pkEQ() dialect.Expression {
	return psql.Group(psql.Quote("tag_cascades", "tag_id"), psql.Quote("tag_cascades", "cascaded_tag_id")).EQ(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the TagCascade
func (o *TagCascade) Update(ctx context.Context, exec bob.Executor, s *TagCascadeSetter) error {
	v, err := TagCascades.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R = v.R
	*o = *v

	return nil
}

// Delete deletes a single TagCascade record with an executor
func (o *TagCascade) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := TagCascades.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the TagCascade using the executor
func (o *TagCascade) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := TagCascades.Query(
		sm.Where(TagCascades.Columns.TagID.EQ(psql.Arg(o.TagID))),
		sm.Where(TagCascades.Columns.CascadedTagID.EQ(psql.Arg(o.CascadedTagID))),
	).One(ctx, exec)
	if err != nil {
		return err
	}
	o2.R = o.R
	*o = *o2

	return nil
}

// AfterQueryHook is called after TagCascadeSlice is retrieved from the database
func (o TagCascadeSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = TagCascades.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = TagCascades.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = TagCascades.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = TagCascades.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o TagCascadeSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Group(psql.Quote("tag_cascades", "tag_id"), psql.Quote("tag_cascades", "cascaded_tag_id")).In(bob.ExpressionFunc(func(ctx context.Context, w io.StringWriter, d bob.Dialect, start int) ([]any, error) {
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
func (o TagCascadeSlice) copyMatchingRows(from ...*TagCascade) {
	for i, old := range o {
		for _, new := range from {
			if new.TagID != old.TagID {
				continue
			}
			if new.CascadedTagID != old.CascadedTagID {
				continue
			}
			new.R = old.R
			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o TagCascadeSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagCascades.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagCascade:
				o.copyMatchingRows(retrieved)
			case []*TagCascade:
				o.copyMatchingRows(retrieved...)
			case TagCascadeSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagCascade or a slice of TagCascade
				// then run the AfterUpdateHooks on the slice
				_, err = TagCascades.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o TagCascadeSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return TagCascades.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *TagCascade:
				o.copyMatchingRows(retrieved)
			case []*TagCascade:
				o.copyMatchingRows(retrieved...)
			case TagCascadeSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a TagCascade or a slice of TagCascade
				// then run the AfterDeleteHooks on the slice
				_, err = TagCascades.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o TagCascadeSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals TagCascadeSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagCascades.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o TagCascadeSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := TagCascades.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o TagCascadeSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := TagCascades.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}

// CascadedTagTag starts a query for related objects on tags
func (o *TagCascade) CascadedTagTag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.Where(Tags.Columns.ID.EQ(psql.Arg(o.CascadedTagID))),
	)...)
}

func (os TagCascadeSlice) CascadedTagTag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	pkCascadedTagID := make(pgtypes.Array[uuid.UUID], 0, len(os))
	for _, o := range os {
		if o == nil {
			continue
		}
		pkCascadedTagID = append(pkCascadedTagID, o.CascadedTagID)
	}
	PKArgExpr := psql.Select(sm.Columns(
		psql.F("unnest", psql.Cast(psql.Arg(pkCascadedTagID), "uuid[]")),
	))

	return Tags.Query(append(mods,
		sm.Where(psql.Group(Tags.Columns.ID).OP("IN", PKArgExpr)),
	)...)
}

// Tag starts a query for related objects on tags
func (o *TagCascade) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
	return Tags.Query(append(mods,
		sm.Where(Tags.Columns.ID.EQ(psql.Arg(o.TagID))),
	)...)
}

func (os TagCascadeSlice) Tag(mods ...bob.Mod[*dialect.SelectQuery]) TagsQuery {
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

func attachTagCascadeCascadedTagTag0(ctx context.Context, exec bob.Executor, count int, tagCascade0 *TagCascade, tag1 *Tag) (*TagCascade, error) {
	setter := &TagCascadeSetter{
		CascadedTagID: func() *uuid.UUID { return &tag1.ID }(),
	}

	err := tagCascade0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachTagCascadeCascadedTagTag0: %w", err)
	}

	return tagCascade0, nil
}

func (tagCascade0 *TagCascade) InsertCascadedTagTag(ctx context.Context, exec bob.Executor, related *TagSetter) error {
	var err error

	tag1, err := Tags.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachTagCascadeCascadedTagTag0(ctx, exec, 1, tagCascade0, tag1)
	if err != nil {
		return err
	}

	tagCascade0.R.CascadedTagTag = tag1

	return nil
}

func (tagCascade0 *TagCascade) AttachCascadedTagTag(ctx context.Context, exec bob.Executor, tag1 *Tag) error {
	var err error

	_, err = attachTagCascadeCascadedTagTag0(ctx, exec, 1, tagCascade0, tag1)
	if err != nil {
		return err
	}

	tagCascade0.R.CascadedTagTag = tag1

	return nil
}

func attachTagCascadeTag0(ctx context.Context, exec bob.Executor, count int, tagCascade0 *TagCascade, tag1 *Tag) (*TagCascade, error) {
	setter := &TagCascadeSetter{
		TagID: func() *uuid.UUID { return &tag1.ID }(),
	}

	err := tagCascade0.Update(ctx, exec, setter)
	if err != nil {
		return nil, fmt.Errorf("attachTagCascadeTag0: %w", err)
	}

	return tagCascade0, nil
}

func (tagCascade0 *TagCascade) InsertTag(ctx context.Context, exec bob.Executor, related *TagSetter) error {
	var err error

	tag1, err := Tags.Insert(related).One(ctx, exec)
	if err != nil {
		return fmt.Errorf("inserting related objects: %w", err)
	}

	_, err = attachTagCascadeTag0(ctx, exec, 1, tagCascade0, tag1)
	if err != nil {
		return err
	}

	tagCascade0.R.Tag = tag1

	return nil
}

func (tagCascade0 *TagCascade) AttachTag(ctx context.Context, exec bob.Executor, tag1 *Tag) error {
	var err error

	_, err = attachTagCascadeTag0(ctx, exec, 1, tagCascade0, tag1)
	if err != nil {
		return err
	}

	tagCascade0.R.Tag = tag1

	return nil
}

type tagCascadeWhere[Q psql.Filterable] struct {
	TagID         psql.WhereMod[Q, uuid.UUID]
	CascadedTagID psql.WhereMod[Q, uuid.UUID]
}

func (tagCascadeWhere[Q]) AliasedAs(alias string) tagCascadeWhere[Q] {
	return buildTagCascadeWhere[Q](buildTagCascadeColumns(alias))
}

func buildTagCascadeWhere[Q psql.Filterable](cols tagCascadeColumns) tagCascadeWhere[Q] {
	return tagCascadeWhere[Q]{
		TagID:         psql.Where[Q, uuid.UUID](cols.TagID),
		CascadedTagID: psql.Where[Q, uuid.UUID](cols.CascadedTagID),
	}
}

func (o *TagCascade) Preload(name string, retrieved any) error {
	if o == nil {
		return nil
	}

	switch name {
	case "CascadedTagTag":
		rel, ok := retrieved.(*Tag)
		if !ok {
			return fmt.Errorf("tagCascade cannot load %T as %q", retrieved, name)
		}

		o.R.CascadedTagTag = rel

		return nil
	case "Tag":
		rel, ok := retrieved.(*Tag)
		if !ok {
			return fmt.Errorf("tagCascade cannot load %T as %q", retrieved, name)
		}

		o.R.Tag = rel

		return nil
	default:
		return fmt.Errorf("tagCascade has no relationship %q", name)
	}
}

type tagCascadePreloader struct {
	CascadedTagTag func(...psql.PreloadOption) psql.Preloader
	Tag            func(...psql.PreloadOption) psql.Preloader
}

func buildTagCascadePreloader() tagCascadePreloader {
	return tagCascadePreloader{
		CascadedTagTag: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Tag, TagSlice](psql.PreloadRel{
				Name: "CascadedTagTag",
				Sides: []psql.PreloadSide{
					{
						From:        TagCascades,
						To:          Tags,
						FromColumns: []string{"cascaded_tag_id"},
						ToColumns:   []string{"id"},
					},
				},
			}, Tags.Columns.Names(), opts...)
		},
		Tag: func(opts ...psql.PreloadOption) psql.Preloader {
			return psql.Preload[*Tag, TagSlice](psql.PreloadRel{
				Name: "Tag",
				Sides: []psql.PreloadSide{
					{
						From:        TagCascades,
						To:          Tags,
						FromColumns: []string{"tag_id"},
						ToColumns:   []string{"id"},
					},
				},
			}, Tags.Columns.Names(), opts...)
		},
	}
}

type tagCascadeThenLoader[Q orm.Loadable] struct {
	CascadedTagTag func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
	Tag            func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q]
}

func buildTagCascadeThenLoader[Q orm.Loadable]() tagCascadeThenLoader[Q] {
	type CascadedTagTagLoadInterface interface {
		LoadCascadedTagTag(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}
	type TagLoadInterface interface {
		LoadTag(context.Context, bob.Executor, ...bob.Mod[*dialect.SelectQuery]) error
	}

	return tagCascadeThenLoader[Q]{
		CascadedTagTag: thenLoadBuilder[Q](
			"CascadedTagTag",
			func(ctx context.Context, exec bob.Executor, retrieved CascadedTagTagLoadInterface, mods ...bob.Mod[*dialect.SelectQuery]) error {
				return retrieved.LoadCascadedTagTag(ctx, exec, mods...)
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

// LoadCascadedTagTag loads the tagCascade's CascadedTagTag into the .R struct
func (o *TagCascade) LoadCascadedTagTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if o == nil {
		return nil
	}

	// Reset the relationship
	o.R.CascadedTagTag = nil

	related, err := o.CascadedTagTag(mods...).One(ctx, exec)
	if err != nil {
		return err
	}

	o.R.CascadedTagTag = related
	return nil
}

// LoadCascadedTagTag loads the tagCascade's CascadedTagTag into the .R struct
func (os TagCascadeSlice) LoadCascadedTagTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
	if len(os) == 0 {
		return nil
	}

	tags, err := os.CascadedTagTag(mods...).All(ctx, exec)
	if err != nil {
		return err
	}

	for _, o := range os {
		if o == nil {
			continue
		}

		for _, rel := range tags {

			if !(o.CascadedTagID == rel.ID) {
				continue
			}

			o.R.CascadedTagTag = rel
			break
		}
	}

	return nil
}

// LoadTag loads the tagCascade's Tag into the .R struct
func (o *TagCascade) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
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

// LoadTag loads the tagCascade's Tag into the .R struct
func (os TagCascadeSlice) LoadTag(ctx context.Context, exec bob.Executor, mods ...bob.Mod[*dialect.SelectQuery]) error {
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

			o.R.Tag = rel
			break
		}
	}

	return nil
}

type tagCascadeJoins[Q dialect.Joinable] struct {
	typ            string
	CascadedTagTag modAs[Q, tagColumns]
	Tag            modAs[Q, tagColumns]
}

func (j tagCascadeJoins[Q]) aliasedAs(alias string) tagCascadeJoins[Q] {
	return buildTagCascadeJoins[Q](buildTagCascadeColumns(alias), j.typ)
}

func buildTagCascadeJoins[Q dialect.Joinable](cols tagCascadeColumns, typ string) tagCascadeJoins[Q] {
	return tagCascadeJoins[Q]{
		typ: typ,
		CascadedTagTag: modAs[Q, tagColumns]{
			c: Tags.Columns,
			f: func(to tagColumns) bob.Mod[Q] {
				mods := make(mods.QueryMods[Q], 0, 1)

				{
					mods = append(mods, dialect.Join[Q](typ, Tags.Name().As(to.Alias())).On(
						to.ID.EQ(cols.CascadedTagID),
					))
				}

				return mods
			},
		},
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
