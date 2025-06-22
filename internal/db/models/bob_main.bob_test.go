// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models_test

import (
	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/stephenafamo/bob"
)

// Set the testDB to enable tests that use the database
var testDB bob.Transactor

// Make sure the type Post runs hooks after queries
var _ bob.HookableType = &models.Post{}

// Make sure the type PostsTag runs hooks after queries
var _ bob.HookableType = &models.PostsTag{}

// Make sure the type TagCategory runs hooks after queries
var _ bob.HookableType = &models.TagCategory{}

// Make sure the type Tag runs hooks after queries
var _ bob.HookableType = &models.Tag{}
