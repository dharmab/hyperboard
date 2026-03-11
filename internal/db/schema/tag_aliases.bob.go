// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var TagAliases = Table[
	tagAliasColumns,
	tagAliasIndexes,
	tagAliasForeignKeys,
	tagAliasUniques,
	tagAliasChecks,
]{
	Schema: "",
	Name:   "tag_aliases",
	Columns: tagAliasColumns{
		ID: column{
			Name:      "id",
			DBType:    "uuid",
			Default:   "gen_random_uuid()",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		TagID: column{
			Name:      "tag_id",
			DBType:    "uuid",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		AliasName: column{
			Name:      "alias_name",
			DBType:    "text",
			Default:   "",
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
	},
	Indexes: tagAliasIndexes{
		TagAliasesPkey: index{
			Type: "btree",
			Name: "tag_aliases_pkey",
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
		IdxTagAliasesTagID: index{
			Type: "btree",
			Name: "idx_tag_aliases_tag_id",
			Columns: []indexColumn{
				{
					Name:         "tag_id",
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
		TagAliasesAliasNameKey: index{
			Type: "btree",
			Name: "tag_aliases_alias_name_key",
			Columns: []indexColumn{
				{
					Name:         "alias_name",
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
		Name:    "tag_aliases_pkey",
		Columns: []string{"id"},
		Comment: "",
	},
	ForeignKeys: tagAliasForeignKeys{
		TagAliasesTagAliasesTagIDFkey: foreignKey{
			constraint: constraint{
				Name:    "tag_aliases.tag_aliases_tag_id_fkey",
				Columns: []string{"tag_id"},
				Comment: "",
			},
			ForeignTable:   "tags",
			ForeignColumns: []string{"id"},
		},
	},
	Uniques: tagAliasUniques{
		TagAliasesAliasNameKey: constraint{
			Name:    "tag_aliases_alias_name_key",
			Columns: []string{"alias_name"},
			Comment: "",
		},
	},

	Comment: "",
}

type tagAliasColumns struct {
	ID        column
	TagID     column
	AliasName column
	CreatedAt column
}

func (c tagAliasColumns) AsSlice() []column {
	return []column{
		c.ID, c.TagID, c.AliasName, c.CreatedAt,
	}
}

type tagAliasIndexes struct {
	TagAliasesPkey         index
	IdxTagAliasesTagID     index
	TagAliasesAliasNameKey index
}

func (i tagAliasIndexes) AsSlice() []index {
	return []index{
		i.TagAliasesPkey, i.IdxTagAliasesTagID, i.TagAliasesAliasNameKey,
	}
}

type tagAliasForeignKeys struct {
	TagAliasesTagAliasesTagIDFkey foreignKey
}

func (f tagAliasForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{
		f.TagAliasesTagAliasesTagIDFkey,
	}
}

type tagAliasUniques struct {
	TagAliasesAliasNameKey constraint
}

func (u tagAliasUniques) AsSlice() []constraint {
	return []constraint{
		u.TagAliasesAliasNameKey,
	}
}

type tagAliasChecks struct{}

func (c tagAliasChecks) AsSlice() []check {
	return []check{}
}
