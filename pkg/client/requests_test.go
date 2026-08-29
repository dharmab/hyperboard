package client

import (
	"encoding/json"
	"testing"
	"time"
	"uuid"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateRequestSerializationExcludesReadOnlyFields(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	id := types.ID(uuid.NewV4())
	tags := []types.TagName{"first", "second"}
	colors := map[string]string{"first": "#123456"}
	aliases := []string{"alias"}
	cascades := []types.TagName{"parent"}
	category := "category"
	count := 42

	tests := map[string]struct {
		request any
		want    map[string]any
	}{
		"post": {
			request: NewPostUpdateRequest(types.Post{
				ID:            id,
				MimeType:      "image/png",
				ContentUrl:    "/content",
				ThumbnailUrl:  "/thumbnail",
				Note:          "note",
				HasAudio:      true,
				Tags:          tags,
				TagColors:     &colors,
				CascadingTags: &cascades,
				CreatedAt:     now,
				UpdatedAt:     now,
			}),
			want: map[string]any{
				"id":   uuid.UUID(id).String(),
				"note": "note",
				"tags": []any{"first", "second"},
			},
		},
		"tag": {
			request: NewTagUpdateRequest(types.Tag{
				Name:          "tag",
				Category:      &category,
				Description:   "description",
				Aliases:       &aliases,
				CascadingTags: &cascades,
				PostCount:     &count,
				CreatedAt:     now,
				UpdatedAt:     now,
			}),
			want: map[string]any{
				"name":          "tag",
				"category":      "category",
				"description":   "description",
				"aliases":       []any{"alias"},
				"cascadingTags": []any{"parent"},
			},
		},
		"tag category": {
			request: NewTagCategoryUpdateRequest(types.TagCategory{
				Name:        "category",
				Description: "description",
				Color:       "#123456",
				TagCount:    &count,
				CreatedAt:   now,
				UpdatedAt:   now,
			}),
			want: map[string]any{
				"name":        "category",
				"description": "description",
				"color":       "#123456",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			encoded, err := json.Marshal(test.request)
			require.NoError(t, err)

			var got map[string]any
			require.NoError(t, json.Unmarshal(encoded, &got))
			require.Equal(t, test.want, got)
		})
	}
}
