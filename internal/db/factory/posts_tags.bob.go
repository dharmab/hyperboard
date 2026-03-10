// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"

	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
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
	PostID func() uuid.UUID
	TagID  func() uuid.UUID

	r postsTagR
	f *Factory

	alreadyPersisted bool
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
		m.PostID = func() *uuid.UUID { return &val }()
	}
	if o.TagID != nil {
		val := o.TagID()
		m.TagID = func() *uuid.UUID { return &val }()
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
	if !(m.PostID != nil) {
		val := random_uuid_UUID(nil)
		m.PostID = func() *uuid.UUID { return &val }()
	}
	if !(m.TagID != nil) {
		val := random_uuid_UUID(nil)
		m.TagID = func() *uuid.UUID { return &val }()
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.PostsTag
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *PostsTagTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.PostsTag) error {
	var err error

	return err
}

// Create builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *PostsTagTemplate) Create(ctx context.Context, exec bob.Executor) (*models.PostsTag, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatablePostsTag(opt)

	if o.r.Post == nil {
		PostsTagMods.WithNewPost().Apply(ctx, o)
	}

	var rel0 *models.Post

	if o.r.Post.o.alreadyPersisted {
		rel0 = o.r.Post.o.Build()
	} else {
		rel0, err = o.r.Post.o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	opt.PostID = func() *uuid.UUID { return &rel0.ID }()

	if o.r.Tag == nil {
		PostsTagMods.WithNewTag().Apply(ctx, o)
	}

	var rel1 *models.Tag

	if o.r.Tag.o.alreadyPersisted {
		rel1 = o.r.Tag.o.Build()
	} else {
		rel1, err = o.r.Tag.o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	opt.TagID = func() *uuid.UUID { return &rel1.ID }()

	m, err := models.PostsTags.Insert(opt).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	m.R.Post = rel0
	m.R.Tag = rel1

	if err := o.insertOptRels(ctx, exec, m); err != nil {
		return nil, err
	}
	return m, err
}

// MustCreate builds a postsTag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *PostsTagTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.PostsTag {
	m, err := o.Create(ctx, exec)
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
	m, err := o.Create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// CreateMany builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o PostsTagTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.PostsTagSlice, error) {
	var err error
	m := make(models.PostsTagSlice, number)

	for i := range m {
		m[i], err = o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MustCreateMany builds multiple postsTags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o PostsTagTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.PostsTagSlice {
	m, err := o.CreateMany(ctx, exec, number)
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
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
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
func (m postsTagMods) PostID(val uuid.UUID) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.PostID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m postsTagMods) PostIDFunc(f func() uuid.UUID) PostsTagMod {
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
		o.PostID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m postsTagMods) TagID(val uuid.UUID) PostsTagMod {
	return PostsTagModFunc(func(_ context.Context, o *PostsTagTemplate) {
		o.TagID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m postsTagMods) TagIDFunc(f func() uuid.UUID) PostsTagMod {
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
		o.TagID = func() uuid.UUID {
			return random_uuid_UUID(f)
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

			related := o.f.NewPostWithContext(ctx, PostMods.WithParentsCascading())
			m.WithPost(related).Apply(ctx, o)
		}
		{

			related := o.f.NewTagWithContext(ctx, TagMods.WithParentsCascading())
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
		related := o.f.NewPostWithContext(ctx, mods...)

		m.WithPost(related).Apply(ctx, o)
	})
}

func (m postsTagMods) WithExistingPost(em *models.Post) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Post = &postsTagRPostR{
			o: o.f.FromExistingPost(em),
		}
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
		related := o.f.NewTagWithContext(ctx, mods...)

		m.WithTag(related).Apply(ctx, o)
	})
}

func (m postsTagMods) WithExistingTag(em *models.Tag) PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Tag = &postsTagRTagR{
			o: o.f.FromExistingTag(em),
		}
	})
}

func (m postsTagMods) WithoutTag() PostsTagMod {
	return PostsTagModFunc(func(ctx context.Context, o *PostsTagTemplate) {
		o.r.Tag = nil
	})
}
