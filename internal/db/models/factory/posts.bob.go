// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"testing"
	"time"

	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/bob"
)

type PostMod interface {
	Apply(context.Context, *PostTemplate)
}

type PostModFunc func(context.Context, *PostTemplate)

func (f PostModFunc) Apply(ctx context.Context, n *PostTemplate) {
	f(ctx, n)
}

type PostModSlice []PostMod

func (mods PostModSlice) Apply(ctx context.Context, n *PostTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// PostTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type PostTemplate struct {
	ID           func() uuid.UUID
	MimeType     func() string
	ContentURL   func() string
	ThumbnailURL func() string
	Note         func() string
	HasAudio     func() bool
	CreatedAt    func() time.Time
	UpdatedAt    func() time.Time

	r postR
	f *Factory
}

type postR struct {
	Tags []*postRTagsR
}

type postRTagsR struct {
	number int
	o      *TagTemplate
}

// Apply mods to the PostTemplate
func (o *PostTemplate) Apply(ctx context.Context, mods ...PostMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.Post
// according to the relationships in the template. Nothing is inserted into the db
func (t PostTemplate) setModelRels(o *models.Post) {
	if t.r.Tags != nil {
		rel := models.TagSlice{}
		for _, r := range t.r.Tags {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.R.Posts = append(rel.R.Posts, o)
			}
			rel = append(rel, related...)
		}
		o.R.Tags = rel
	}
}

// BuildSetter returns an *models.PostSetter
// this does nothing with the relationship templates
func (o PostTemplate) BuildSetter() *models.PostSetter {
	m := &models.PostSetter{}

	if o.ID != nil {
		val := o.ID()
		m.ID = &val
	}
	if o.MimeType != nil {
		val := o.MimeType()
		m.MimeType = &val
	}
	if o.ContentURL != nil {
		val := o.ContentURL()
		m.ContentURL = &val
	}
	if o.ThumbnailURL != nil {
		val := o.ThumbnailURL()
		m.ThumbnailURL = &val
	}
	if o.Note != nil {
		val := o.Note()
		m.Note = &val
	}
	if o.HasAudio != nil {
		val := o.HasAudio()
		m.HasAudio = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}
	if o.UpdatedAt != nil {
		val := o.UpdatedAt()
		m.UpdatedAt = &val
	}

	return m
}

// BuildManySetter returns an []*models.PostSetter
// this does nothing with the relationship templates
func (o PostTemplate) BuildManySetter(number int) []*models.PostSetter {
	m := make([]*models.PostSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Post
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PostTemplate.Create
func (o PostTemplate) Build() *models.Post {
	m := &models.Post{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.MimeType != nil {
		m.MimeType = o.MimeType()
	}
	if o.ContentURL != nil {
		m.ContentURL = o.ContentURL()
	}
	if o.ThumbnailURL != nil {
		m.ThumbnailURL = o.ThumbnailURL()
	}
	if o.Note != nil {
		m.Note = o.Note()
	}
	if o.HasAudio != nil {
		m.HasAudio = o.HasAudio()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}
	if o.UpdatedAt != nil {
		m.UpdatedAt = o.UpdatedAt()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.PostSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use PostTemplate.CreateMany
func (o PostTemplate) BuildMany(number int) models.PostSlice {
	m := make(models.PostSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatablePost(m *models.PostSetter) {
	if m.MimeType == nil {
		val := random_string(nil)
		m.MimeType = &val
	}
	if m.ContentURL == nil {
		val := random_string(nil)
		m.ContentURL = &val
	}
	if m.ThumbnailURL == nil {
		val := random_string(nil)
		m.ThumbnailURL = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Post
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *PostTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Post) (context.Context, error) {
	var err error

	isTagsDone, _ := postRelTagsCtx.Value(ctx)
	if !isTagsDone && o.r.Tags != nil {
		ctx = postRelTagsCtx.WithValue(ctx, true)
		for _, r := range o.r.Tags {
			var rel0 models.TagSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachTags(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, err
}

// Create builds a post and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *PostTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Post, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a post and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *PostTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.Post {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a post and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *PostTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.Post {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a post and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *PostTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.Post, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatablePost(opt)

	m, err := models.Posts.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = postCtx.WithValue(ctx, m)

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple posts and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o PostTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.PostSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple posts and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o PostTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.PostSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple posts and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o PostTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.PostSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple posts and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o PostTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.PostSlice, error) {
	var err error
	m := make(models.PostSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// Post has methods that act as mods for the PostTemplate
var PostMods postMods

type postMods struct{}

func (m postMods) RandomizeAllColumns(f *faker.Faker) PostMod {
	return PostModSlice{
		PostMods.RandomID(f),
		PostMods.RandomMimeType(f),
		PostMods.RandomContentURL(f),
		PostMods.RandomThumbnailURL(f),
		PostMods.RandomNote(f),
		PostMods.RandomHasAudio(f),
		PostMods.RandomCreatedAt(f),
		PostMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m postMods) ID(val uuid.UUID) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m postMods) IDFunc(f func() uuid.UUID) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m postMods) UnsetID() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomID(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) MimeType(val string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.MimeType = func() string { return val }
	})
}

// Set the Column from the function
func (m postMods) MimeTypeFunc(f func() string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.MimeType = f
	})
}

// Clear any values for the column
func (m postMods) UnsetMimeType() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.MimeType = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomMimeType(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.MimeType = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) ContentURL(val string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ContentURL = func() string { return val }
	})
}

// Set the Column from the function
func (m postMods) ContentURLFunc(f func() string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ContentURL = f
	})
}

// Clear any values for the column
func (m postMods) UnsetContentURL() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ContentURL = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomContentURL(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ContentURL = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) ThumbnailURL(val string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ThumbnailURL = func() string { return val }
	})
}

// Set the Column from the function
func (m postMods) ThumbnailURLFunc(f func() string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ThumbnailURL = f
	})
}

// Clear any values for the column
func (m postMods) UnsetThumbnailURL() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ThumbnailURL = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomThumbnailURL(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.ThumbnailURL = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) Note(val string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.Note = func() string { return val }
	})
}

// Set the Column from the function
func (m postMods) NoteFunc(f func() string) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.Note = f
	})
}

// Clear any values for the column
func (m postMods) UnsetNote() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.Note = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomNote(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.Note = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) HasAudio(val bool) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.HasAudio = func() bool { return val }
	})
}

// Set the Column from the function
func (m postMods) HasAudioFunc(f func() bool) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.HasAudio = f
	})
}

// Clear any values for the column
func (m postMods) UnsetHasAudio() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.HasAudio = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomHasAudio(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.HasAudio = func() bool {
			return random_bool(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) CreatedAt(val time.Time) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.CreatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m postMods) CreatedAtFunc(f func() time.Time) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m postMods) UnsetCreatedAt() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomCreatedAt(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.CreatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

// Set the model columns to this value
func (m postMods) UpdatedAt(val time.Time) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.UpdatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m postMods) UpdatedAtFunc(f func() time.Time) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m postMods) UnsetUpdatedAt() PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m postMods) RandomUpdatedAt(f *faker.Faker) PostMod {
	return PostModFunc(func(_ context.Context, o *PostTemplate) {
		o.UpdatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

func (m postMods) WithParentsCascading() PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		if isDone, _ := postWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = postWithParentsCascadingCtx.WithValue(ctx, true)
	})
}

func (m postMods) WithTags(number int, related *TagTemplate) PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		o.r.Tags = []*postRTagsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m postMods) WithNewTags(number int, mods ...TagMod) PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		related := o.f.NewTag(ctx, mods...)
		m.WithTags(number, related).Apply(ctx, o)
	})
}

func (m postMods) AddTags(number int, related *TagTemplate) PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		o.r.Tags = append(o.r.Tags, &postRTagsR{
			number: number,
			o:      related,
		})
	})
}

func (m postMods) AddNewTags(number int, mods ...TagMod) PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		related := o.f.NewTag(ctx, mods...)
		m.AddTags(number, related).Apply(ctx, o)
	})
}

func (m postMods) WithoutTags() PostMod {
	return PostModFunc(func(ctx context.Context, o *PostTemplate) {
		o.r.Tags = nil
	})
}
