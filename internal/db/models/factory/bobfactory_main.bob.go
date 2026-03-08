// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import "context"

type Factory struct {
	baseNoteMods        NoteModSlice
	basePostMods        PostModSlice
	basePostsTagMods    PostsTagModSlice
	baseTagCategoryMods TagCategoryModSlice
	baseTagMods         TagModSlice
}

func New() *Factory {
	return &Factory{}
}

func (f *Factory) NewNote(ctx context.Context, mods ...NoteMod) *NoteTemplate {
	o := &NoteTemplate{f: f}

	if f != nil {
		f.baseNoteMods.Apply(ctx, o)
	}

	NoteModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) NewPost(ctx context.Context, mods ...PostMod) *PostTemplate {
	o := &PostTemplate{f: f}

	if f != nil {
		f.basePostMods.Apply(ctx, o)
	}

	PostModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) NewPostsTag(ctx context.Context, mods ...PostsTagMod) *PostsTagTemplate {
	o := &PostsTagTemplate{f: f}

	if f != nil {
		f.basePostsTagMods.Apply(ctx, o)
	}

	PostsTagModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) NewTagCategory(ctx context.Context, mods ...TagCategoryMod) *TagCategoryTemplate {
	o := &TagCategoryTemplate{f: f}

	if f != nil {
		f.baseTagCategoryMods.Apply(ctx, o)
	}

	TagCategoryModSlice(mods).Apply(ctx, o)

	return o
}

func (f *Factory) NewTag(ctx context.Context, mods ...TagMod) *TagTemplate {
	o := &TagTemplate{f: f}

	if f != nil {
		f.baseTagMods.Apply(ctx, o)
	}

	TagModSlice(mods).Apply(ctx, o)

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
