// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package schema

import "github.com/aarondl/opt/null"

var TagCascades = Table[
	tagCascadeColumns,
	tagCascadeIndexes,
	tagCascadeForeignKeys,
	tagCascadeUniques,
	tagCascadeChecks,
]{
	Schema: "",
	Name:   "tag_cascades",
	Columns: tagCascadeColumns{
		TagID: column{
			Name:      "tag_id",
			DBType:    "uuid",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
		CascadedTagID: column{
			Name:      "cascaded_tag_id",
			DBType:    "uuid",
			Default:   "",
			Comment:   "",
			Nullable:  false,
			Generated: false,
			AutoIncr:  false,
		},
	},
	Indexes: tagCascadeIndexes{
		TagCascadesPkey: index{
			Type: "btree",
			Name: "tag_cascades_pkey",
			Columns: []indexColumn{
				{
					Name:         "tag_id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
				{
					Name:         "cascaded_tag_id",
					Desc:         null.FromCond(false, true),
					IsExpression: false,
				},
			},
			Unique:        true,
			Comment:       "",
			NullsFirst:    []bool{false, false},
			NullsDistinct: false,
			Where:         "",
			Include:       []string{},
		},
		IdxTagCascadesCascadedTagID: index{
			Type: "btree",
			Name: "idx_tag_cascades_cascaded_tag_id",
			Columns: []indexColumn{
				{
					Name:         "cascaded_tag_id",
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
		Name:    "tag_cascades_pkey",
		Columns: []string{"tag_id", "cascaded_tag_id"},
		Comment: "",
	},
	ForeignKeys: tagCascadeForeignKeys{
		TagCascadesTagCascadesCascadedTagIDFkey: foreignKey{
			constraint: constraint{
				Name:    "tag_cascades.tag_cascades_cascaded_tag_id_fkey",
				Columns: []string{"cascaded_tag_id"},
				Comment: "",
			},
			ForeignTable:   "tags",
			ForeignColumns: []string{"id"},
		},
		TagCascadesTagCascadesTagIDFkey: foreignKey{
			constraint: constraint{
				Name:    "tag_cascades.tag_cascades_tag_id_fkey",
				Columns: []string{"tag_id"},
				Comment: "",
			},
			ForeignTable:   "tags",
			ForeignColumns: []string{"id"},
		},
	},

	Checks: tagCascadeChecks{
		CHKNoSelfCascade: check{
			constraint: constraint{
				Name:    "chk_no_self_cascade",
				Columns: []string{"tag_id", "cascaded_tag_id"},
				Comment: "",
			},
			Expression: "(tag_id <> cascaded_tag_id)",
		},
	},
	Comment: "",
}

type tagCascadeColumns struct {
	TagID         column
	CascadedTagID column
}

func (c tagCascadeColumns) AsSlice() []column {
	return []column{
		c.TagID, c.CascadedTagID,
	}
}

type tagCascadeIndexes struct {
	TagCascadesPkey             index
	IdxTagCascadesCascadedTagID index
}

func (i tagCascadeIndexes) AsSlice() []index {
	return []index{
		i.TagCascadesPkey, i.IdxTagCascadesCascadedTagID,
	}
}

type tagCascadeForeignKeys struct {
	TagCascadesTagCascadesCascadedTagIDFkey foreignKey
	TagCascadesTagCascadesTagIDFkey         foreignKey
}

func (f tagCascadeForeignKeys) AsSlice() []foreignKey {
	return []foreignKey{
		f.TagCascadesTagCascadesCascadedTagIDFkey, f.TagCascadesTagCascadesTagIDFkey,
	}
}

type tagCascadeUniques struct{}

func (u tagCascadeUniques) AsSlice() []constraint {
	return []constraint{}
}

type tagCascadeChecks struct {
	CHKNoSelfCascade check
}

func (c tagCascadeChecks) AsSlice() []check {
	return []check{
		c.CHKNoSelfCascade,
	}
}
