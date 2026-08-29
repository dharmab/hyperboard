package api

import (
	"net/http"
	"net/url"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"uuid"
)

func TestParameterBinding(t *testing.T) {
	t.Parallel()
	response := performRequest(
		newTestHandler(t),
		http.MethodGet,
		"/api/v1/posts?limit=not-an-integer",
		"",
		true,
	)

	responseBody := response.Body.String()
	require.Equalf(t, http.StatusBadRequest, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var responseError Error
	decodeJSON(t, response.Body.Bytes(), &responseError)
	assert.NotEmpty(t, responseError.Message)
}

func TestLimitBoundaries(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)

	for _, endpoint := range []string{"/api/v1/posts", "/api/v1/tags", "/api/v1/tagCategories"} {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()
			for _, limit := range []string{"-1", "0", "1", "64", "65"} {
				t.Run("limit "+limit, func(t *testing.T) {
					t.Parallel()
					response := performRequest(handler, http.MethodGet, endpoint+"?limit="+limit, "", true)
					responseBody := response.Body.String()
					require.Equalf(t, http.StatusOK, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusOK, responseBody)
					contentType := response.Header().Get("Content-Type")
					assertJSONContentType(t, contentType)
				})
			}
		})
	}
}

func TestRepeatedLimitIsRejected(t *testing.T) {
	t.Parallel()
	response := performRequest(newTestHandler(t), http.MethodGet, "/api/v1/posts?limit=1&limit=2", "", true)
	responseBody := response.Body.String()
	require.Equalf(t, http.StatusBadRequest, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var responseError Error
	decodeJSON(t, response.Body.Bytes(), &responseError)
	assert.NotEmpty(t, responseError.Message)
}

func TestMalformedCursors(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)
	for _, endpoint := range []string{"/api/v1/posts", "/api/v1/tags", "/api/v1/tagCategories"} {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()
			response := performRequest(handler, http.MethodGet, endpoint+"?cursor=%25%25%25", "", true)
			responseBody := response.Body.String()
			require.Equalf(t, http.StatusBadRequest, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, responseBody)
			contentType := response.Header().Get("Content-Type")
			assertJSONContentType(t, contentType)
		})
	}
}

func TestInvalidUUIDBinding(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)
	for _, tt := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/posts/not-a-uuid"},
		{http.MethodPut, "/api/v1/posts/not-a-uuid"},
		{http.MethodDelete, "/api/v1/posts/not-a-uuid"},
		{http.MethodGet, "/api/v1/posts/not-a-uuid/content"},
		{http.MethodPut, "/api/v1/posts/not-a-uuid/content"},
		{http.MethodGet, "/api/v1/posts/not-a-uuid/content/download"},
		{http.MethodGet, "/api/v1/posts/not-a-uuid/thumbnail"},
		{http.MethodPost, "/api/v1/posts/not-a-uuid/thumbnail"},
		{http.MethodPut, "/api/v1/posts/not-a-uuid/thumbnail"},
		{http.MethodGet, "/api/v1/posts/not-a-uuid/similar"},
		{http.MethodGet, "/api/v1/notes/not-a-uuid"},
		{http.MethodPut, "/api/v1/notes/not-a-uuid"},
		{http.MethodDelete, "/api/v1/notes/not-a-uuid"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()
			response := performRequest(handler, tt.method, tt.path, "", true)
			responseBody := response.Body.String()
			require.Equalf(t, http.StatusBadRequest, response.Code, "status = %d, want %d; body = %q", response.Code, http.StatusBadRequest, responseBody)
			assertJSONContentType(t, response.Header().Get("Content-Type"))
			var responseError Error
			decodeJSON(t, response.Body.Bytes(), &responseError)
			assert.NotEmpty(t, responseError.Message)
		})
	}
}

func TestPercentEncodedNames(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)
	suffix := uuid.NewV4().String()[:8]

	for _, tt := range []struct {
		collection string
		name       string
		body       func(string) string
	}{
		{"/api/v1/tags/", "tag " + suffix, func(name string) string { return `{"name":"` + name + `","description":"percent encoded"}` }},
		{"/api/v1/tagCategories/", "category " + suffix, func(name string) string {
			return `{"name":"` + name + `","description":"percent encoded","color":"#abcdef"}`
		}},
	} {
		t.Run(tt.collection, func(t *testing.T) {
			t.Parallel()
			path := tt.collection + url.PathEscape(tt.name)
			created := performRequest(handler, http.MethodPut, path, tt.body(tt.name), true)
			createdResponseBody := created.Body.String()
			require.Equalf(t, http.StatusCreated, created.Code, "create status = %d, want %d; body = %q", created.Code, http.StatusCreated, createdResponseBody)
			fetched := performRequest(handler, http.MethodGet, path, "", true)
			fetchedResponseBody := fetched.Body.String()
			require.Equalf(t, http.StatusOK, fetched.Code, "get status = %d, want %d; body = %q", fetched.Code, http.StatusOK, fetchedResponseBody)
			deleted := performRequest(handler, http.MethodDelete, path, "", true)
			deletedResponseBody := deleted.Body.String()
			assertEmptyNoContent(t, deleted.Code, deletedResponseBody)
		})
	}
}
