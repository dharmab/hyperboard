// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"

	models "github.com/dharmab/hyperboard/internal/db/models"
)

type contextKey string

var (
	// Table context

	noteCtx        = newContextual[*models.Note]("note")
	postCtx        = newContextual[*models.Post]("post")
	postsTagCtx    = newContextual[*models.PostsTag]("postsTag")
	tagCategoryCtx = newContextual[*models.TagCategory]("tagCategory")
	tagCtx         = newContextual[*models.Tag]("tag")

	// Relationship Contexts for notes
	noteWithParentsCascadingCtx = newContextual[bool]("noteWithParentsCascading")

	// Relationship Contexts for posts
	postWithParentsCascadingCtx = newContextual[bool]("postWithParentsCascading")
	postRelTagsCtx              = newContextual[bool]("posts.tags.posts_tags.posts_tags_post_id_fkeyposts_tags.posts_tags_tag_id_fkey")

	// Relationship Contexts for posts_tags
	postsTagWithParentsCascadingCtx = newContextual[bool]("postsTagWithParentsCascading")
	postsTagRelPostCtx              = newContextual[bool]("posts.posts_tags.posts_tags.posts_tags_post_id_fkey")
	postsTagRelTagCtx               = newContextual[bool]("posts_tags.tags.posts_tags.posts_tags_tag_id_fkey")

	// Relationship Contexts for tag_categories
	tagCategoryWithParentsCascadingCtx = newContextual[bool]("tagCategoryWithParentsCascading")
	tagCategoryRelTagsCtx              = newContextual[bool]("tag_categories.tags.tags.tags_tag_category_id_fkey")

	// Relationship Contexts for tags
	tagWithParentsCascadingCtx = newContextual[bool]("tagWithParentsCascading")
	tagRelPostsCtx             = newContextual[bool]("posts.tags.posts_tags.posts_tags_post_id_fkeyposts_tags.posts_tags_tag_id_fkey")
	tagRelTagCategoryCtx       = newContextual[bool]("tag_categories.tags.tags.tags_tag_category_id_fkey")
)

// Contextual is a convienience wrapper around context.WithValue and context.Value
type contextual[V any] struct {
	key contextKey
}

func newContextual[V any](key string) contextual[V] {
	return contextual[V]{key: contextKey(key)}
}

func (k contextual[V]) WithValue(ctx context.Context, val V) context.Context {
	return context.WithValue(ctx, k.key, val)
}

func (k contextual[V]) Value(ctx context.Context) (V, bool) {
	v, ok := ctx.Value(k.key).(V)
	return v, ok
}
