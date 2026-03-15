package api

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/search"
	"github.com/dharmab/hyperboard/pkg/types"
)

// parseSearch parses a comma-separated search query string into a Query.
func parseSearch(query string) search.Query {
	postSearch := search.Query{
		IncludedTags: []types.TagName{},
	}

	if query == "" {
		return postSearch
	}

	// Split search string by commas and trim whitespace from each term
	for part := range strings.SplitSeq(query, ",") {
		term := strings.TrimSpace(part)
		if term == "" {
			continue
		}
		if term == string(search.SortRandom) {
			postSearch.Sort = search.SortRandom
		} else if term == string(search.SortCreatedAt) {
			postSearch.Sort = search.SortCreatedAt
		} else if term == string(search.SortUpdatedAt) {
			postSearch.Sort = search.SortUpdatedAt
		} else if strings.HasPrefix(term, "sort:") {
			// Ignore unknown sort values
		} else if term == search.TagTaggedTrue {
			postSearch.Tagged = new(true)
		} else if term == search.TagTaggedFalse {
			postSearch.Tagged = new(false)
		} else if term == string(search.OrderAsc) {
			postSearch.Order = search.OrderAsc
		} else if term == string(search.OrderDesc) {
			postSearch.Order = search.OrderDesc
		} else if strings.HasPrefix(term, "order:") {
			// Ignore unknown order values
		} else if after, ok := strings.CutPrefix(term, "created_after:"); ok {
			if ts, err := time.Parse(time.RFC3339Nano, after); err == nil {
				postSearch.CreatedAfter = &ts
			} else if ts, err := time.Parse(time.RFC3339, after); err == nil {
				postSearch.CreatedAfter = &ts
			}
		} else if before, ok := strings.CutPrefix(term, "created_before:"); ok {
			if ts, err := time.Parse(time.RFC3339Nano, before); err == nil {
				postSearch.CreatedBefore = &ts
			} else if ts, err := time.Parse(time.RFC3339, before); err == nil {
				postSearch.CreatedBefore = &ts
			}
		} else if term == search.TagImage {
			postSearch.TypeImage = true
		} else if term == search.TagVideo {
			postSearch.TypeVideo = true
		} else if term == search.TagAudio {
			postSearch.TypeAudio = true
		} else if excluded, ok := strings.CutPrefix(term, "-"); ok && excluded != "" {
			postSearch.ExcludedTags = append(postSearch.ExcludedTags, excluded)
		} else {
			postSearch.IncludedTags = append(postSearch.IncludedTags, term)
		}
	}

	return postSearch
}

// postCursor holds cursor state for deterministic post pagination.
type postCursor struct {
	Timestamp string `json:"t"`
	ID        string `json:"id"`
}

// encodePostCursor serializes a postCursor to a base64-encoded string.
func encodePostCursor(pc postCursor) string {
	//nolint:errchkjson // postCursor contains only string fields, json.Marshal cannot fail
	data, _ := json.Marshal(pc)
	return base64.URLEncoding.EncodeToString(data)
}

// decodePostCursor deserializes a base64-encoded string into a postCursor.
func decodePostCursor(s string) (postCursor, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return postCursor{}, err
	}
	var pc postCursor
	return pc, json.Unmarshal(data, &pc)
}

// randomCursor holds cursor state for random-order post pagination.
type randomCursor struct {
	Seed   int64 `json:"seed"`
	Offset int   `json:"offset"`
}

// encodeRandomCursor serializes a randomCursor to a base64-encoded string.
func encodeRandomCursor(rc randomCursor) string {
	//nolint:errchkjson // randomCursor contains only primitive fields, json.Marshal cannot fail
	data, _ := json.Marshal(rc)
	return base64.URLEncoding.EncodeToString(data)
}

// decodeRandomCursor deserializes a base64-encoded string into a randomCursor.
func decodeRandomCursor(s string, rc *randomCursor) error {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rc)
}
