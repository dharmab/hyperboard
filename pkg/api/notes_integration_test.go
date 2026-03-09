package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
)

func TestNotesIntegration(t *testing.T) {
	srv := newTestServer(t)

	var noteID types.ID

	t.Run("create note", func(t *testing.T) {
		body := CreateNoteJSONBody{
			Title:   "Test Note",
			Content: "This is test content.",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.CreateNote(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("CreateNote status = %d, want %d; body = %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var note types.Note
		if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if note.Title != body.Title {
			t.Errorf("Title = %q, want %q", note.Title, body.Title)
		}
		if note.Content != body.Content {
			t.Errorf("Content = %q, want %q", note.Content, body.Content)
		}
		noteID = note.ID
	})

	t.Run("get note", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+uuid.UUID(noteID).String(), nil)
		w := httptest.NewRecorder()
		srv.GetNote(w, req, noteID)

		if w.Code != http.StatusOK {
			t.Fatalf("GetNote status = %d, want %d", w.Code, http.StatusOK)
		}

		var note types.Note
		if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if note.ID != noteID {
			t.Errorf("ID = %v, want %v", note.ID, noteID)
		}
	})

	t.Run("list notes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
		w := httptest.NewRecorder()
		srv.GetNotes(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("GetNotes status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp NotesResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Items == nil || len(*resp.Items) == 0 {
			t.Error("expected at least one note in listing")
		}
	})

	t.Run("update note", func(t *testing.T) {
		body := PutNoteJSONBody{
			Title:   "Updated Note",
			Content: "Updated content.",
		}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/notes/"+uuid.UUID(noteID).String(), bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.PutNote(w, req, noteID)

		if w.Code != http.StatusOK {
			t.Fatalf("PutNote status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
		}

		var note types.Note
		if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if note.Title != body.Title {
			t.Errorf("Title = %q, want %q", note.Title, body.Title)
		}
	})

	t.Run("delete note", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/notes/"+uuid.UUID(noteID).String(), nil)
		w := httptest.NewRecorder()
		srv.DeleteNote(w, req, noteID)

		if w.Code != http.StatusNoContent {
			t.Fatalf("DeleteNote status = %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("get deleted note returns not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+uuid.UUID(noteID).String(), nil)
		w := httptest.NewRecorder()
		srv.GetNote(w, req, noteID)

		if w.Code != http.StatusNotFound {
			t.Fatalf("GetNote status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("get nonexistent note returns not found", func(t *testing.T) {
		fakeID := types.ID(uuid.Must(uuid.NewV4()))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/"+uuid.UUID(fakeID).String(), nil)
		w := httptest.NewRecorder()
		srv.GetNote(w, req, fakeID)

		if w.Code != http.StatusNotFound {
			t.Fatalf("GetNote status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
