// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var TagCategories = Table[
	tagCategoryColumns,
	tagCategoryIndexes,
	tagCategoryForeignKeys,
	tagCategoryUniques,
	tagCategoryChecks,
]{
	Schema: "",
	Name:   "tag_categories",
	Columns: tagCategoryColumns{
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
		Color: column{
			Name:      "color",
			DBType:    "text",
			Default:   "'#888888'::text",
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
	Indexes: tagCategoryIndexes{
		TagCategoriesPkey: index{
			Type: "btree",
			Name: "tag_categories_pkey",
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
		IdxTagCategoriesCreatedAt: index{
			Type: "btree",
			Name: "idx_tag_categories_created_at",
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
		TagCategoriesNameKey: index{
			Type: "btree",
			Name: "tag_categories_name_key",
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
		Name:    "tag_categories_pkey",
		Columns: []string{"id"},
		Comment: "",
	},

	Uniques: tagCategoryUniques{
		TagCategoriesNameKey: constraint{
			Name:    "tag_categories_name_key",
			Columns: []string{"name"},
			Comment: "",
		},
	},

	Comment: "",
}

type tagCategoryColumns struct {
	ID          column
	Name        column
	Description column
	Color       column
	CreatedAt   column
	UpdatedAt   column
}

func (c tagCategoryColumns) AsSlice() []column {
	return []column{
		c.ID, c.Name, c.Description, c.Color, c.CreatedAt, c.UpdatedAt,
	}
}

type tagCategoryIndexes struct {
	TagCategoriesPkey         index
	IdxTagCategoriesCreatedAt index
	TagCategoriesNameKey      index
}

func (i tagCategoryIndexes) AsSlice() []index {
	return []index{
		i.TagCategoriesPkey, i.IdxTagCategoriesCreatedAt, i.TagCategoriesNameKey,
	}
}

type tagCategoryForeignKeys struct{}

func (f tagCategoryForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{}
}

type tagCategoryUniques struct {
	TagCategoriesNameKey constraint
}

func (u tagCategoryUniques) AsSlice() []constraint {
	return []constraint{
		u.TagCategoriesNameKey,
	}
}

type tagCategoryChecks struct{}

func (c tagCategoryChecks) AsSlice() []check {
	return []check{}
}
