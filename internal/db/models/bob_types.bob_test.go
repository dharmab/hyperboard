// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"database/sql"
	"database/sql/driver"

	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
)

// Set the testDB to enable tests that use the database
var testDB bob.Transactor[bob.Tx]

// Make sure the type Note runs hooks after queries
var _ bob.HookableType = &Note{}

// Make sure the type Post runs hooks after queries
var _ bob.HookableType = &Post{}

// Make sure the type PostsTag runs hooks after queries
var _ bob.HookableType = &PostsTag{}

// Make sure the type TagAlias runs hooks after queries
var _ bob.HookableType = &TagAlias{}

// Make sure the type TagCascade runs hooks after queries
var _ bob.HookableType = &TagCascade{}

// Make sure the type TagCategory runs hooks after queries
var _ bob.HookableType = &TagCategory{}

// Make sure the type Tag runs hooks after queries
var _ bob.HookableType = &Tag{}

// Make sure the type uuid.UUID satisfies database/sql.Scanner
var _ sql.Scanner = (*uuid.UUID)(nil)

// Make sure the type uuid.UUID satisfies database/sql/driver.Valuer
var _ driver.Valuer = *new(uuid.UUID)
