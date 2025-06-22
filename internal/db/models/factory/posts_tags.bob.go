// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"

	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/bob"
)

type PostsTagMod interface {
	Apply(context.Context, *PostsTagTemplate)
}

type PostsTagModFunc func(context.Context, *PostsTagTemplate)

func (f PostsTagModFunc) Apply(ctx context.Context, n *PostsTagTemplate) {
	f(ctx, n)
}

type PostsTagModSlice []PostsTagMod

func (mods PostsTagModSlice) Apply(ctx context.Context, n *PostsTagTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// PostsTagTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type PostsTagTemplate struct {
	PostID func() int32
	TagID  func() int32

	r postsTagR
	f *Factory
}

type postsTagR struct {
	Post *postsTagRPostR
	Tag  *postsTagRTagR
}

type postsTagRPostR struct {
	o *PostTemplate
}
type postsTagRTagR struct {
	o *TagTemplate
}

// Apply mods to the PostsTagTemplate
func (o *PostsTagTemplate) Apply(ctx context.Context, mods ...PostsTagMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.PostsTag
// according to the relationships in the template. Nothing is inserted into the db
func (t PostsTagTemplate) setModelRels(o *models.PostsTag) {
	if t.r.Post != nil {
		rel := t.r.Post.o.Build()
		o.PostID = rel.ID // h2
		o.R.Post = rel
	}

	if t.r.Tag != nil {
		rel := t.r.Tag.o.Build()
		o.TagID = rel.ID // h2
		o.R.Tag = rel
	}
}

// BuildSetter returns an *models.PostsTagSetter
// this does nothing with the relationship templates
func (o PostsTagTemplate) BuildSetter() *models.PostsTagSetter {
	m := &models.PostsTagSetter{}

	if o.PostID != nil {
		val := o.PostID()
		m.PostID = &val
	}
	if o.TagID != nil {
		val := o.TagID()
		m.TagID = &val
	}

	return m
}

// BuildManySetter returns an []*models.PostsTagSetter
// this does nothing with the relationship templates
func (o PostsTagTemplate) BuildManySetter(number int) []*models.PostsTagSetter {
	m := make([]*models.PostsTagSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.PostsTag
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PostsTagTemplate.Create
func (o PostsTagTemplate) Build() *models.PostsTag {
	m := &models.PostsTag{}

	if o.PostID != nil {
		m.PostID = o.PostID()
	}
	if o.TagID != nil {
		m.TagID = o.TagID()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.PostsTagSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PostsTagTemplate.CreateMany
func (o PostsTagTemplate) BuildMany(number int) models.PostsTagSlice {
	m := make(models.PostsTagSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatablePostsTag(m *models.PostsTagSetter) {
	if m.PostID == nil {
		val := random_int32(nil)
		m.PostID = &val
	}
	if m.TagID == nil {
		val := random_int32(nil)
		m.TagID = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.PostsTag
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *PostsTagTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.PostsTag) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *PostsTagTemplate) Create(ctx context.Context, exec bob.Executor) (*models.PostsTag, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *PostsTagTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.PostsTag {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *PostsTagTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.PostsTag {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *PostsTagTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.PostsTag, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatablePostsTag(opt)

	if o.r.Post == nil {
		PostsTagMods.WithNewPost().Apply(ctx, o)
	}

	ctx, rel0, err := o.r.Post.o.create(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}

	opt.PostID = &rel0.ID

	if o.r.Tag == nil {
		PostsTagMods.WithNewTag().Apply(ctx, o)
	}

	ctx, rel1, err := o.r.Tag.o.create(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}

	opt.TagID = &rel1.ID

	m, err := models.PostsTags.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = postsTagCtx.WithValue(ctx, m)

	m.R.Post = rel0
	m.R.Tag = rel1

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o PostsTagTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.PostsTagSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o PostsTagTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.PostsTagSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o PostsTagTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.PostsTagSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o PostsTagTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.PostsTagSlice, error) {
	var err error
	m := make(models.PostsTagSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// PostsTag has methods that act as mods for the PostsTagTemplate
var PostsTagMods postsTagMods

type postsTagMods struct{}

func (m postsTagMods) RandomizeAllColumns(f *faker.Faker) PostsTagMod {
	return PostsTagModSlice{
		PostsTagMods.RandomPostID(f),
		PostsTagMods.RandomTagID(f),
	}
}

// Set the model columns to this value
func (m postsTagMods) PostID(val int32) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.PostID = func() int32 { return val }
	})
}

// Set the Column from the function
func (m postsTagMods) PostIDFunc(f func() int32) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.PostID = f
	})
}

// Clear any values for the column
func (m postsTagMods) UnsetPostID() PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.PostID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postsTagMods) RandomPostID(f *faker.Faker) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.PostID = func() int32 {
			return random_int32(f)
		}
	})
}

// Set the model columns to this value
func (m postsTagMods) TagID(val int32) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.TagID = func() int32 { return val }
	})
}

// Set the Column from the function
func (m postsTagMods) TagIDFunc(f func() int32) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.TagID = f
	})
}

// Clear any values for the column
func (m postsTagMods) UnsetTagID() PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.TagID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postsTagMods) RandomTagID(f *faker.Faker) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.TagID = func() int32 {
			return random_int32(f)
		}
	})
}

func (m postsTagMods) WithParentsCascading() PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		if isDone, _ := postsTagWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = postsTagWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewPost(ctx, PostMods.WithParentsCascading())
			m.WithPost(related).Apply(ctx, o)
		}
		{

			related := o.f.NewTag(ctx, TagMods.WithParentsCascading())
			m.WithTag(related).Apply(ctx, o)
		}
	})
}

func (m postsTagMods) WithPost(rel *PostTemplate) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Post = &postsTagRPostR{
			o: rel,
		}
	})
}

func (m postsTagMods) WithNewPost(mods ...PostMod) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		related := o.f.NewPost(ctx, mods...)

		m.WithPost(related).Apply(ctx, o)
	})
}

func (m postsTagMods) WithoutPost() PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Post = nil
	})
}

func (m postsTagMods) WithTag(rel *TagTemplate) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Tag = &postsTagRTagR{
			o: rel,
		}
	})
}

func (m postsTagMods) WithNewTag(mods ...TagMod) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		related := o.f.NewTag(ctx, mods...)

		m.WithTag(related).Apply(ctx, o)
	})
}

func (m postsTagMods) WithoutTag() PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Tag = nil
	})
}
