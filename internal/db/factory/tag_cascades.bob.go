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

type TagCascadeMod interface {
	Apply(context.Context, *TagCascadeTemplate)
}

type TagCascadeModFunc func(context.Context, *TagCascadeTemplate)

func (f TagCascadeModFunc) Apply(ctx context.Context, n *TagCascadeTemplate) {
	f(ctx, n)
}

type TagCascadeModSlice []TagCascadeMod

func (mods TagCascadeModSlice) Apply(ctx context.Context, n *TagCascadeTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// TagCascadeTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type TagCascadeTemplate struct {
	TagID         func() uuid.UUID
	CascadedTagID func() uuid.UUID

	r tagCascadeR
	f *Factory

	alreadyPersisted bool
}

type tagCascadeR struct {
	CascadedTagTag *tagCascadeRCascadedTagTagR
	Tag            *tagCascadeRTagR
}

type tagCascadeRCascadedTagTagR struct {
	o *TagTemplate
}
type tagCascadeRTagR struct {
	o *TagTemplate
}

// Apply mods to the TagCascadeTemplate
func (o *TagCascadeTemplate) Apply(ctx context.Context, mods ...TagCascadeMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.TagCascade
// according to the relationships in the template. Nothing is inserted into the db
func (t TagCascadeTemplate) setModelRels(o *models.TagCascade) {
	if t.r.CascadedTagTag != nil {
		rel := t.r.CascadedTagTag.o.Build()
		o.CascadedTagID = rel.ID // h2
		o.R.CascadedTagTag = rel
	}

	if t.r.Tag != nil {
		rel := t.r.Tag.o.Build()
		o.TagID = rel.ID // h2
		o.R.Tag = rel
	}
}

// BuildSetter returns an *models.TagCascadeSetter
// this does nothing with the relationship templates
func (o TagCascadeTemplate) BuildSetter() *models.TagCascadeSetter {
	m := &models.TagCascadeSetter{}

	if o.TagID != nil {
		val := o.TagID()
		m.TagID = func() *uuid.UUID { return &val }()
	}
	if o.CascadedTagID != nil {
		val := o.CascadedTagID()
		m.CascadedTagID = func() *uuid.UUID { return &val }()
	}

	return m
}

// BuildManySetter returns an []*models.TagCascadeSetter
// this does nothing with the relationship templates
func (o TagCascadeTemplate) BuildManySetter(number int) []*models.TagCascadeSetter {
	m := make([]*models.TagCascadeSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.TagCascade
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagCascadeTemplate.Create
func (o TagCascadeTemplate) Build() *models.TagCascade {
	m := &models.TagCascade{}

	if o.TagID != nil {
		m.TagID = o.TagID()
	}
	if o.CascadedTagID != nil {
		m.CascadedTagID = o.CascadedTagID()
	}

	o.setModelRels(m)

	return m
}

// BuildMany returns an models.TagCascadeSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagCascadeTemplate.CreateMany
func (o TagCascadeTemplate) BuildMany(number int) models.TagCascadeSlice {
	m := make(models.TagCascadeSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableTagCascade(m *models.TagCascadeSetter) {
	if !(m.TagID != nil) {
		val := random_uuid_UUID(nil)
		m.TagID = func() *uuid.UUID { return &val }()
	}
	if !(m.CascadedTagID != nil) {
		val := random_uuid_UUID(nil)
		m.CascadedTagID = func() *uuid.UUID { return &val }()
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.TagCascade
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *TagCascadeTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.TagCascade) error {
	var err error

	return err
}

// Create builds a tagCascade and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *TagCascadeTemplate) Create(ctx context.Context, exec bob.Executor) (*models.TagCascade, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableTagCascade(opt)

	if o.r.CascadedTagTag == nil {
		TagCascadeMods.WithNewCascadedTagTag().Apply(ctx, o)
	}

	var rel0 *models.Tag

	if o.r.CascadedTagTag.o.alreadyPersisted {
		rel0 = o.r.CascadedTagTag.o.Build()
	} else {
		rel0, err = o.r.CascadedTagTag.o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	opt.CascadedTagID = func() *uuid.UUID { return &rel0.ID }()

	if o.r.Tag == nil {
		TagCascadeMods.WithNewTag().Apply(ctx, o)
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

	m, err := models.TagCascades.Insert(opt).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	m.R.CascadedTagTag = rel0
	m.R.Tag = rel1

	if err := o.insertOptRels(ctx, exec, m); err != nil {
		return nil, err
	}
	return m, err
}

// MustCreate builds a tagCascade and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *TagCascadeTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.TagCascade {
	m, err := o.Create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a tagCascade and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *TagCascadeTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.TagCascade {
	tb.Helper()
	m, err := o.Create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// CreateMany builds multiple tagCascades and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o TagCascadeTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.TagCascadeSlice, error) {
	var err error
	m := make(models.TagCascadeSlice, number)

	for i := range m {
		m[i], err = o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MustCreateMany builds multiple tagCascades and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o TagCascadeTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.TagCascadeSlice {
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple tagCascades and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o TagCascadeTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.TagCascadeSlice {
	tb.Helper()
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// TagCascade has methods that act as mods for the TagCascadeTemplate
var TagCascadeMods tagCascadeMods

type tagCascadeMods struct{}

func (m tagCascadeMods) RandomizeAllColumns(f *faker.Faker) TagCascadeMod {
	return TagCascadeModSlice{
		TagCascadeMods.RandomTagID(f),
		TagCascadeMods.RandomCascadedTagID(f),
	}
}

// Set the model columns to this value
func (m tagCascadeMods) TagID(val uuid.UUID) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.TagID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagCascadeMods) TagIDFunc(f func() uuid.UUID) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.TagID = f
	})
}

// Clear any values for the column
func (m tagCascadeMods) UnsetTagID() TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.TagID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCascadeMods) RandomTagID(f *faker.Faker) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.TagID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m tagCascadeMods) CascadedTagID(val uuid.UUID) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.CascadedTagID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagCascadeMods) CascadedTagIDFunc(f func() uuid.UUID) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.CascadedTagID = f
	})
}

// Clear any values for the column
func (m tagCascadeMods) UnsetCascadedTagID() TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.CascadedTagID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCascadeMods) RandomCascadedTagID(f *faker.Faker) TagCascadeMod {
	return TagCascadeModFunc(func(_ context.Context, o *TagCascadeTemplate) {
		o.CascadedTagID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

func (m tagCascadeMods) WithParentsCascading() TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		if isDone, _ := tagCascadeWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = tagCascadeWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewTagWithContext(ctx, TagMods.WithParentsCascading())
			m.WithCascadedTagTag(related).Apply(ctx, o)
		}
		{

			related := o.f.NewTagWithContext(ctx, TagMods.WithParentsCascading())
			m.WithTag(related).Apply(ctx, o)
		}
	})
}

func (m tagCascadeMods) WithCascadedTagTag(rel *TagTemplate) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.CascadedTagTag = &tagCascadeRCascadedTagTagR{
			o: rel,
		}
	})
}

func (m tagCascadeMods) WithNewCascadedTagTag(mods ...TagMod) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		related := o.f.NewTagWithContext(ctx, mods...)

		m.WithCascadedTagTag(related).Apply(ctx, o)
	})
}

func (m tagCascadeMods) WithExistingCascadedTagTag(em *models.Tag) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.CascadedTagTag = &tagCascadeRCascadedTagTagR{
			o: o.f.FromExistingTag(em),
		}
	})
}

func (m tagCascadeMods) WithoutCascadedTagTag() TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.CascadedTagTag = nil
	})
}

func (m tagCascadeMods) WithTag(rel *TagTemplate) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.Tag = &tagCascadeRTagR{
			o: rel,
		}
	})
}

func (m tagCascadeMods) WithNewTag(mods ...TagMod) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		related := o.f.NewTagWithContext(ctx, mods...)

		m.WithTag(related).Apply(ctx, o)
	})
}

func (m tagCascadeMods) WithExistingTag(em *models.Tag) TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.Tag = &tagCascadeRTagR{
			o: o.f.FromExistingTag(em),
		}
	})
}

func (m tagCascadeMods) WithoutTag() TagCascadeMod {
	return TagCascadeModFunc(func(ctx context.Context, o *TagCascadeTemplate) {
		o.r.Tag = nil
	})
}
