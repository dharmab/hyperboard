// Code generated . DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package factory

import (
	"context"
	"database/sql"
	"testing"
	"time"

	models "github.com/dharmab/hyperboard/internal/db/models"
	"github.com/gofrs/uuid/v5"
	"github.com/jaswdr/faker/v2"
	"github.com/stephenafamo/bob"
)

type TagCategoryMod interface {
	Apply(context.Context, *TagCategoryTemplate)
}

type TagCategoryModFunc func(context.Context, *TagCategoryTemplate)

func (f TagCategoryModFunc) Apply(ctx context.Context, n *TagCategoryTemplate) {
	f(ctx, n)
}

type TagCategoryModSlice []TagCategoryMod

func (mods TagCategoryModSlice) Apply(ctx context.Context, n *TagCategoryTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// TagCategoryTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type TagCategoryTemplate struct {
	ID          func() uuid.UUID
	Name        func() string
	Description func() string
	Color       func() string
	CreatedAt   func() time.Time
	UpdatedAt   func() time.Time

	r tagCategoryR
	f *Factory

	alreadyPersisted bool
}

type tagCategoryR struct {
	Tags []*tagCategoryRTagsR
}

type tagCategoryRTagsR struct {
	number int
	o      *TagTemplate
}

// Apply mods to the TagCategoryTemplate
func (o *TagCategoryTemplate) Apply(ctx context.Context, mods ...TagCategoryMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.TagCategory
// according to the relationships in the template. Nothing is inserted into the db
func (t TagCategoryTemplate) setModelRels(o *models.TagCategory) {
	if t.r.Tags != nil {
		rel := models.TagSlice{}
		for _, r := range t.r.Tags {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.TagCategoryID = sql.Null[uuid.UUID]{V: o.ID, Valid: true} // h2
				rel.R.TagCategory = o
			}
			rel = append(rel, related...)
		}
		o.R.Tags = rel
	}
}

// BuildSetter returns an *models.TagCategorySetter
// this does nothing with the relationship templates
func (o TagCategoryTemplate) BuildSetter() *models.TagCategorySetter {
	m := &models.TagCategorySetter{}

	if o.ID != nil {
		val := o.ID()
		m.ID = func() *uuid.UUID { return &val }()
	}
	if o.Name != nil {
		val := o.Name()
		m.Name = func() *string { return &val }()
	}
	if o.Description != nil {
		val := o.Description()
		m.Description = func() *string { return &val }()
	}
	if o.Color != nil {
		val := o.Color()
		m.Color = func() *string { return &val }()
	}
	if o.CreatedAt != nil {
		val := o.CreatedAt()
		m.CreatedAt = func() *time.Time { return &val }()
	}
	if o.UpdatedAt != nil {
		val := o.UpdatedAt()
		m.UpdatedAt = func() *time.Time { return &val }()
	}

	return m
}

// BuildManySetter returns an []*models.TagCategorySetter
// this does nothing with the relationship templates
func (o TagCategoryTemplate) BuildManySetter(number int) []*models.TagCategorySetter {
	m := make([]*models.TagCategorySetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.TagCategory
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagCategoryTemplate.Create
func (o TagCategoryTemplate) Build() *models.TagCategory {
	m := &models.TagCategory{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Name != nil {
		m.Name = o.Name()
	}
	if o.Description != nil {
		m.Description = o.Description()
	}
	if o.Color != nil {
		m.Color = o.Color()
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

// BuildMany returns an models.TagCategorySlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagCategoryTemplate.CreateMany
func (o TagCategoryTemplate) BuildMany(number int) models.TagCategorySlice {
	m := make(models.TagCategorySlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableTagCategory(m *models.TagCategorySetter) {
	if !(m.Name != nil) {
		val := random_string(nil)
		m.Name = func() *string { return &val }()
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.TagCategory
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *TagCategoryTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.TagCategory) error {
	var err error

	isTagsDone, _ := tagCategoryRelTagsCtx.Value(ctx)
	if !isTagsDone && o.r.Tags != nil {
		ctx = tagCategoryRelTagsCtx.WithValue(ctx, true)
		for _, r := range o.r.Tags {
			if r.o.alreadyPersisted {
				m.R.Tags = append(m.R.Tags, r.o.Build())
			} else {
				rel0, err := r.o.CreateMany(ctx, exec, r.number)
				if err != nil {
					return err
				}

				err = m.AttachTags(ctx, exec, rel0...)
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

// Create builds a tagCategory and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *TagCategoryTemplate) Create(ctx context.Context, exec bob.Executor) (*models.TagCategory, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableTagCategory(opt)

	m, err := models.TagCategories.Insert(opt).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	if err := o.insertOptRels(ctx, exec, m); err != nil {
		return nil, err
	}
	return m, err
}

// MustCreate builds a tagCategory and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *TagCategoryTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.TagCategory {
	m, err := o.Create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a tagCategory and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *TagCategoryTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.TagCategory {
	tb.Helper()
	m, err := o.Create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// CreateMany builds multiple tagCategories and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o TagCategoryTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.TagCategorySlice, error) {
	var err error
	m := make(models.TagCategorySlice, number)

	for i := range m {
		m[i], err = o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MustCreateMany builds multiple tagCategories and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o TagCategoryTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.TagCategorySlice {
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple tagCategories and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o TagCategoryTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.TagCategorySlice {
	tb.Helper()
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// TagCategory has methods that act as mods for the TagCategoryTemplate
var TagCategoryMods tagCategoryMods

type tagCategoryMods struct{}

func (m tagCategoryMods) RandomizeAllColumns(f *faker.Faker) TagCategoryMod {
	return TagCategoryModSlice{
		TagCategoryMods.RandomID(f),
		TagCategoryMods.RandomName(f),
		TagCategoryMods.RandomDescription(f),
		TagCategoryMods.RandomColor(f),
		TagCategoryMods.RandomCreatedAt(f),
		TagCategoryMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m tagCategoryMods) ID(val uuid.UUID) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.ID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) IDFunc(f func() uuid.UUID) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetID() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomID(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.ID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m tagCategoryMods) Name(val string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Name = func() string { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) NameFunc(f func() string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Name = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetName() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Name = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomName(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Name = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagCategoryMods) Description(val string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Description = func() string { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) DescriptionFunc(f func() string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Description = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetDescription() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Description = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomDescription(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Description = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagCategoryMods) Color(val string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Color = func() string { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) ColorFunc(f func() string) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Color = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetColor() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Color = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomColor(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.Color = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagCategoryMods) CreatedAt(val time.Time) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.CreatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) CreatedAtFunc(f func() time.Time) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetCreatedAt() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomCreatedAt(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.CreatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

// Set the model columns to this value
func (m tagCategoryMods) UpdatedAt(val time.Time) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.UpdatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m tagCategoryMods) UpdatedAtFunc(f func() time.Time) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m tagCategoryMods) UnsetUpdatedAt() TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagCategoryMods) RandomUpdatedAt(f *faker.Faker) TagCategoryMod {
	return TagCategoryModFunc(func(_ context.Context, o *TagCategoryTemplate) {
		o.UpdatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

func (m tagCategoryMods) WithParentsCascading() TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		if isDone, _ := tagCategoryWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = tagCategoryWithParentsCascadingCtx.WithValue(ctx, true)
	})
}

func (m tagCategoryMods) WithTags(number int, related *TagTemplate) TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		o.r.Tags = []*tagCategoryRTagsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m tagCategoryMods) WithNewTags(number int, mods ...TagMod) TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		related := o.f.NewTagWithContext(ctx, mods...)
		m.WithTags(number, related).Apply(ctx, o)
	})
}

func (m tagCategoryMods) AddTags(number int, related *TagTemplate) TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		o.r.Tags = append(o.r.Tags, &tagCategoryRTagsR{
			number: number,
			o:      related,
		})
	})
}

func (m tagCategoryMods) AddNewTags(number int, mods ...TagMod) TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		related := o.f.NewTagWithContext(ctx, mods...)
		m.AddTags(number, related).Apply(ctx, o)
	})
}

func (m tagCategoryMods) AddExistingTags(existingModels ...*models.Tag) TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		for _, em := range existingModels {
			o.r.Tags = append(o.r.Tags, &tagCategoryRTagsR{
				o: o.f.FromExistingTag(em),
			})
		}
	})
}

func (m tagCategoryMods) WithoutTags() TagCategoryMod {
	return TagCategoryModFunc(func(ctx context.Context, o *TagCategoryTemplate) {
		o.r.Tags = nil
	})
}
