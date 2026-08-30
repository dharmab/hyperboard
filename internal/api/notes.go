package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/rs/zerolog"
)

const (
	maxNoteTitleLength   = 1024
	maxNoteContentLength = 4 << 20
)

// noteFromModel converts a database Note model to an API Note type.
func noteFromModel(model *models.Note) types.Note {
	return types.Note{
		ID:        types.ID(uuid.UUID(model.ID)),
		Title:     model.Title,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// GetNotes handles listing all notes.
func (s *Server) GetNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notes, err := s.sqlStore.ListNotes(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}

	// Stable sort in case timestamps are the same
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].CreatedAt.Equal(notes[j].CreatedAt) {
			return notes[i].ID.String() > notes[j].ID.String()
		}
		return notes[i].CreatedAt.After(notes[j].CreatedAt)
	})

	items := make([]types.Note, 0, len(notes))
	for _, note := range notes {
		items = append(items, noteFromModel(note))
	}

	resp := NotesResponse{
		Items: &items,
	}
	respond(w, http.StatusOK, resp)
}

// CreateNote handles creating a new note.
func (s *Server) CreateNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body CreateNoteJSONRequestBody
	if err := decodeNoteRequest(r, &body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateNote(body.Title, body.Content); err != nil {
		respondWithError(w, http.StatusBadRequest, "%s", err)
		return
	}

	model, err := s.sqlStore.CreateNote(ctx, body.Title, body.Content)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	zerolog.Ctx(ctx).Info().Stringer("note_id", model.ID).Msg("note created")
	respond(w, http.StatusCreated, noteFromModel(model))
}

// GetNote handles retrieving a single note by ID.
func (s *Server) GetNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	model, err := s.sqlStore.GetNote(ctx, uuid.UUID(id))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	respond(w, http.StatusOK, noteFromModel(model))
}

// PutNote handles updating an existing note.
func (s *Server) PutNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	var body PutNoteJSONRequestBody
	if err := decodeNoteRequest(r, &body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validateNote(body.Title, body.Content); err != nil {
		respondWithError(w, http.StatusBadRequest, "%s", err)
		return
	}

	model, err := s.sqlStore.UpdateNote(ctx, uuid.UUID(id), body.Title, body.Content)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update note")
		return
	}

	zerolog.Ctx(ctx).Info().Stringer("note_id", uuid.UUID(id)).Msg("note updated")
	respond(w, http.StatusOK, noteFromModel(model))
}

func decodeNoteRequest(r *http.Request, body any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(body); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body contains multiple JSON values")
		}
		return err
	}
	return nil
}

func validateNote(title, content string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if utf8.RuneCountInString(title) > maxNoteTitleLength {
		return errors.New("title is too long")
	}
	if utf8.RuneCountInString(content) > maxNoteContentLength {
		return errors.New("content is too long")
	}
	return nil
}

func (s *Server) DeleteNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	noteID := uuid.UUID(id)

	err := s.sqlStore.DeleteNote(ctx, noteID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	zerolog.Ctx(ctx).Info().Stringer("note_id", noteID).Msg("note deleted")
	w.WriteHeader(http.StatusNoContent)
}
