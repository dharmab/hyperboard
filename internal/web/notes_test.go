package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestHandleNotes_GET(t *testing.T) {
	t.Parallel()
	noteID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()
	notes := []types.Note{{ID: noteID, Title: "Test Note", CreatedAt: now, UpdatedAt: now}}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == notesAPIPath {
			jsonResponse(w, http.StatusOK, client.NotesResponse{Items: &notes})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Test Note") {
		t.Error("expected note title in response body")
	}
}

func TestHandleNotes_POST(t *testing.T) {
	t.Parallel()
	createdID := types.ID(uuid.Must(uuid.NewV4()))
	now := time.Now().UTC()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == notesAPIPath {
			jsonResponse(w, http.StatusCreated, types.Note{ID: createdID, Title: "New Note", CreatedAt: now, UpdatedAt: now})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusSeeOther, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, createdID.String()) {
		t.Errorf("redirect location = %q, want it to contain %s", loc, createdID)
	}
}

func TestHandleNotes_POST_APIError(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == notesAPIPath:
			jsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Too many notes"})
		case r.Method == http.MethodGet && r.URL.Path == notesAPIPath:
			jsonResponse(w, http.StatusOK, client.NotesResponse{})
		default:
			http.NotFound(w, r)
		}
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/notes", nil)
	w := httptest.NewRecorder()
	app.handleNotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (template render with error); body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Too many notes") {
		t.Errorf("expected error body to contain API message, got %q", w.Body.String())
	}
}

func TestHandleNote_PUT_APIError(t *testing.T) {
	t.Parallel()
	noteID := types.ID(uuid.Must(uuid.NewV4()))

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v1/notes/") {
			jsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Title is required"})
			return
		}
		http.NotFound(w, r)
	}))

	form := strings.NewReader("title=&content=hello")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/notes/"+noteID.String(), form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", noteID.String())
	w := httptest.NewRecorder()
	app.handleNote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Title is required") {
		t.Errorf("expected error body to contain API message, got %q", w.Body.String())
	}
}
