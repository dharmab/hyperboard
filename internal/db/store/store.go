package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/search"
	"github.com/gofrs/uuid/v5"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAliasConflict = errors.New("alias conflicts with existing tag name")
)

// SQLStore combines all sub-interfaces for database operations.
type SQLStore interface {
	Pinger
	NoteStore
	TagCategoryStore
	TagStore
	PostStore
}

// Pinger provides database connectivity checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// NoteStore provides CRUD operations for notes.
type NoteStore interface {
	ListNotes(ctx context.Context) (models.NoteSlice, error)
	GetNote(ctx context.Context, id uuid.UUID) (*models.Note, error)
	CreateNote(ctx context.Context, title, content string) (*models.Note, error)
	UpdateNote(ctx context.Context, id uuid.UUID, title, content string) (*models.Note, error)
	DeleteNote(ctx context.Context, id uuid.UUID) error
}

// TagCategoryStore provides CRUD operations for tag categories.
type TagCategoryStore interface {
	ListTagCategories(ctx context.Context, cursor *string, limit int) (models.TagCategorySlice, bool, error)
	GetTagCategory(ctx context.Context, name string) (*models.TagCategory, error)
	UpsertTagCategory(ctx context.Context, urlName string, input TagCategoryInput, now time.Time) (*models.TagCategory, bool, error)
	DeleteTagCategory(ctx context.Context, name string) error
	GetTagCountsByCategory(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error)
}

// TagCategoryInput holds the fields for creating or updating a tag category.
type TagCategoryInput struct {
	Name        string
	Description string
	Color       string
}

// TagStore provides CRUD operations for tags.
type TagStore interface {
	ListTags(ctx context.Context, cursor *string, limit int) (models.TagSlice, bool, error)
	GetTag(ctx context.Context, name string) (*models.Tag, error)
	UpsertTag(ctx context.Context, urlName string, input TagInput, now time.Time) (*models.Tag, bool, error)
	DeleteTag(ctx context.Context, name string) error
	GetTagPostCounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error)
	GetTagAliases(ctx context.Context, ids ...uuid.UUID) (map[uuid.UUID][]string, error)
	ResolveAlias(ctx context.Context, name string) (string, error)
	ConvertTagToAlias(ctx context.Context, sourceName, targetName string) (*ConvertTagToAliasResult, error)
}

// TagInput holds the fields for creating or updating a tag.
type TagInput struct {
	Name          string
	Description   string
	Category      *string
	Aliases       []string
	TagCategoryID sql.Null[uuid.UUID]
}

// ConvertTagToAliasResult holds the result of converting a tag to an alias.
type ConvertTagToAliasResult struct {
	Tag     *models.Tag
	Aliases []string
}

// PostStore provides CRUD operations for posts.
type PostStore interface {
	ListPosts(ctx context.Context, params ListPostsParams) (models.PostSlice, bool, error)
	GetPost(ctx context.Context, id uuid.UUID) (*models.Post, error)
	CreatePost(ctx context.Context, setter *models.PostSetter) (*models.Post, error)
	UpdatePost(ctx context.Context, id uuid.UUID, note string, tagNames []string, now time.Time) (*models.Post, error)
	UpdatePostContent(ctx context.Context, id uuid.UUID, setter *models.PostSetter) (*models.Post, error)
	UpdatePostThumbnail(ctx context.Context, id uuid.UUID, thumbnailURL string, now time.Time) (*models.Post, error)
	DeletePost(ctx context.Context, id uuid.UUID) (*models.Post, error)
	FindPostBySha256(ctx context.Context, hash string) (*models.Post, error)
	FindSimilarPosts(ctx context.Context, excludeID uuid.UUID, pHash int64, limit int) (models.PostSlice, error)
}

// ListPostsParams holds parameters for listing posts.
type ListPostsParams struct {
	Query        search.Query
	Limit        int
	CursorTime   *time.Time
	CursorID     *uuid.UUID
	RandomSeed   *int64
	RandomOffset int
}
