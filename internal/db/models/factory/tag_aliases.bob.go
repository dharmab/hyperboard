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

type TagAliasMod interface {
	Apply(context.Context, *TagAliasTemplate)
}

type TagAliasModFunc func(context.Context, *TagAliasTemplate)

func (f TagAliasModFunc) Apply(ctx context.Context, n *TagAliasTemplate) {
	f(ctx, n)
}

type TagAliasModSlice []TagAliasMod

func (mods TagAliasModSlice) Apply(ctx context.Context, n *TagAliasTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// TagAliasTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type TagAliasTemplate struct {
	ID        func() uuid.UUID
	TagID     func() uuid.UUID
	AliasName func() string
	CreatedAt func() time.Time

	r tagAliasR
	f *Factory
}

type tagAliasR struct {
	Tag *tagAliasRTagR
}

type tagAliasRTagR struct {
	o *TagTemplate
}

// Apply mods to the TagAliasTemplate
func (o *TagAliasTemplate) Apply(ctx context.Context, mods ...TagAliasMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.TagAlias
// according to the relationships in the template. Nothing is inserted into the db
func (t TagAliasTemplate) setModelRels(o *models.TagAlias) {
	if t.r.Tag != nil {
		rel := t.r.Tag.o.Build()
		rel.R.TagAliases = append(rel.R.TagAliases, o)
		o.TagID = rel.ID // h2
		o.R.Tag = rel
	}
}

// BuildSetter returns an *models.TagAliasSetter
// this does nothing with the relationship templates
func (o TagAliasTemplate) BuildSetter() *models.TagAliasSetter {
	m := &models.TagAliasSetter{}

	if o.ID != nil {
		val := o.ID()
		m.ID = &val
	}
	if o.TagID != nil {
		val := o.TagID()
		m.TagID = &val
	}
	if o.AliasName != nil {
		val := o.AliasName()
		m.AliasName = &val
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = &val
	}

	return m
}

// BuildManySetter returns an []*models.TagAliasSetter
// this does nothing with the relationship templates
func (o TagAliasTemplate) BuildManySetter(number int) []*models.TagAliasSetter {
	m := make([]*models.TagAliasSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.TagAlias
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagAliasTemplate.Create
func (o TagAliasTemplate) Build() *models.TagAlias {
	m := &models.TagAlias{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.TagID != nil {
		m.TagID = o.TagID()
	}
	if o.AliasName != nil {
		m.AliasName = o.AliasName()
	}
	if o.CreatedAt != nil {
		m.CreatedAt = o.CreatedAt()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.TagAliasSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagAliasTemplate.CreateMany
func (o TagAliasTemplate) BuildMany(number int) models.TagAliasSlice {
	m := make(models.TagAliasSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableTagAlias(m *models.TagAliasSetter) {
	if m.TagID == nil {
		val := random_uuid_UUID(nil)
		m.TagID = &val
	}
	if m.AliasName == nil {
		val := random_string(nil)
		m.AliasName = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.TagAlias
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *TagAliasTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.TagAlias) (context.Context, error) {
	var err error

	return ctx, err
}

// Create builds a tagAlias and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *TagAliasTemplate) Create(ctx context.Context, exec bob.Executor) (*models.TagAlias, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a tagAlias and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *TagAliasTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.TagAlias {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a tagAlias and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *TagAliasTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.TagAlias {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a tagAlias and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *TagAliasTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.TagAlias, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableTagAlias(opt)

	if o.r.Tag == nil {
		TagAliasMods.WithNewTag().Apply(ctx, o)
	}

	rel0, ok := tagCtx.Value(ctx)
	if !ok {
		ctx, rel0, err = o.r.Tag.o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	opt.TagID = &rel0.ID

	m, err := models.TagAliases.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = tagAliasCtx.WithValue(ctx, m)

	m.R.Tag = rel0

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple tagAliases and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o TagAliasTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.TagAliasSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple tagAliases and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o TagAliasTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.TagAliasSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple tagAliases and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o TagAliasTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.TagAliasSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple tagAliases and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o TagAliasTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.TagAliasSlice, error) {
	var err error
	m := make(models.TagAliasSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// TagAlias has methods that act as mods for the TagAliasTemplate
var TagAliasMods tagAliasMods

type tagAliasMods struct{}

func (m tagAliasMods) RandomizeAllColumns(f *faker.Faker) TagAliasMod {
	return TagAliasModSlice{
		TagAliasMods.RandomID(f),
		TagAliasMods.RandomTagID(f),
		TagAliasMods.RandomAliasName(f),
		TagAliasMods.RandomCreatedAt(f),
	}
}

// Set the model columns to this value
func (m tagAliasMods) ID(val uuid.UUID) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.ID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagAliasMods) IDFunc(f func() uuid.UUID) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m tagAliasMods) UnsetID() TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagAliasMods) RandomID(f *faker.Faker) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.ID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m tagAliasMods) TagID(val uuid.UUID) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.TagID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagAliasMods) TagIDFunc(f func() uuid.UUID) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.TagID = f
	})
}

// Clear any values for the column
func (m tagAliasMods) UnsetTagID() TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.TagID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagAliasMods) RandomTagID(f *faker.Faker) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.TagID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m tagAliasMods) AliasName(val string) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.AliasName = func() string { return val }
	})
}

// Set the Column from the function
func (m tagAliasMods) AliasNameFunc(f func() string) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.AliasName = f
	})
}

// Clear any values for the column
func (m tagAliasMods) UnsetAliasName() TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.AliasName = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagAliasMods) RandomAliasName(f *faker.Faker) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.AliasName = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagAliasMods) CreatedAt(val time.Time) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.CreatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m tagAliasMods) CreatedAtFunc(f func() time.Time) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m tagAliasMods) UnsetCreatedAt() TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagAliasMods) RandomCreatedAt(f *faker.Faker) TagAliasMod {
	return TagAliasModFunc(func(_ context.Context, o *TagAliasTemplate) {
		o.CreatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

func (m tagAliasMods) WithParentsCascading() TagAliasMod {
	return TagAliasModFunc(func(ctx context.Context, o *TagAliasTemplate) {
		if isDone, _ := tagAliasWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = tagAliasWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewTag(ctx, TagMods.WithParentsCascading())
			m.WithTag(related).Apply(ctx, o)
		}
	})
}

func (m tagAliasMods) WithTag(rel *TagTemplate) TagAliasMod {
	return TagAliasModFunc(func(ctx context.Context, o *TagAliasTemplate) {
		o.r.Tag = &tagAliasRTagR{
			o: rel,
		}
	})
}

func (m tagAliasMods) WithNewTag(mods ...TagMod) TagAliasMod {
	return TagAliasModFunc(func(ctx context.Context, o *TagAliasTemplate) {
		related := o.f.NewTag(ctx, mods...)

		m.WithTag(related).Apply(ctx, o)
	})
}

func (m tagAliasMods) WithoutTag() TagAliasMod {
	return TagAliasModFunc(func(ctx context.Context, o *TagAliasTemplate) {
		o.r.Tag = nil
	})
}
