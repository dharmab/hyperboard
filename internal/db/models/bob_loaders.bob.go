// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/orm"
)

var Preload = getPreloaders()

type preloaders struct {
	Post        postPreloader
	PostsTag    postsTagPreloader
	TagAlias    tagAliasPreloader
	TagCategory tagCategoryPreloader
	Tag         tagPreloader
}

func getPreloaders() preloaders {
	return preloaders{
		Post:        buildPostPreloader(),
		PostsTag:    buildPostsTagPreloader(),
		TagAlias:    buildTagAliasPreloader(),
		TagCategory: buildTagCategoryPreloader(),
		Tag:         buildTagPreloader(),
	}
}

var (
	SelectThenLoad = getThenLoaders[*dialect.SelectQuery]()
	InsertThenLoad = getThenLoaders[*dialect.InsertQuery]()
	UpdateThenLoad = getThenLoaders[*dialect.UpdateQuery]()
)

type thenLoaders[Q orm.Loadable] struct {
	Post        postThenLoader[Q]
	PostsTag    postsTagThenLoader[Q]
	TagAlias    tagAliasThenLoader[Q]
	TagCategory tagCategoryThenLoader[Q]
	Tag         tagThenLoader[Q]
}

func getThenLoaders[Q orm.Loadable]() thenLoaders[Q] {
	return thenLoaders[Q]{
		Post:        buildPostThenLoader[Q](),
		PostsTag:    buildPostsTagThenLoader[Q](),
		TagAlias:    buildTagAliasThenLoader[Q](),
		TagCategory: buildTagCategoryThenLoader[Q](),
		Tag:         buildTagThenLoader[Q](),
	}
}

func thenLoadBuilder[Q orm.Loadable, T any](name string, f func(context.Context, bob.Executor, T, ...bob.Mod[*dialect.SelectQuery]) error) func(...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q] {
	return func(queryMods ...bob.Mod[*dialect.SelectQuery]) orm.Loader[Q] {
		return func(ctx context.Context, exec bob.Executor, retrieved any) error {
			loader, isLoader := retrieved.(T)
			if !isLoader {
				return fmt.Errorf("object %T cannot load %q", retrieved, name)
			}

			err := f(ctx, exec, loader, queryMods...)

			// Don't cause an issue due to missing relationships
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			return err
		}
	}
}
