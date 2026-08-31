package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
)

func TestHandleNotes_GET(t *testing.T) {
	t.Parallel()
	noteID := types.ID(uuid.NewV4())
	now := time.Now().UTC()
	notes := []types.Note{{ID: noteID, Title: "Test Note", CreatedAt: now, UpdatedAt: now}}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/notes" {
			jsonResponse(w, http.StatusOK, client.NotesResponse{Items: &notes})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	body := w.Body.String()
	require.Equal(t, http.StatusOK, w.Code, "body = %s", body)
	assert.Contains(t, body, "Test Note")
}

func TestHandleNote_PUTRejectsOversizedForm(t *testing.T) {
	t.Parallel()

	apiCalled := false
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		http.Error(w, "unexpected API call", http.StatusInternalServerError)
	}))

	form := url.Values{
		"title":   {"Oversized note"},
		"content": {strings.Repeat("x", int(maxFormBody))},
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/notes/"+uuid.NewV4().String(), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", uuid.NewV4().String())
	w := httptest.NewRecorder()
	maxBody(maxFormBody, app.handleNote).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.False(t, apiCalled)
}

func TestHandleNotes_POST(t *testing.T) {
	t.Parallel()
	createdID := types.ID(uuid.NewV4())
	now := time.Now().UTC()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/notes" {
			jsonResponse(w, http.StatusCreated, types.Note{ID: createdID, Title: "New Note", CreatedAt: now, UpdatedAt: now})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	body := w.Body.String()
	location := w.Header().Get("Location")
	createdIDString := createdID.String()
	require.Equal(t, http.StatusSeeOther, w.Code, "body = %s", body)
	assert.Contains(t, location, createdIDString)
}
