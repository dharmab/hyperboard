package web

import (
	"fmt"
	"net/http"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/google/uuid"
)

func (a *app) handleNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		// Create new note
		resp, err := a.api.CreateNoteWithResponse(ctx, client.CreateNoteJSONRequestBody{
			Title: "New Note",
		})
		if err != nil || resp.StatusCode() >= 400 {
			notesResp, _ := a.api.GetNotesWithResponse(ctx)
			notes := []types.Note{}
			if notesResp != nil && notesResp.JSON200 != nil && notesResp.JSON200.Items != nil {
				notes = *notesResp.JSON200.Items
			}
			errMsg := "Failed to create note"
			if err != nil {
				errMsg = fmt.Sprintf("Failed to create note: %v", err)
			}
			a.renderTemplate(w, r, "notes", notesData{Notes: notes, Error: errMsg})
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/notes/%s", resp.JSON201.ID), http.StatusSeeOther)
		return
	}

	resp, err := a.api.GetNotesWithResponse(ctx)
	var loadErr string
	if err != nil {
		loadErr = fmt.Sprintf("Failed to load notes: %v", err)
	} else if resp.StatusCode() >= 400 {
		loadErr = fmt.Sprintf("Failed to load notes: %s", resp.Body)
	}
	notes := []types.Note{}
	if resp != nil && resp.JSON200 != nil && resp.JSON200.Items != nil {
		notes = *resp.JSON200.Items
	}
	a.renderTemplate(w, r, "notes", notesData{Notes: notes, Error: loadErr})
}

func (a *app) handleNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		noteID, err := uuid.Parse(id)
		if err != nil {
			a.renderTemplate(w, r, "note", noteData{Error: fmt.Sprintf("Invalid note ID: %v", err)})
			return
		}
		resp, err := a.api.GetNoteWithResponse(ctx, noteID)
		if err != nil || resp.JSON200 == nil {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Note not found: %v", err)
			} else {
				errMsg = fmt.Sprintf("Note not found: %s", resp.Body)
			}
			a.renderTemplate(w, r, "note", noteData{Error: errMsg})
			return
		}
		note := *resp.JSON200
		rendered := renderMarkdown(note.Content)
		isNew := note.Content == ""
		a.renderTemplate(w, r, "note", noteData{Note: note, RenderedContent: rendered, IsNew: isNew})

	case http.MethodPut:
		noteID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Invalid note ID", http.StatusBadRequest)
			return
		}
		resp, err := a.api.PutNoteWithResponse(ctx, noteID, client.PutNoteJSONRequestBody{
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		})
		if err != nil || resp.StatusCode() >= 400 {
			http.Error(w, "Failed to save note", http.StatusInternalServerError)
			return
		}
		// Return rendered markdown for HTMX swap
		rendered := renderMarkdown(r.FormValue("content"))
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<div id="note-view" class="note-content mt-2">%s</div>`, string(rendered))

	case http.MethodDelete:
		noteID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Invalid note ID", http.StatusBadRequest)
			return
		}
		resp, err := a.api.DeleteNoteWithResponse(ctx, noteID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete note: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete note: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
	}
}
