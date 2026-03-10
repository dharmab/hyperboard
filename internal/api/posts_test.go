package api

import (
	"testing"

	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/pkg/types"
)

func TestParseSearch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect search.PostSearch
	}{
		{
			name:  "empty string",
			input: "",
			expect: search.PostSearch{
				Tags: []types.TagName{},
			},
		},
		{
			name:  "single tag",
			input: "landscape",
			expect: search.PostSearch{
				Tags: []types.TagName{"landscape"},
			},
		},
		{
			name:  "multiple tags",
			input: "landscape,portrait",
			expect: search.PostSearch{
				Tags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "tags with whitespace",
			input: " landscape , portrait ",
			expect: search.PostSearch{
				Tags: []types.TagName{"landscape", "portrait"},
			},
		},
		{
			name:  "sort created",
			input: "sort:created",
			expect: search.PostSearch{
				Tags: []types.TagName{},
				Sort: search.SortCreatedAt,
			},
		},
		{
			name:  "sort updated",
			input: "sort:updated",
			expect: search.PostSearch{
				Tags: []types.TagName{},
				Sort: search.SortUpdatedAt,
			},
		},
		{
			name:  "sort random",
			input: "sort:random",
			expect: search.PostSearch{
				Tags: []types.TagName{},
				Sort: search.SortRandom,
			},
		},
		{
			name:  "invalid sort ignored",
			input: "sort:invalid",
			expect: search.PostSearch{
				Tags: []types.TagName{},
			},
		},
		{
			name:  "tagged true",
			input: "tagged:true",
			expect: search.PostSearch{
				Tags:   []types.TagName{},
				Tagged: search.TaggedFilterTrue,
			},
		},
		{
			name:  "tagged false",
			input: "tagged:false",
			expect: search.PostSearch{
				Tags:   []types.TagName{},
				Tagged: search.TaggedFilterFalse,
			},
		},
		{
			name:  "type image",
			input: "type:image",
			expect: search.PostSearch{
				Tags:      []types.TagName{},
				TypeImage: true,
			},
		},
		{
			name:  "type video",
			input: "type:video",
			expect: search.PostSearch{
				Tags:      []types.TagName{},
				TypeVideo: true,
			},
		},
		{
			name:  "type audio",
			input: "type:audio",
			expect: search.PostSearch{
				Tags:      []types.TagName{},
				TypeAudio: true,
			},
		},
		{
			name:  "excluded tag",
			input: "-nsfw",
			expect: search.PostSearch{
				Tags:        []types.TagName{},
				ExcludeTags: []string{"nsfw"},
			},
		},
		{
			name:  "mixed input",
			input: "landscape,-nsfw,sort:random,tagged:true,type:image",
			expect: search.PostSearch{
				Tags:        []types.TagName{"landscape"},
				ExcludeTags: []string{"nsfw"},
				Sort:        search.SortRandom,
				Tagged:      search.TaggedFilterTrue,
				TypeImage:   true,
			},
		},
		{
			name:  "empty terms ignored",
			input: "landscape,,portrait,",
			expect: search.PostSearch{
				Tags: []types.TagName{"landscape", "portrait"},
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
			if got.Tagged != tt.expect.Tagged {
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
			if len(got.Tags) != len(tt.expect.Tags) {
				t.Fatalf("len(Tags) = %d, want %d", len(got.Tags), len(tt.expect.Tags))
			}
			for i, tag := range got.Tags {
				if tag != tt.expect.Tags[i] {
					t.Errorf("Tags[%d] = %q, want %q", i, tag, tt.expect.Tags[i])
				}
			}
			if len(got.ExcludeTags) != len(tt.expect.ExcludeTags) {
				t.Fatalf("len(ExcludeTags) = %d, want %d", len(got.ExcludeTags), len(tt.expect.ExcludeTags))
			}
			for i, tag := range got.ExcludeTags {
				if tag != tt.expect.ExcludeTags[i] {
					t.Errorf("ExcludeTags[%d] = %q, want %q", i, tag, tt.expect.ExcludeTags[i])
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
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		parseSearch(input)
	})
}
