package api

import (
	"net/http"
	"net/url"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  *int
		expect int
	}{
		{"nil defaults to MaxLimit", nil, MaxLimit},
		{"zero defaults to MaxLimit", new(0), MaxLimit},
		{"negative defaults to MaxLimit", new(-1), MaxLimit},
		{"within range", new(10), 10},
		{"at max", new(MaxLimit), MaxLimit},
		{"exceeds max capped", new(MaxLimit + 100), MaxLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsedLimit := parseLimit(tt.input)
			assert.Equal(t, tt.expect, parsedLimit)
		})
	}
}

func TestObfuscateCursorRoundTrip(t *testing.T) {
	t.Parallel()
	original := "test-cursor-value"
	encoded := obfuscateCursor(original)
	decoded, err := deobfuscateCursor(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestDeobfuscateCursorInvalid(t *testing.T) {
	t.Parallel()
	_, err := deobfuscateCursor("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestPaginate(t *testing.T) {
	t.Parallel()
	t.Run("more results available", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(11, 10, func() string { return "page-value" })
		assert.True(t, more)
		require.NotNil(t, cursor)
		decoded, err := deobfuscateCursor(*cursor)
		require.NoError(t, err)
		assert.Equal(t, "page-value", decoded)
	})

	t.Run("no more results", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(10, 10, func() string { return "unused" })
		assert.False(t, more)
		assert.Nil(t, cursor)
	})

	t.Run("fewer results than limit", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(5, 10, func() string { return "unused" })
		assert.False(t, more)
		assert.Nil(t, cursor)
	})
}

func traverseNamedPages(t *testing.T, handler http.Handler, endpoint string, cursor string, decode func([]byte) ([]string, *string)) []string {
	t.Helper()
	var collectedNames []string
	for page := 0; ; page++ {
		target := endpoint + "?limit=2"
		if cursor != "" {
			target += "&cursor=" + url.QueryEscape(cursor)
		}
		response := performRequest(handler, http.MethodGet, target, "", true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusOK, response.Code, "page %d status = %d; body = %q", page, response.Code, responseBody)
		responseBytes := response.Body.Bytes()
		names, next := decode(responseBytes)
		require.NotEmptyf(t, names, "page %d names", page)
		nameCount := len(names)
		require.LessOrEqualf(t, nameCount, 2, "page %d names", page)
		collectedNames = append(collectedNames, names...)
		if next == nil {
			break
		}
		cursor = *next
	}
	uniqueNames := slices.Compact(slices.Clone(collectedNames))
	nameCount := len(collectedNames)
	assert.Lenf(t, uniqueNames, nameCount, "traversal contains duplicates: %v", collectedNames)
	return collectedNames
}

func assertStaleNamedCursor(t *testing.T, handler http.Handler, endpoint, stale, want string, decodeFirst func([]byte) string) {
	t.Helper()
	response := performRequest(handler, http.MethodGet, endpoint+"?limit=1&cursor="+url.QueryEscape(obfuscateCursor(stale)), "", true)
	responseBody := response.Body.String()
	require.Equalf(t, http.StatusOK, response.Code, "stale cursor status = %d; body = %q", response.Code, responseBody)
	responseBytes := response.Body.Bytes()
	firstItem := decodeFirst(responseBytes)
	assert.Equalf(t, want, firstItem, "first item after stale cursor = %q, want %q", firstItem, want)
}
