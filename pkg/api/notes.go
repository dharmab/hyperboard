package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func noteFromModel(model *models.Note) types.Note {
	return types.Note{
		ID:        types.ID(model.ID),
		Title:     model.Title,
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func (s *Server) GetNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	notes, err := models.Notes.Query(
		sm.OrderBy(models.NoteColumns.CreatedAt).Desc(),
	).All(ctx, s.db)
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

func (s *Server) CreateNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body CreateNoteJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate note ID")
		return
	}

	now := new(time.Now().UTC())
	model, err := models.Notes.Insert(
		&models.NoteSetter{
			ID:        &id,
			Title:     &body.Title,
			Content:   &body.Content,
			CreatedAt: now,
			UpdatedAt: now,
		},
	).One(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	respond(w, http.StatusCreated, noteFromModel(model))
}

func (s *Server) GetNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	noteID := uuid.UUID(id)

	model, err := models.Notes.Query(
		sm.Where(models.NoteColumns.ID.EQ(psql.Arg(noteID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	respond(w, http.StatusOK, noteFromModel(model))
}

func (s *Server) PutNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	noteID := uuid.UUID(id)

	var body PutNoteJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	model, err := models.Notes.Query(
		sm.Where(models.NoteColumns.ID.EQ(psql.Arg(noteID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	err = model.Update(ctx, s.db, &models.NoteSetter{
		Title:     &body.Title,
		Content:   &body.Content,
		UpdatedAt: new(time.Now().UTC()),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update note")
		return
	}

	model, err = models.Notes.Query(
		sm.Where(models.NoteColumns.ID.EQ(psql.Arg(noteID))),
	).One(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated note")
		return
	}

	respond(w, http.StatusOK, noteFromModel(model))
}

func (s *Server) DeleteNote(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	noteID := uuid.UUID(id)

	_, err := models.Notes.Query(
		sm.Where(models.NoteColumns.ID.EQ(psql.Arg(noteID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve note")
		return
	}

	_, err = models.Notes.Delete(
		dm.Where(models.NoteColumns.ID.EQ(psql.Arg(noteID))),
	).Exec(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
