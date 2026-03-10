// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"database/sql"
	"time"

	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
)

type Factory struct {
	baseNoteMods        NoteModSlice
	basePostMods        PostModSlice
	basePostsTagMods    PostsTagModSlice
	baseTagAliasMods    TagAliasModSlice
	baseTagCategoryMods TagCategoryModSlice
	baseTagMods         TagModSlice
}

func New() *Factory {
	return &Factory{}
}

func (f *Factory) NewNote(mods ...NoteMod) *NoteTemplate {
	return f.NewNoteWithContext(context.Background(), mods...)
}

func (f *Factory) NewNoteWithContext(ctx context.Context, mods ...NoteMod) *NoteTemplate {
	o := &NoteTemplate{f: f}

	if f != nil {
		f.baseNoteMods.Apply(ctx, o)
	}

	NoteModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingNote(m *models.Note) *NoteTemplate {
	o := &NoteTemplate{f: f, alreadyPersisted: true}

	o.ID = func() uuid.UUID { return m.ID }
	o.Title = func() string { return m.Title }
	o.Content = func() string { return m.Content }
	o.CreatedAt = func() time.Time { return m.CreatedAt }
	o.UpdatedAt = func() time.Time { return m.UpdatedAt }

	return o
}

func (f *Factory) NewPost(mods ...PostMod) *PostTemplate {
	return f.NewPostWithContext(context.Background(), mods...)
}

func (f *Factory) NewPostWithContext(ctx context.Context, mods ...PostMod) *PostTemplate {
	o := &PostTemplate{f: f}

	if f != nil {
		f.basePostMods.Apply(ctx, o)
	}

	PostModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingPost(m *models.Post) *PostTemplate {
	o := &PostTemplate{f: f, alreadyPersisted: true}

	o.ID = func() uuid.UUID { return m.ID }
	o.MimeType = func() string { return m.MimeType }
	o.ContentURL = func() string { return m.ContentURL }
	o.ThumbnailURL = func() string { return m.ThumbnailURL }
	o.Note = func() string { return m.Note }
	o.HasAudio = func() bool { return m.HasAudio }
	o.Sha256 = func() string { return m.Sha256 }
	o.Phash = func() sql.Null[int64] { return m.Phash }
	o.CreatedAt = func() time.Time { return m.CreatedAt }
	o.UpdatedAt = func() time.Time { return m.UpdatedAt }

	ctx := context.Background()
	if len(m.R.Tags) > 0 {
		PostMods.AddExistingTags(m.R.Tags...).Apply(ctx, o)
	}

	return o
}

func (f *Factory) NewPostsTag(mods ...PostsTagMod) *PostsTagTemplate {
	return f.NewPostsTagWithContext(context.Background(), mods...)
}

func (f *Factory) NewPostsTagWithContext(ctx context.Context, mods ...PostsTagMod) *PostsTagTemplate {
	o := &PostsTagTemplate{f: f}

	if f != nil {
		f.basePostsTagMods.Apply(ctx, o)
	}

	PostsTagModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingPostsTag(m *models.PostsTag) *PostsTagTemplate {
	o := &PostsTagTemplate{f: f, alreadyPersisted: true}

	o.PostID = func() uuid.UUID { return m.PostID }
	o.TagID = func() uuid.UUID { return m.TagID }

	ctx := context.Background()
	if m.R.Post != nil {
		PostsTagMods.WithExistingPost(m.R.Post).Apply(ctx, o)
	}
	if m.R.Tag != nil {
		PostsTagMods.WithExistingTag(m.R.Tag).Apply(ctx, o)
	}

	return o
}

func (f *Factory) NewTagAlias(mods ...TagAliasMod) *TagAliasTemplate {
	return f.NewTagAliasWithContext(context.Background(), mods...)
}

func (f *Factory) NewTagAliasWithContext(ctx context.Context, mods ...TagAliasMod) *TagAliasTemplate {
	o := &TagAliasTemplate{f: f}

	if f != nil {
		f.baseTagAliasMods.Apply(ctx, o)
	}

	TagAliasModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingTagAlias(m *models.TagAlias) *TagAliasTemplate {
	o := &TagAliasTemplate{f: f, alreadyPersisted: true}

	o.ID = func() uuid.UUID { return m.ID }
	o.TagID = func() uuid.UUID { return m.TagID }
	o.AliasName = func() string { return m.AliasName }
	o.CreatedAt = func() time.Time { return m.CreatedAt }

	ctx := context.Background()
	if m.R.Tag != nil {
		TagAliasMods.WithExistingTag(m.R.Tag).Apply(ctx, o)
	}

	return o
}

func (f *Factory) NewTagCategory(mods ...TagCategoryMod) *TagCategoryTemplate {
	return f.NewTagCategoryWithContext(context.Background(), mods...)
}

func (f *Factory) NewTagCategoryWithContext(ctx context.Context, mods ...TagCategoryMod) *TagCategoryTemplate {
	o := &TagCategoryTemplate{f: f}

	if f != nil {
		f.baseTagCategoryMods.Apply(ctx, o)
	}

	TagCategoryModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingTagCategory(m *models.TagCategory) *TagCategoryTemplate {
	o := &TagCategoryTemplate{f: f, alreadyPersisted: true}

	o.ID = func() uuid.UUID { return m.ID }
	o.Name = func() string { return m.Name }
	o.Description = func() string { return m.Description }
	o.Color = func() string { return m.Color }
	o.CreatedAt = func() time.Time { return m.CreatedAt }
	o.UpdatedAt = func() time.Time { return m.UpdatedAt }

	ctx := context.Background()
	if len(m.R.Tags) > 0 {
		TagCategoryMods.AddExistingTags(m.R.Tags...).Apply(ctx, o)
	}

	return o
}

func (f *Factory) NewTag(mods ...TagMod) *TagTemplate {
	return f.NewTagWithContext(context.Background(), mods...)
}

func (f *Factory) NewTagWithContext(ctx context.Context, mods ...TagMod) *TagTemplate {
	o := &TagTemplate{f: f}

	if f != nil {
		f.baseTagMods.Apply(ctx, o)
	}

	TagModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) FromExistingTag(m *models.Tag) *TagTemplate {
	o := &TagTemplate{f: f, alreadyPersisted: true}

	o.ID = func() uuid.UUID { return m.ID }
	o.Name = func() string { return m.Name }
	o.Description = func() string { return m.Description }
	o.TagCategoryID = func() sql.Null[uuid.UUID] { return m.TagCategoryID }
	o.CreatedAt = func() time.Time { return m.CreatedAt }
	o.UpdatedAt = func() time.Time { return m.UpdatedAt }

	ctx := context.Background()
	if len(m.R.Posts) > 0 {
		TagMods.AddExistingPosts(m.R.Posts...).Apply(ctx, o)
	}
	if len(m.R.TagAliases) > 0 {
		TagMods.AddExistingTagAliases(m.R.TagAliases...).Apply(ctx, o)
	}
	if m.R.TagCategory != nil {
		TagMods.WithExistingTagCategory(m.R.TagCategory).Apply(ctx, o)
	}

	return o
}

func (f *Factory) ClearBaseNoteMods() {
	f.baseNoteMods = nil
}

func (f *Factory) AddBaseNoteMod(mods ...NoteMod) {
	f.baseNoteMods = append(f.baseNoteMods, mods...)
}

func (f *Factory) ClearBasePostMods() {
	f.basePostMods = nil
}

func (f *Factory) AddBasePostMod(mods ...PostMod) {
	f.basePostMods = append(f.basePostMods, mods...)
}

func (f *Factory) ClearBasePostsTagMods() {
	f.basePostsTagMods = nil
}

func (f *Factory) AddBasePostsTagMod(mods ...PostsTagMod) {
	f.basePostsTagMods = append(f.basePostsTagMods, mods...)
}

func (f *Factory) ClearBaseTagAliasMods() {
	f.baseTagAliasMods = nil
}

func (f *Factory) AddBaseTagAliasMod(mods ...TagAliasMod) {
	f.baseTagAliasMods = append(f.baseTagAliasMods, mods...)
}

func (f *Factory) ClearBaseTagCategoryMods() {
	f.baseTagCategoryMods = nil
}

func (f *Factory) AddBaseTagCategoryMod(mods ...TagCategoryMod) {
	f.baseTagCategoryMods = append(f.baseTagCategoryMods, mods...)
}

func (f *Factory) ClearBaseTagMods() {
	f.baseTagMods = nil
}

func (f *Factory) AddBaseTagMod(mods ...TagMod) {
	f.baseTagMods = append(f.baseTagMods, mods...)
}
