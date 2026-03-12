package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
)

// noteFromModel converts a database Note model to an API Note type.
func noteFromModel(model *models.Note) types.Note {
	return types.Note{
		ID:        types.ID(model.ID),
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
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
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
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Title is required")
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
