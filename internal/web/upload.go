package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/rs/zerolog/log"
)

// uploadConflict records a conflict response from a single file upload.
type uploadConflict struct {
	Filename string
	Body     []byte
}

// handleUpload serves the upload page and handles multipart file uploads.
func (a *app) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isXHR := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	if r.Method == http.MethodGet {
		a.renderTemplate(w, r, "upload", nil)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if isXHR {
			a.respondJSON(w, http.StatusBadRequest, map[string]any{"error": "Failed to parse form"})
		} else {
			a.renderTemplate(w, r, "upload", map[string]any{"Error": "Failed to parse form"})
		}
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		if isXHR {
			a.respondJSON(w, http.StatusBadRequest, map[string]any{"error": "No files provided"})
		} else {
			a.renderTemplate(w, r, "upload", map[string]any{"Error": "No files provided"})
		}
		return
	}

	var lastPostID types.ID
	var errors []string
	var conflicts []uploadConflict
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, header.Filename+": failed to open")
			continue
		}

		data, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil {
			errors = append(errors, header.Filename+": failed to read")
			continue
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		force := r.FormValue("force") == "true"
		forceParam := &client.UploadPostParams{}
		if force {
			forceParam.Force = &force
		}

		resp, err := a.api.UploadPostWithBodyWithResponse(ctx, forceParam, contentType, bytes.NewReader(data))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", header.Filename, err))
			continue
		}
		if resp.StatusCode() == http.StatusConflict {
			conflicts = append(conflicts, uploadConflict{Filename: header.Filename, Body: resp.Body})
			continue
		}
		if resp.StatusCode() >= 400 {
			var apiErr struct {
				Message string `json:"message"`
			}
			if json.Unmarshal(resp.Body, &apiErr) == nil && apiErr.Message != "" {
				errors = append(errors, fmt.Sprintf("%s: %s", header.Filename, apiErr.Message))
			} else {
				errors = append(errors, fmt.Sprintf("%s: HTTP %d", header.Filename, resp.StatusCode()))
			}
			continue
		}
		if resp.JSON201 != nil {
			lastPostID = resp.JSON201.ID
			log.Info().Stringer("id", resp.JSON201.ID).Str("filename", header.Filename).Msg("uploaded post")
		}
	}

	// Convert conflicts to errors for non-XHR responses
	for _, c := range conflicts {
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(c.Body, &apiErr) == nil && apiErr.Message != "" {
			errors = append(errors, fmt.Sprintf("%s: %s", c.Filename, apiErr.Message))
		} else {
			errors = append(errors, c.Filename+": duplicate detected")
		}
	}

	if isXHR {
		// XHR uploads send one file per request, so there is at most one conflict.
		// Return its body so the frontend can display similar-post suggestions.
		if len(conflicts) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write(conflicts[0].Body)
			return
		}
		if len(errors) == len(files) {
			a.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": strings.Join(errors, "; ")})
		} else if len(errors) > 0 {
			a.respondJSON(w, http.StatusOK, map[string]any{
				"id":    lastPostID,
				"error": strings.Join(errors, "; "),
			})
		} else {
			a.respondJSON(w, http.StatusOK, map[string]any{"id": lastPostID})
		}
		return
	}

	totalErrors := len(errors)
	if totalErrors == len(files) {
		a.renderTemplate(w, r, "upload", map[string]any{"Error": strings.Join(errors, "; ")})
		return
	}

	if totalErrors > 0 {
		a.renderTemplate(w, r, "upload", map[string]any{"Error": "Some uploads failed: " + strings.Join(errors, "; ")})
		return
	}

	if len(files) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/posts/%s", lastPostID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (a *app) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data) //nolint:errchkjson // encoding response to HTTP writer, error is not actionable
}
