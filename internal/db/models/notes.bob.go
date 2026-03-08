// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
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
)

// Note is an object representing the database table.
type Note struct {
	ID        uuid.UUID `db:"id,pk" `
	Title     string    `db:"title" `
	Content   string    `db:"content" `
	CreatedAt time.Time `db:"created_at" `
	UpdatedAt time.Time `db:"updated_at" `
}

// NoteSlice is an alias for a slice of pointers to Note.
// This should almost always be used instead of []*Note.
type NoteSlice []*Note

// Notes contains methods to work with the notes table
var Notes = psql.NewTablex[*Note, NoteSlice, *NoteSetter]("", "notes")

// NotesQuery is a query on the notes table
type NotesQuery = *psql.ViewQuery[*Note, NoteSlice]

type noteColumnNames struct {
	ID        string
	Title     string
	Content   string
	CreatedAt string
	UpdatedAt string
}

var NoteColumns = buildNoteColumns("notes")

type noteColumns struct {
	tableAlias string
	ID         psql.Expression
	Title      psql.Expression
	Content    psql.Expression
	CreatedAt  psql.Expression
	UpdatedAt  psql.Expression
}

func (c noteColumns) Alias() string {
	return c.tableAlias
}

func (noteColumns) AliasedAs(alias string) noteColumns {
	return buildNoteColumns(alias)
}

func buildNoteColumns(alias string) noteColumns {
	return noteColumns{
		tableAlias: alias,
		ID:         psql.Quote(alias, "id"),
		Title:      psql.Quote(alias, "title"),
		Content:    psql.Quote(alias, "content"),
		CreatedAt:  psql.Quote(alias, "created_at"),
		UpdatedAt:  psql.Quote(alias, "updated_at"),
	}
}

type noteWhere[Q psql.Filterable] struct {
	ID        psql.WhereMod[Q, uuid.UUID]
	Title     psql.WhereMod[Q, string]
	Content   psql.WhereMod[Q, string]
	CreatedAt psql.WhereMod[Q, time.Time]
	UpdatedAt psql.WhereMod[Q, time.Time]
}

func (noteWhere[Q]) AliasedAs(alias string) noteWhere[Q] {
	return buildNoteWhere[Q](buildNoteColumns(alias))
}

func buildNoteWhere[Q psql.Filterable](cols noteColumns) noteWhere[Q] {
	return noteWhere[Q]{
		ID:        psql.Where[Q, uuid.UUID](cols.ID),
		Title:     psql.Where[Q, string](cols.Title),
		Content:   psql.Where[Q, string](cols.Content),
		CreatedAt: psql.Where[Q, time.Time](cols.CreatedAt),
		UpdatedAt: psql.Where[Q, time.Time](cols.UpdatedAt),
	}
}

var NoteErrors = &noteErrors{
	ErrUniqueNotesPkey: &UniqueConstraintError{
		schema:  "",
		table:   "notes",
		columns: []string{"id"},
		s:       "notes_pkey",
	},
}

type noteErrors struct {
	ErrUniqueNotesPkey *UniqueConstraintError
}

// NoteSetter is used for insert/upsert/update operations
// All values are optional, and do not have to be set
// Generated columns are not included
type NoteSetter struct {
	ID        *uuid.UUID `db:"id,pk" `
	Title     *string    `db:"title" `
	Content   *string    `db:"content" `
	CreatedAt *time.Time `db:"created_at" `
	UpdatedAt *time.Time `db:"updated_at" `
}

func (s NoteSetter) SetColumns() []string {
	vals := make([]string, 0, 5)
	if s.ID != nil {
		vals = append(vals, "id")
	}

	if s.Title != nil {
		vals = append(vals, "title")
	}

	if s.Content != nil {
		vals = append(vals, "content")
	}

	if s.CreatedAt != nil {
		vals = append(vals, "created_at")
	}

	if s.UpdatedAt != nil {
		vals = append(vals, "updated_at")
	}

	return vals
}

func (s NoteSetter) Overwrite(t *Note) {
	if s.ID != nil {
		t.ID = *s.ID
	}
	if s.Title != nil {
		t.Title = *s.Title
	}
	if s.Content != nil {
		t.Content = *s.Content
	}
	if s.CreatedAt != nil {
		t.CreatedAt = *s.CreatedAt
	}
	if s.UpdatedAt != nil {
		t.UpdatedAt = *s.UpdatedAt
	}
}

func (s *NoteSetter) Apply(q *dialect.InsertQuery) {
	q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
		return Notes.BeforeInsertHooks.RunHooks(ctx, exec, s)
	})

	q.AppendValues(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		vals := make([]bob.Expression, 5)
		if s.ID != nil {
			vals[0] = psql.Arg(*s.ID)
		} else {
			vals[0] = psql.Raw("DEFAULT")
		}

		if s.Title != nil {
			vals[1] = psql.Arg(*s.Title)
		} else {
			vals[1] = psql.Raw("DEFAULT")
		}

		if s.Content != nil {
			vals[2] = psql.Arg(*s.Content)
		} else {
			vals[2] = psql.Raw("DEFAULT")
		}

		if s.CreatedAt != nil {
			vals[3] = psql.Arg(*s.CreatedAt)
		} else {
			vals[3] = psql.Raw("DEFAULT")
		}

		if s.UpdatedAt != nil {
			vals[4] = psql.Arg(*s.UpdatedAt)
		} else {
			vals[4] = psql.Raw("DEFAULT")
		}

		return bob.ExpressSlice(ctx, w, d, start, vals, "", ", ", "")
	}))
}

func (s NoteSetter) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return um.Set(s.Expressions()...)
}

func (s NoteSetter) Expressions(prefix ...string) []bob.Expression {
	exprs := make([]bob.Expression, 0, 5)

	if s.ID != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "id")...),
			psql.Arg(s.ID),
		}})
	}

	if s.Title != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "title")...),
			psql.Arg(s.Title),
		}})
	}

	if s.Content != nil {
		exprs = append(exprs, expr.Join{Sep: " = ", Exprs: []bob.Expression{
			psql.Quote(append(prefix, "content")...),
			psql.Arg(s.Content),
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

// FindNote retrieves a single record by primary key
// If cols is empty Find will return all columns.
func FindNote(ctx context.Context, exec bob.Executor, IDPK uuid.UUID, cols ...string) (*Note, error) {
	if len(cols) == 0 {
		return Notes.Query(
			SelectWhere.Notes.ID.EQ(IDPK),
		).One(ctx, exec)
	}

	return Notes.Query(
		SelectWhere.Notes.ID.EQ(IDPK),
		sm.Columns(Notes.Columns().Only(cols...)),
	).One(ctx, exec)
}

// NoteExists checks the presence of a single record by primary key
func NoteExists(ctx context.Context, exec bob.Executor, IDPK uuid.UUID) (bool, error) {
	return Notes.Query(
		SelectWhere.Notes.ID.EQ(IDPK),
	).Exists(ctx, exec)
}

// AfterQueryHook is called after Note is retrieved from the database
func (o *Note) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Notes.AfterSelectHooks.RunHooks(ctx, exec, NoteSlice{o})
	case bob.QueryTypeInsert:
		ctx, err = Notes.AfterInsertHooks.RunHooks(ctx, exec, NoteSlice{o})
	case bob.QueryTypeUpdate:
		ctx, err = Notes.AfterUpdateHooks.RunHooks(ctx, exec, NoteSlice{o})
	case bob.QueryTypeDelete:
		ctx, err = Notes.AfterDeleteHooks.RunHooks(ctx, exec, NoteSlice{o})
	}

	return err
}

// primaryKeyVals returns the primary key values of the Note
func (o *Note) primaryKeyVals() bob.Expression {
	return psql.Arg(o.ID)
}

func (o *Note) pkEQ() dialect.Expression {
	return psql.Quote("notes", "id").EQ(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
		return o.primaryKeyVals().WriteSQL(ctx, w, d, start)
	}))
}

// Update uses an executor to update the Note
func (o *Note) Update(ctx context.Context, exec bob.Executor, s *NoteSetter) error {
	v, err := Notes.Update(s.UpdateMod(), um.Where(o.pkEQ())).One(ctx, exec)
	if err != nil {
		return err
	}

	*o = *v

	return nil
}

// Delete deletes a single Note record with an executor
func (o *Note) Delete(ctx context.Context, exec bob.Executor) error {
	_, err := Notes.Delete(dm.Where(o.pkEQ())).Exec(ctx, exec)
	return err
}

// Reload refreshes the Note using the executor
func (o *Note) Reload(ctx context.Context, exec bob.Executor) error {
	o2, err := Notes.Query(
		SelectWhere.Notes.ID.EQ(o.ID),
	).One(ctx, exec)
	if err != nil {
		return err
	}

	*o = *o2

	return nil
}

// AfterQueryHook is called after NoteSlice is retrieved from the database
func (o NoteSlice) AfterQueryHook(ctx context.Context, exec bob.Executor, queryType bob.QueryType) error {
	var err error

	switch queryType {
	case bob.QueryTypeSelect:
		ctx, err = Notes.AfterSelectHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeInsert:
		ctx, err = Notes.AfterInsertHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeUpdate:
		ctx, err = Notes.AfterUpdateHooks.RunHooks(ctx, exec, o)
	case bob.QueryTypeDelete:
		ctx, err = Notes.AfterDeleteHooks.RunHooks(ctx, exec, o)
	}

	return err
}

func (o NoteSlice) pkIN() dialect.Expression {
	if len(o) == 0 {
		return psql.Raw("NULL")
	}

	return psql.Quote("notes", "id").In(bob.ExpressionFunc(func(ctx context.Context, w io.Writer, d bob.Dialect, start int) ([]any, error) {
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
func (o NoteSlice) copyMatchingRows(from ...*Note) {
	for i, old := range o {
		for _, new := range from {
			if new.ID != old.ID {
				continue
			}

			o[i] = new
			break
		}
	}
}

// UpdateMod modifies an update query with "WHERE primary_key IN (o...)"
func (o NoteSlice) UpdateMod() bob.Mod[*dialect.UpdateQuery] {
	return bob.ModFunc[*dialect.UpdateQuery](func(q *dialect.UpdateQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Notes.BeforeUpdateHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Note:
				o.copyMatchingRows(retrieved)
			case []*Note:
				o.copyMatchingRows(retrieved...)
			case NoteSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Note or a slice of Note
				// then run the AfterUpdateHooks on the slice
				_, err = Notes.AfterUpdateHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

// DeleteMod modifies an delete query with "WHERE primary_key IN (o...)"
func (o NoteSlice) DeleteMod() bob.Mod[*dialect.DeleteQuery] {
	return bob.ModFunc[*dialect.DeleteQuery](func(q *dialect.DeleteQuery) {
		q.AppendHooks(func(ctx context.Context, exec bob.Executor) (context.Context, error) {
			return Notes.BeforeDeleteHooks.RunHooks(ctx, exec, o)
		})

		q.AppendLoader(bob.LoaderFunc(func(ctx context.Context, exec bob.Executor, retrieved any) error {
			var err error
			switch retrieved := retrieved.(type) {
			case *Note:
				o.copyMatchingRows(retrieved)
			case []*Note:
				o.copyMatchingRows(retrieved...)
			case NoteSlice:
				o.copyMatchingRows(retrieved...)
			default:
				// If the retrieved value is not a Note or a slice of Note
				// then run the AfterDeleteHooks on the slice
				_, err = Notes.AfterDeleteHooks.RunHooks(ctx, exec, o)
			}

			return err
		}))

		q.AppendWhere(o.pkIN())
	})
}

func (o NoteSlice) UpdateAll(ctx context.Context, exec bob.Executor, vals NoteSetter) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Notes.Update(vals.UpdateMod(), o.UpdateMod()).All(ctx, exec)
	return err
}

func (o NoteSlice) DeleteAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	_, err := Notes.Delete(o.DeleteMod()).Exec(ctx, exec)
	return err
}

func (o NoteSlice) ReloadAll(ctx context.Context, exec bob.Executor) error {
	if len(o) == 0 {
		return nil
	}

	o2, err := Notes.Query(sm.Where(o.pkIN())).All(ctx, exec)
	if err != nil {
		return err
	}

	o.copyMatchingRows(o2...)

	return nil
}
