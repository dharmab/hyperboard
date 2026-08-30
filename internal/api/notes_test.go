package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateNote(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))

	response := performRequest(handler, http.MethodPost, "/api/v1/notes", `{"title":"note","content":"content"}`, true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var note types.Note
	decodeJSON(t, response.Body.Bytes(), &note)
	assert.NotEqual(t, (types.ID{}), note.ID)
	assert.Equal(t, "note", note.Title)
	assert.Equal(t, "content", note.Content)
	assert.False(t, note.CreatedAt.IsZero())
	assert.False(t, note.UpdatedAt.IsZero())
}

func TestGetNote(t *testing.T) {
	t.Parallel()
	handler, note := createNoteFixture(t)

	response := performRequest(handler, http.MethodGet, notePath(note), "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var fetched types.Note
	decodeJSON(t, response.Body.Bytes(), &fetched)
	assert.Equal(t, note, fetched)
}

func TestListNotes(t *testing.T) {
	t.Parallel()
	handler, note := createNoteFixture(t)

	response := performRequest(handler, http.MethodGet, "/api/v1/notes", "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var notes NotesResponse
	decodeJSON(t, response.Body.Bytes(), &notes)
	require.NotNil(t, notes.Items)
	require.Len(t, *notes.Items, 1)
	assert.Equal(t, note, (*notes.Items)[0])
}

func TestUpdateNote(t *testing.T) {
	t.Parallel()
	handler, note := createNoteFixture(t)

	response := performRequest(handler, http.MethodPut, notePath(note), `{"title":"updated","content":"updated content"}`, true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var updated types.Note
	decodeJSON(t, response.Body.Bytes(), &updated)
	assert.Equal(t, note.ID, updated.ID)
	assert.Equal(t, "updated", updated.Title)
	assert.Equal(t, "updated content", updated.Content)
	assert.Equal(t, note.CreatedAt, updated.CreatedAt)
	assert.False(t, updated.UpdatedAt.Before(note.UpdatedAt))
}

func TestDeleteNote(t *testing.T) {
	t.Parallel()
	handler, note := createNoteFixture(t)

	response := performRequest(handler, http.MethodDelete, notePath(note), "", true)
	assertEmptyNoContent(t, response.Code, response.Body.String())

	missingResponse := performRequest(handler, http.MethodGet, notePath(note), "", true)
	assert.Equal(t, http.StatusNotFound, missingResponse.Code, "body: %s", missingResponse.Body.String())
}

func createNoteFixture(t *testing.T) (http.Handler, types.Note) {
	t.Helper()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	response := performRequest(handler, http.MethodPost, "/api/v1/notes", `{"title":"note","content":"content"}`, true)
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", response.Body.String())
	var note types.Note
	decodeJSON(t, response.Body.Bytes(), &note)
	return handler, note
}

func notePath(note types.Note) string {
	return "/api/v1/notes/" + uuid.UUID(note.ID).String()
}

func TestNotesEmptyCollectionAndExactOrdering(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))

	empty := performRequest(handler, http.MethodGet, "/api/v1/notes", "", true)
	emptyResponseBody := empty.Body.String()
	require.Equalf(t, http.StatusOK, empty.Code, "empty list status = %d; body = %q", empty.Code, emptyResponseBody)
	var emptyBody NotesResponse
	emptyResponseBytes := empty.Body.Bytes()
	decodeJSON(t, emptyResponseBytes, &emptyBody)
	require.NotNil(t, emptyBody.Items)
	assert.Empty(t, *emptyBody.Items)
	trimmedEmptyResponseBody := strings.TrimSpace(emptyResponseBody)
	assert.JSONEq(t, `{"items":[]}`, trimmedEmptyResponseBody)

	var want []types.ID
	for _, title := range []string{"first", "second", "third"} {
		response := performRequest(handler, http.MethodPost, "/api/v1/notes", noteJSON(t, title, title+" content"), true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusCreated, response.Code, "create %q status = %d; body = %q", title, response.Code, responseBody)
		var note types.Note
		decodeJSON(t, response.Body.Bytes(), &note)
		want = append([]types.ID{note.ID}, want...)
		time.Sleep(time.Millisecond)
	}

	response := performRequest(handler, http.MethodGet, "/api/v1/notes", "", true)
	var listed NotesResponse
	decodeJSON(t, response.Body.Bytes(), &listed)
	listedIDs := make([]types.ID, len(*listed.Items))
	for i, note := range *listed.Items {
		listedIDs[i] = note.ID
	}
	assert.Equalf(t, want, listedIDs, "listed IDs = %v, want newest-first %v", listedIDs, want)
}

func TestNoteInputEdges(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))

	for _, tc := range []struct {
		name   string
		body   string
		status int
	}{
		{name: "whitespace-only title", body: noteJSON(t, " \t\n ", "content"), status: http.StatusBadRequest},
		{name: "maximum title", body: noteJSON(t, strings.Repeat("t", maxNoteTitleLength), "content"), status: http.StatusCreated},
		{name: "title too long", body: noteJSON(t, strings.Repeat("t", maxNoteTitleLength+1), "content"), status: http.StatusBadRequest},
		{name: "maximum content", body: noteJSON(t, "maximum content", strings.Repeat("c", maxNoteContentLength)), status: http.StatusCreated},
		{name: "content too long", body: noteJSON(t, "content too long", strings.Repeat("c", maxNoteContentLength+1)), status: http.StatusBadRequest},
		{name: "trailing JSON", body: `{"title":"valid","content":"valid"} {}`, status: http.StatusBadRequest},
	} {
		response := performRequest(handler, http.MethodPost, "/api/v1/notes", tc.body, true)
		responseBody := response.Body.String()
		assert.Equalf(t, tc.status, response.Code, "%s status = %d, want %d; body = %q", tc.name, response.Code, tc.status, responseBody)
	}

	created := performRequest(handler, http.MethodPost, "/api/v1/notes", noteJSON(t, "update target", "content"), true)
	createdResponseBody := created.Body.String()
	require.Equalf(t, http.StatusCreated, created.Code, "create update target status = %d; body = %q", created.Code, createdResponseBody)
	var note types.Note
	decodeJSON(t, created.Body.Bytes(), &note)
	path := "/api/v1/notes/" + uuid.UUID(note.ID).String()
	for i, body := range []string{
		noteJSON(t, "  ", "content"),
		noteJSON(t, strings.Repeat("t", maxNoteTitleLength+1), "content"),
		noteJSON(t, "content too long", strings.Repeat("c", maxNoteContentLength+1)),
		`{"title":"valid","content":"valid"} []`,
	} {
		response := performRequest(handler, http.MethodPut, path, body, true)
		responseBody := response.Body.String()
		assert.Equalf(t, http.StatusBadRequest, response.Code, "invalid update %d status = %d, want %d; body = %q", i, response.Code, http.StatusBadRequest, responseBody)
	}
}

func TestNoteUpdateDeleteNonexistent(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	path := "/api/v1/notes/" + uuid.NewV4().String()

	update := performRequest(handler, http.MethodPut, path, noteJSON(t, "missing", "missing"), true)
	updateResponseBody := update.Body.String()
	assert.Equalf(t, http.StatusNotFound, update.Code, "update nonexistent status = %d, want %d; body = %q", update.Code, http.StatusNotFound, updateResponseBody)
	deleteResponse := performRequest(handler, http.MethodDelete, path, "", true)
	deleteResponseBody := deleteResponse.Body.String()
	assert.Equalf(t, http.StatusNotFound, deleteResponse.Code, "delete nonexistent status = %d, want %d; body = %q", deleteResponse.Code, http.StatusNotFound, deleteResponseBody)
}

func noteJSON(t *testing.T, title, content string) string {
	t.Helper()
	body, err := json.Marshal(map[string]string{"title": title, "content": content})
	require.NoError(t, err)
	return string(body)
}
