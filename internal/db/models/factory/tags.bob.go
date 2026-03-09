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

type TagMod interface {
	Apply(context.Context, *TagTemplate)
}

type TagModFunc func(context.Context, *TagTemplate)

func (f TagModFunc) Apply(ctx context.Context, n *TagTemplate) {
	f(ctx, n)
}

type TagModSlice []TagMod

func (mods TagModSlice) Apply(ctx context.Context, n *TagTemplate) {
	for _, f := range mods {
		f.Apply(ctx, n)
	}
}

// TagTemplate is an object representing the database table.
// all columns are optional and should be set by mods
type TagTemplate struct {
	ID            func() uuid.UUID
	Name          func() string
	Description   func() string
	TagCategoryID func() sql.Null[uuid.UUID]
	CreatedAt     func() time.Time
	UpdatedAt     func() time.Time

	r tagR
	f *Factory
}

type tagR struct {
	Posts       []*tagRPostsR
	TagAliases  []*tagRTagAliasesR
	TagCategory *tagRTagCategoryR
}

type tagRPostsR struct {
	number int
	o      *PostTemplate
}
type tagRTagAliasesR struct {
	number int
	o      *TagAliasTemplate
}
type tagRTagCategoryR struct {
	o *TagCategoryTemplate
}

// Apply mods to the TagTemplate
func (o *TagTemplate) Apply(ctx context.Context, mods ...TagMod) {
	for _, mod := range mods {
		mod.Apply(ctx, o)
	}
}

// setModelRels creates and sets the relationships on *models.Tag
// according to the relationships in the template. Nothing is inserted into the db
func (t TagTemplate) setModelRels(o *models.Tag) {
	if t.r.Posts != nil {
		rel := models.PostSlice{}
		for _, r := range t.r.Posts {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.R.Tags = append(rel.R.Tags, o)
			}
			rel = append(rel, related...)
		}
		o.R.Posts = rel
	}

	if t.r.TagAliases != nil {
		rel := models.TagAliasSlice{}
		for _, r := range t.r.TagAliases {
			related := r.o.BuildMany(r.number)
			for _, rel := range related {
				rel.TagID = o.ID // h2
				rel.R.Tag = o
			}
			rel = append(rel, related...)
		}
		o.R.TagAliases = rel
	}

	if t.r.TagCategory != nil {
		rel := t.r.TagCategory.o.Build()
		rel.R.Tags = append(rel.R.Tags, o)
		o.TagCategoryID = sql.Null[uuid.UUID]{V: rel.ID, Valid: true} // h2
		o.R.TagCategory = rel
	}
}

// BuildSetter returns an *models.TagSetter
// this does nothing with the relationship templates
func (o TagTemplate) BuildSetter() *models.TagSetter {
	m := &models.TagSetter{}

	if o.ID != nil {
		val := o.ID()
		m.ID = &val
	}
	if o.Name != nil {
		val := o.Name()
		m.Name = &val
	}
	if o.Description != nil {
		val := o.Description()
		m.Description = &val
	}
	if o.TagCategoryID != nil {
		val := o.TagCategoryID()
		m.TagCategoryID = &val
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

// BuildManySetter returns an []*models.TagSetter
// this does nothing with the relationship templates
func (o TagTemplate) BuildManySetter(number int) []*models.TagSetter {
	m := make([]*models.TagSetter, number)

	for i := range m {
		m[i] = o.BuildSetter()
	}

	return m
}

// Build returns an *models.Tag
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagTemplate.Create
func (o TagTemplate) Build() *models.Tag {
	m := &models.Tag{}

	if o.ID != nil {
		m.ID = o.ID()
	}
	if o.Name != nil {
		m.Name = o.Name()
	}
	if o.Description != nil {
		m.Description = o.Description()
	}
	if o.TagCategoryID != nil {
		m.TagCategoryID = o.TagCategoryID()
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

// BuildMany returns an models.TagSlice
// Related objects are also created and placed in the .R field
// NOTE: Objects are not inserted into the database. Use TagTemplate.CreateMany
func (o TagTemplate) BuildMany(number int) models.TagSlice {
	m := make(models.TagSlice, number)

	for i := range m {
		m[i] = o.Build()
	}

	return m
}

func ensureCreatableTag(m *models.TagSetter) {
	if m.Name == nil {
		val := random_string(nil)
		m.Name = &val
	}
}

// insertOptRels creates and inserts any optional the relationships on *models.Tag
// according to the relationships in the template.
// any required relationship should have already exist on the model
func (o *TagTemplate) insertOptRels(ctx context.Context, exec bob.Executor, m *models.Tag) (context.Context, error) {
	var err error

	isPostsDone, _ := tagRelPostsCtx.Value(ctx)
	if !isPostsDone && o.r.Posts != nil {
		ctx = tagRelPostsCtx.WithValue(ctx, true)
		for _, r := range o.r.Posts {
			var rel0 models.PostSlice
			ctx, rel0, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachPosts(ctx, exec, rel0...)
			if err != nil {
				return ctx, err
			}
		}
	}

	isTagAliasesDone, _ := tagRelTagAliasesCtx.Value(ctx)
	if !isTagAliasesDone && o.r.TagAliases != nil {
		ctx = tagRelTagAliasesCtx.WithValue(ctx, true)
		for _, r := range o.r.TagAliases {
			var rel1 models.TagAliasSlice
			ctx, rel1, err = r.o.createMany(ctx, exec, r.number)
			if err != nil {
				return ctx, err
			}

			err = m.AttachTagAliases(ctx, exec, rel1...)
			if err != nil {
				return ctx, err
			}
		}
	}

	isTagCategoryDone, _ := tagRelTagCategoryCtx.Value(ctx)
	if !isTagCategoryDone && o.r.TagCategory != nil {
		ctx = tagRelTagCategoryCtx.WithValue(ctx, true)
		var rel2 *models.TagCategory
		ctx, rel2, err = o.r.TagCategory.o.create(ctx, exec)
		if err != nil {
			return ctx, err
		}
		err = m.AttachTagCategory(ctx, exec, rel2)
		if err != nil {
			return ctx, err
		}

	}

	return ctx, err
}

// Create builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
func (o *TagTemplate) Create(ctx context.Context, exec bob.Executor) (*models.Tag, error) {
	_, m, err := o.create(ctx, exec)
	return m, err
}

// MustCreate builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o *TagTemplate) MustCreate(ctx context.Context, exec bob.Executor) *models.Tag {
	_, m, err := o.create(ctx, exec)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateOrFail builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o *TagTemplate) CreateOrFail(ctx context.Context, tb testing.TB, exec bob.Executor) *models.Tag {
	tb.Helper()
	_, m, err := o.create(ctx, exec)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// create builds a tag and inserts it into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted model
func (o *TagTemplate) create(ctx context.Context, exec bob.Executor) (context.Context, *models.Tag, error) {
	var err error
	opt := o.BuildSetter()
	ensureCreatableTag(opt)

	m, err := models.Tags.Insert(opt).One(ctx, exec)
	if err != nil {
		return ctx, nil, err
	}
	ctx = tagCtx.WithValue(ctx, m)

	ctx, err = o.insertOptRels(ctx, exec, m)
	return ctx, m, err
}

// CreateMany builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
func (o TagTemplate) CreateMany(ctx context.Context, exec bob.Executor, number int) (models.TagSlice, error) {
	_, m, err := o.createMany(ctx, exec, number)
	return m, err
}

// MustCreateMany builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// panics if an error occurs
func (o TagTemplate) MustCreateMany(ctx context.Context, exec bob.Executor, number int) models.TagSlice {
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		panic(err)
	}
	return m
}

// CreateManyOrFail builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// It calls `tb.Fatal(err)` on the test/benchmark if an error occurs
func (o TagTemplate) CreateManyOrFail(ctx context.Context, tb testing.TB, exec bob.Executor, number int) models.TagSlice {
	tb.Helper()
	_, m, err := o.createMany(ctx, exec, number)
	if err != nil {
		tb.Fatal(err)
		return nil
	}
	return m
}

// createMany builds multiple tags and inserts them into the database
// Relations objects are also inserted and placed in the .R field
// this returns a context that includes the newly inserted models
func (o TagTemplate) createMany(ctx context.Context, exec bob.Executor, number int) (context.Context, models.TagSlice, error) {
	var err error
	m := make(models.TagSlice, number)

	for i := range m {
		ctx, m[i], err = o.create(ctx, exec)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, m, nil
}

// Tag has methods that act as mods for the TagTemplate
var TagMods tagMods

type tagMods struct{}

func (m tagMods) RandomizeAllColumns(f *faker.Faker) TagMod {
	return TagModSlice{
		TagMods.RandomID(f),
		TagMods.RandomName(f),
		TagMods.RandomDescription(f),
		TagMods.RandomTagCategoryID(f),
		TagMods.RandomCreatedAt(f),
		TagMods.RandomUpdatedAt(f),
	}
}

// Set the model columns to this value
func (m tagMods) ID(val uuid.UUID) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.ID = func() uuid.UUID { return val }
	})
}

// Set the Column from the function
func (m tagMods) IDFunc(f func() uuid.UUID) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.ID = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetID() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.ID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomID(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.ID = func() uuid.UUID {
			return random_uuid_UUID(f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) Name(val string) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Name = func() string { return val }
	})
}

// Set the Column from the function
func (m tagMods) NameFunc(f func() string) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Name = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetName() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Name = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomName(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Name = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) Description(val string) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Description = func() string { return val }
	})
}

// Set the Column from the function
func (m tagMods) DescriptionFunc(f func() string) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Description = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetDescription() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Description = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomDescription(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.Description = func() string {
			return random_string(f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) TagCategoryID(val sql.Null[uuid.UUID]) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.TagCategoryID = func() sql.Null[uuid.UUID] { return val }
	})
}

// Set the Column from the function
func (m tagMods) TagCategoryIDFunc(f func() sql.Null[uuid.UUID]) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.TagCategoryID = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetTagCategoryID() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.TagCategoryID = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is sometimes null
func (m tagMods) RandomTagCategoryID(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.TagCategoryID = func() sql.Null[uuid.UUID] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_uuid_UUID(f)
			return sql.Null[uuid.UUID]{V: val, Valid: f.Bool()}
		}
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
// The generated value is never null
func (m tagMods) RandomTagCategoryIDNotNull(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.TagCategoryID = func() sql.Null[uuid.UUID] {
			if f == nil {
				f = &defaultFaker
			}

			val := random_uuid_UUID(f)
			return sql.Null[uuid.UUID]{V: val, Valid: true}
		}
	})
}

// Set the model columns to this value
func (m tagMods) CreatedAt(val time.Time) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.CreatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m tagMods) CreatedAtFunc(f func() time.Time) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.CreatedAt = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetCreatedAt() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.CreatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomCreatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.CreatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

// Set the model columns to this value
func (m tagMods) UpdatedAt(val time.Time) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.UpdatedAt = func() time.Time { return val }
	})
}

// Set the Column from the function
func (m tagMods) UpdatedAtFunc(f func() time.Time) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.UpdatedAt = f
	})
}

// Clear any values for the column
func (m tagMods) UnsetUpdatedAt() TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.UpdatedAt = nil
	})
}

// Generates a random value for the column using the given faker
// if faker is nil, a default faker is used
func (m tagMods) RandomUpdatedAt(f *faker.Faker) TagMod {
	return TagModFunc(func(_ context.Context, o *TagTemplate) {
		o.UpdatedAt = func() time.Time {
			return random_time_Time(f)
		}
	})
}

func (m tagMods) WithParentsCascading() TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		if isDone, _ := tagWithParentsCascadingCtx.Value(ctx); isDone {
			return
		}
		ctx = tagWithParentsCascadingCtx.WithValue(ctx, true)
		{

			related := o.f.NewTagCategory(ctx, TagCategoryMods.WithParentsCascading())
			m.WithTagCategory(related).Apply(ctx, o)
		}
	})
}

func (m tagMods) WithTagCategory(rel *TagCategoryTemplate) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.TagCategory = &tagRTagCategoryR{
			o: rel,
		}
	})
}

func (m tagMods) WithNewTagCategory(mods ...TagCategoryMod) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		related := o.f.NewTagCategory(ctx, mods...)

		m.WithTagCategory(related).Apply(ctx, o)
	})
}

func (m tagMods) WithoutTagCategory() TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.TagCategory = nil
	})
}

func (m tagMods) WithPosts(number int, related *PostTemplate) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.Posts = []*tagRPostsR{{
			number: number,
			o:      related,
		}}
	})
}

func (m tagMods) WithNewPosts(number int, mods ...PostMod) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		related := o.f.NewPost(ctx, mods...)
		m.WithPosts(number, related).Apply(ctx, o)
	})
}

func (m tagMods) AddPosts(number int, related *PostTemplate) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.Posts = append(o.r.Posts, &tagRPostsR{
			number: number,
			o:      related,
		})
	})
}

func (m tagMods) AddNewPosts(number int, mods ...PostMod) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		related := o.f.NewPost(ctx, mods...)
		m.AddPosts(number, related).Apply(ctx, o)
	})
}

func (m tagMods) WithoutPosts() TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.Posts = nil
	})
}

func (m tagMods) WithTagAliases(number int, related *TagAliasTemplate) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.TagAliases = []*tagRTagAliasesR{{
			number: number,
			o:      related,
		}}
	})
}

func (m tagMods) WithNewTagAliases(number int, mods ...TagAliasMod) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		related := o.f.NewTagAlias(ctx, mods...)
		m.WithTagAliases(number, related).Apply(ctx, o)
	})
}

func (m tagMods) AddTagAliases(number int, related *TagAliasTemplate) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.TagAliases = append(o.r.TagAliases, &tagRTagAliasesR{
			number: number,
			o:      related,
		})
	})
}

func (m tagMods) AddNewTagAliases(number int, mods ...TagAliasMod) TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		related := o.f.NewTagAlias(ctx, mods...)
		m.AddTagAliases(number, related).Apply(ctx, o)
	})
}

func (m tagMods) WithoutTagAliases() TagMod {
	return TagModFunc(func(ctx context.Context, o *TagTemplate) {
		o.r.TagAliases = nil
	})
}
