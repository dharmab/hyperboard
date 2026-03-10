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

type NoteMod interface {
	Apply(context.Context, *NoteTemplate)
}

type NoteModFunc func(context.Context, *NoteTemplate)

func (f NoteModFunc) Apply(ctx context.Context, n *NoteTemplate) {
	f(ctx, n)
}

type NoteModSlice []NoteMod

func (mods NoteModSlice) Apply(ctx context.Context, n *NoteTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// NoteTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type NoteTemplate struct {
	ID        func() uuid.UUID
	Title     func() string
	Content   func() string
	CreatedAt func() time.Time
	UpdatedAt func() time.Time

	f *Factory

	alreadyPersisted bool
}

// Apply mods to the NoteTemplate
func (o *NoteTemplate) Apply(ctx context.Context, mods ...NoteMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.Note
// according to the relationships in the template. Nothing is inserted into the db
func (t NoteTemplate) setModelRels(o *models.Note) {}

// BuildSetter returns an *models.NoteSetter
// this does nothing with the relationship templates
func (o NoteTemplate) BuildSetter() *models.NoteSetter {
	m := &models.NoteSetter{}

	if o.ID != nil {
		val := o.ID()
		m.ID = func() *uuid.UUID { return &val }()
	}
	if o.Title != nil {
		val := o.Title()
		m.Title = func() *string { return &val }()
	}
	if o.Content != nil {
		val := o.Content()
		m.Content = func() *string { return &val }()
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

// BuildManySetter returns an []*models.NoteSetter
// this does nothing with the relationship templates
func (o NoteTemplate) BuildManySetter(number int) []*models.NoteSetter {
	m := make([]*models.NoteSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Note
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use NoteTemplate.Create
func (o NoteTemplate) Build() *models.Note {
	m := &models.Note{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Title != nil {
		m.Title = o.Title()
	}
	if o.Content != nil {
		m.Content = o.Content()
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

// BuildMany returns an models.NoteSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use NoteTemplate.CreateMany
func (o NoteTemplate) BuildMany(number int) models.NoteSlice {
	m := make(models.NoteSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableNote(m *models.NoteSetter) {
	if !(m.Title != nil) {
		val := random_string(nil)
		m.Title = func() *string { return &val }()
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Note
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *NoteTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Note) error {
	var err error

	return err
}

// Create builds a note and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *NoteTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Note, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableNote(opt)

	m, err := models.Notes.Insert(opt).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	if err := o.insertOptRels(ctx, exec, m); err != nil {
		return nil, err
	}
	return m, err
}

// MustCreate builds a note and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *NoteTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.Note {
	m, err := o.Create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a note and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *NoteTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.Note {
	tb.Helper()
	m, err := o.Create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// CreateMany builds multiple notes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o NoteTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.NoteSlice, error) {
	var err error
	m := make(models.NoteSlice, number)

	for i := range m {
		m[i], err = o.Create(ctx, exec)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// MustCreateMany builds multiple notes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o NoteTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.NoteSlice {
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple notes and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o NoteTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.NoteSlice {
	tb.Helper()
	m, err := o.CreateMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// Note has methods that act as mods for the NoteTemplate
var NoteMods noteMods

type noteMods struct{}

func (m noteMods) RandomizeAllColumns(f *faker.Faker) NoteMod {
	return NoteModSlice{
		NoteMods.RandomID(f),
		NoteMods.RandomTitle(f),
		NoteMods.RandomContent(f),
		NoteMods.RandomCreatedAt(f),
		NoteMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m noteMods) ID(val uuid.UUID) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.ID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m noteMods) IDFunc(f func() uuid.UUID) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m noteMods) UnsetID() NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m noteMods) RandomID(f *faker.Faker) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.ID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m noteMods) Title(val string) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Title = func() string { return val }
	})
}

// Set the Column from the function
func (m noteMods) TitleFunc(f func() string) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Title = f
	})
}

// Clear any values for the column
func (m noteMods) UnsetTitle() NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Title = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m noteMods) RandomTitle(f *faker.Faker) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Title = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m noteMods) Content(val string) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Content = func() string { return val }
	})
}

// Set the Column from the function
func (m noteMods) ContentFunc(f func() string) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Content = f
	})
}

// Clear any values for the column
func (m noteMods) UnsetContent() NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Content = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m noteMods) RandomContent(f *faker.Faker) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.Content = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m noteMods) CreatedAt(val time.Time) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.CreatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m noteMods) CreatedAtFunc(f func() time.Time) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m noteMods) UnsetCreatedAt() NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m noteMods) RandomCreatedAt(f *faker.Faker) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.CreatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

// Set the model columns to this value
func (m noteMods) UpdatedAt(val time.Time) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.UpdatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m noteMods) UpdatedAtFunc(f func() time.Time) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m noteMods) UnsetUpdatedAt() NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m noteMods) RandomUpdatedAt(f *faker.Faker) NoteMod {
	return NoteModFunc(func(_ context.Context, o *NoteTemplate) {
		o.UpdatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

func (m noteMods) WithParentsCascading() NoteMod {
	return NoteModFunc(func(ctx context.Context, o *NoteTemplate) {
		if isDone, _ := noteWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = noteWithParentsCascadingCtx.WithValue(ctx, true)
	})
}
