package api

import (
	"testing"
	"time"

	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/pkg/types"
)

func TestParseSearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect search.Query
	}{
		{
			name:  "empty string",
			input: "",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "single tag",
			input: "landscape",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape"},
			},
		},
		{
			name:  "multiple tags",
			input: "landscape,portrait",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "tags with whitespace",
			input: " landscape , portrait ",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "sort created",
			input: "sort:created",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortCreatedAt,
			},
		},
		{
			name:  "sort updated",
			input: "sort:updated",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortUpdatedAt,
			},
		},
		{
			name:  "sort random",
			input: "sort:random",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortRandom,
			},
		},
		{
			name:  "invalid sort ignored",
			input: "sort:invalid",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "tagged true",
			input: "tagged:true",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(true),
			},
		},
		{
			name:  "tagged false",
			input: "tagged:false",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(false),
			},
		},
		{
			name:  "tagged true then false uses last value",
			input: "tagged:true,tagged:false",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Tagged:       new(false),
			},
		},
		{
			name:  "type image",
			input: "type:image",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeImage:    true,
			},
		},
		{
			name:  "type video",
			input: "type:video",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeVideo:    true,
			},
		},
		{
			name:  "type audio",
			input: "type:audio",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				TypeAudio:    true,
			},
		},
		{
			name:  "order asc",
			input: "order:asc",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Order:        search.OrderAsc,
			},
		},
		{
			name:  "order desc",
			input: "order:desc",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Order:        search.OrderDesc,
			},
		},
		{
			name:  "order invalid ignored",
			input: "order:invalid",
			expect: search.Query{
				IncludedTags: []types.TagName{},
			},
		},
		{
			name:  "created_after",
			input: "created_after:2025-01-01T00:00:00Z",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				CreatedAfter: func() *time.Time { t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "created_before",
			input: "created_before:2025-06-15T12:30:00Z",
			expect: search.Query{
				IncludedTags:  []types.TagName{},
				CreatedBefore: func() *time.Time { t := time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "combined order and created_after with sort",
			input: "sort:created,order:asc,created_after:2025-01-01T00:00:00Z",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				Sort:         search.SortCreatedAt,
				Order:        search.OrderAsc,
				CreatedAfter: func() *time.Time { t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC); return &t }(),
			},
		},
		{
			name:  "excluded tag",
			input: "-nsfw",
			expect: search.Query{
				IncludedTags: []types.TagName{},
				ExcludedTags: []string{"nsfw"},
			},
		},
		{
			name:  "mixed input",
			input: "landscape,-nsfw,sort:random,tagged:true,type:image",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape"},
				ExcludedTags: []string{"nsfw"},
				Sort:         search.SortRandom,
				Tagged:       new(true),
				TypeImage:    true,
			},
		},
		{
			name:  "empty terms ignored",
			input: "landscape,,portrait,",
			expect: search.Query{
				IncludedTags: []types.TagName{"landscape", "portrait"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseSearch(tt.input)
			if got.Sort != tt.expect.Sort {
				t.Errorf("Sort = %q, want %q", got.Sort, tt.expect.Sort)
			}
			if got.Order != tt.expect.Order {
				t.Errorf("Order = %q, want %q", got.Order, tt.expect.Order)
			}
			if (got.CreatedAfter == nil) != (tt.expect.CreatedAfter == nil) {
				t.Errorf("CreatedAfter nil = %v, want nil = %v", got.CreatedAfter == nil, tt.expect.CreatedAfter == nil)
			} else if got.CreatedAfter != nil && !got.CreatedAfter.Equal(*tt.expect.CreatedAfter) {
				t.Errorf("CreatedAfter = %v, want %v", got.CreatedAfter, tt.expect.CreatedAfter)
			}
			if (got.CreatedBefore == nil) != (tt.expect.CreatedBefore == nil) {
				t.Errorf("CreatedBefore nil = %v, want nil = %v", got.CreatedBefore == nil, tt.expect.CreatedBefore == nil)
			} else if got.CreatedBefore != nil && !got.CreatedBefore.Equal(*tt.expect.CreatedBefore) {
				t.Errorf("CreatedBefore = %v, want %v", got.CreatedBefore, tt.expect.CreatedBefore)
			}
			if (got.Tagged == nil) != (tt.expect.Tagged == nil) || (got.Tagged != nil && *got.Tagged != *tt.expect.Tagged) {
				t.Errorf("Tagged = %v, want %v", got.Tagged, tt.expect.Tagged)
			}
			if got.TypeImage != tt.expect.TypeImage {
				t.Errorf("TypeImage = %v, want %v", got.TypeImage, tt.expect.TypeImage)
			}
			if got.TypeVideo != tt.expect.TypeVideo {
				t.Errorf("TypeVideo = %v, want %v", got.TypeVideo, tt.expect.TypeVideo)
			}
			if got.TypeAudio != tt.expect.TypeAudio {
				t.Errorf("TypeAudio = %v, want %v", got.TypeAudio, tt.expect.TypeAudio)
			}
			if len(got.IncludedTags) != len(tt.expect.IncludedTags) {
				t.Fatalf("len(Tags) = %d, want %d", len(got.IncludedTags), len(tt.expect.IncludedTags))
			}
			for i, tag := range got.IncludedTags {
				if tag != tt.expect.IncludedTags[i] {
					t.Errorf("Tags[%d] = %q, want %q", i, tag, tt.expect.IncludedTags[i])
				}
			}
			if len(got.ExcludedTags) != len(tt.expect.ExcludedTags) {
				t.Fatalf("len(ExcludeTags) = %d, want %d", len(got.ExcludedTags), len(tt.expect.ExcludedTags))
			}
			for i, tag := range got.ExcludedTags {
				if tag != tt.expect.ExcludedTags[i] {
					t.Errorf("ExcludeTags[%d] = %q, want %q", i, tag, tt.expect.ExcludedTags[i])
				}
			}
		})
	}
}

func TestMimeToExt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mime string
		ext  string
	}{
		{"image/webp", "webp"},
		{"image/jpeg", "jpg"},
		{"image/png", "png"},
		{"image/gif", "gif"},
		{"video/mp4", "mp4"},
		{"video/webm", "webm"},
		{"video/quicktime", "mov"},
		{"application/octet-stream", "bin"},
	}
	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			t.Parallel()
			if got := mimeToExt(tt.mime); got != tt.ext {
				t.Errorf("mimeToExt(%q) = %q, want %q", tt.mime, got, tt.ext)
			}
		})
	}
}

func TestEncodeDecodeRandomCursor(t *testing.T) {
	t.Parallel()
	original := randomCursor{Seed: 12345, Offset: 64}
	encoded := encodeRandomCursor(original)
	if encoded == "" {
		t.Fatal("encoded cursor should not be empty")
	}

	var decoded randomCursor
	if err := decodeRandomCursor(encoded, &decoded); err != nil {
		t.Fatalf("decodeRandomCursor() error = %v", err)
	}
	if decoded.Seed != original.Seed {
		t.Errorf("Seed = %d, want %d", decoded.Seed, original.Seed)
	}
	if decoded.Offset != original.Offset {
		t.Errorf("Offset = %d, want %d", decoded.Offset, original.Offset)
	}
}

func TestDecodeRandomCursorInvalid(t *testing.T) {
	t.Parallel()
	var rc randomCursor
	if err := decodeRandomCursor("not-valid-base64!!!", &rc); err == nil {
		t.Error("expected error for invalid base64")
	}
}

func FuzzParseSearch(f *testing.F) {
	seeds := []string{
		"",
		"landscape",
		"landscape,portrait",
		" landscape , portrait ",
		"sort:created",
		"sort:updated",
		"sort:random",
		"sort:invalid",
		"tagged:true",
		"tagged:false",
		"type:image",
		"type:video",
		"type:audio",
		"-nsfw",
		"landscape,-nsfw,sort:random,tagged:true,type:image",
		"landscape,,portrait,",
		"order:asc",
		"order:desc",
		"order:invalid",
		"created_after:2025-01-01T00:00:00Z",
		"created_before:2025-06-15T12:30:00Z",
		"sort:created,order:asc,created_after:2025-01-01T00:00:00Z",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		parseSearch(input)
	})
}
