// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var Tags = Table[
	tagColumns,
	tagIndexes,
	tagForeignKeys,
	tagUniques,
	tagChecks,
]{
	Schema: "",
	Name:   "tags",
	Columns: tagColumns{
		ID: column{
			Name:      "id",
			DBType:    "uuid",
			Default:   "gen_random_uuid()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Name: column{
			Name:      "name",
			DBType:    "text",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		Description: column{
			Name:      "description",
			DBType:    "text",
			Default:   "''::text",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		TagCategoryID: column{
			Name:      "tag_category_id",
			DBType:    "uuid",
			Default:   "NULL",
			Comment:   "",
			Nullable:  true,
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
	Indexes: tagIndexes{
		TagsPkey: index{
			Type: "btree",
			Name: "tags_pkey",
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
		IdxTagsCreatedAt: index{
			Type: "btree",
			Name: "idx_tags_created_at",
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
		TagsNameKey: index{
			Type: "btree",
			Name: "tags_name_key",
			Columns: []indexColumn{
				{
					Name:         "name",
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
	},
	PrimaryKey: &constraint{
		Name:    "tags_pkey",
		Columns: []string{"id"},
		Comment: "",
	},
	ForeignKeys: tagForeignKeys{
		TagsTagsTagCategoryIDFkey: foreignKey{
			constraint: constraint{
				Name:    "tags.tags_tag_category_id_fkey",
				Columns: []string{"tag_category_id"},
				Comment: "",
			},
			ForeignTable:   "tag_categories",
			ForeignColumns: []string{"id"},
		},
	},
	Uniques: tagUniques{
		TagsNameKey: constraint{
			Name:    "tags_name_key",
			Columns: []string{"name"},
			Comment: "",
		},
	},

	Comment: "",
}

type tagColumns struct {
	ID            column
	Name          column
	Description   column
	TagCategoryID column
	CreatedAt     column
	UpdatedAt     column
}

func (c tagColumns) AsSlice() []column {
	return []column{
		c.ID, c.Name, c.Description, c.TagCategoryID, c.CreatedAt, c.UpdatedAt,
	}
}

type tagIndexes struct {
	TagsPkey         index
	IdxTagsCreatedAt index
	TagsNameKey      index
}

func (i tagIndexes) AsSlice() []index {
	return []index{
		i.TagsPkey, i.IdxTagsCreatedAt, i.TagsNameKey,
	}
}

type tagForeignKeys struct {
	TagsTagsTagCategoryIDFkey foreignKey
}

func (f tagForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{
		f.TagsTagsTagCategoryIDFkey,
	}
}

type tagUniques struct {
	TagsNameKey constraint
}

func (u tagUniques) AsSlice() []constraint {
	return []constraint{
		u.TagsNameKey,
	}
}

type tagChecks struct{}

func (c tagChecks) AsSlice() []check {
	return []check{}
}
