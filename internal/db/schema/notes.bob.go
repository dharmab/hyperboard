// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var Notes = Table[
	noteColumns,
	noteIndexes,
	noteForeignKeys,
	noteUniques,
	noteChecks,
]{
	Schema: "",
	Name:   "notes",
	Columns: noteColumns{
		ID: column{
			Name:      "id",
			DBType:    "uuid",
			Default:   "gen_random_uuid()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Title: column{
			Name:      "title",
			DBType:    "text",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Content: column{
			Name:      "content",
			DBType:    "text",
			Default:   "''::text",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		CreatedAt: column{
			Name:      "created_at",
			DBType:    "timestamp with time zone",
			Default:   "now()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		UpdatedAt: column{
			Name:      "updated_at",
			DBType:    "timestamp with time zone",
			Default:   "now()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
	},
	Indexes: noteIndexes{
		NotesPkey: index{
			Type: "btree",
			Name: "notes_pkey",
			Columns: []indexColumn{
				{
					Name:         "id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        true,
			Comment:       "",
			NullsFirst:    []bool{false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
		IdxNotesCreatedAt: index{
			Type: "btree",
			Name: "idx_notes_created_at",
			Columns: []indexColumn{
				{
					Name:         "created_at",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        false,
			Comment:       "",
			NullsFirst:    []bool{false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
	},
	PrimaryKey: &constraint{
		Name:    "notes_pkey",
		Columns: []string{"id"},
		Comment: "",
	},

	Comment: "",
}

type noteColumns struct {
	ID        column
	Title     column
	Content   column
	CreatedAt column
	UpdatedAt column
}

func (c noteColumns) AsSlice() []column {
	return []column{
		c.ID, c.Title, c.Content, c.CreatedAt, c.UpdatedAt,
	}
}

type noteIndexes struct {
	NotesPkey         index
	IdxNotesCreatedAt index
}

func (i noteIndexes) AsSlice() []index {
	return []index{
		i.NotesPkey, i.IdxNotesCreatedAt,
	}
}

type noteForeignKeys struct{}

func (f noteForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{}
}

type noteUniques struct{}

func (u noteUniques) AsSlice() []constraint {
	return []constraint{}
}

type noteChecks struct{}

func (c noteChecks) AsSlice() []check {
	return []check{}
}
