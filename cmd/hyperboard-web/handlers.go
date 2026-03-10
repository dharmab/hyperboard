package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const newResourceName = "_new"

func (app *App) handlePosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")
	cursor := r.URL.Query().Get("cursor")

	limit := 24
	params := &client.GetPostsParams{Limit: &limit}
	if search != "" {
		params.Search = &search
	}
	if cursor != "" {
		params.Cursor = &cursor
	}

	resp, err := app.api.GetPostsWithResponse(ctx, params)
	var loadErr string
	if err != nil {
		log.Error().Err(err).Str("search", search).Str("cursor", cursor).Msg("Failed to load posts")
		loadErr = fmt.Sprintf("Failed to load posts: %v", err)
	} else if resp.StatusCode() >= 400 {
		loadErr = fmt.Sprintf("Failed to load posts: %s", resp.Body)
	}

	posts := []types.Post{}
	nextCursor := ""
	if resp != nil && resp.JSON200 != nil {
		if resp.JSON200.Items != nil {
			posts = *resp.JSON200.Items
		}
		if resp.JSON200.Cursor != nil {
			nextCursor = *resp.JSON200.Cursor
		}
	}

	data := PostsData{
		Posts:      posts,
		NextCursor: nextCursor,
		Search:     search,
		TagFilters: parseTagFilters(app.cfg.TagFilters),
		Error:      loadErr,
	}

	// HTMX partial request (search or infinite scroll)
	if r.Header.Get("HX-Request") == "true" {
		if loadErr != "" {
			http.Error(w, loadErr, http.StatusInternalServerError)
			return
		}
		app.renderTemplate(w, r, "posts-items", data)
		return
	}

	app.renderTemplate(w, r, "posts", data)
}

func (app *App) handlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		postID, err := uuid.Parse(id)
		if err != nil {
			app.renderTemplate(w, r, "post", PostData{Error: fmt.Sprintf("Invalid post ID: %v", err)})
			return
		}
		resp, err := app.api.GetPostWithResponse(ctx, postID)
		if err != nil || resp.JSON200 == nil {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Post not found: %v", err)
			} else {
				errMsg = fmt.Sprintf("Post not found: %s", resp.Body)
			}
			app.renderTemplate(w, r, "post", PostData{Error: errMsg})
			return
		}
		post := *resp.JSON200

		var fileSize int64
		if headResp, err := app.media.head(ctx, "/media"+mediaPath(post.ContentUrl)); err == nil {
			_ = headResp.Body.Close()
			fileSize = headResp.ContentLength
		}

		var similarPosts []types.Post
		similarLimit := 12
		if similarResp, err := app.api.GetSimilarPostsWithResponse(ctx, postID, &client.GetSimilarPostsParams{Limit: &similarLimit}); err == nil && similarResp.JSON200 != nil {
			if similarResp.JSON200.Items != nil {
				similarPosts = *similarResp.JSON200.Items
			}
		}

		app.renderTemplate(w, r, "post", PostData{
			Post:         post,
			IsVideo:      strings.HasPrefix(post.MimeType, "video/"),
			FileSize:     fileSize,
			SimilarPosts: similarPosts,
		})

	case http.MethodDelete:
		postID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid post ID: %v", err), http.StatusBadRequest)
			return
		}
		resp, err := app.api.DeletePostWithResponse(ctx, postID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete post: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete post: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *App) handlePostNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	note := r.FormValue("note")

	postID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	resp, err := app.api.GetPostWithResponse(ctx, postID)
	if err != nil || resp.JSON200 == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	post := *resp.JSON200
	post.Note = note
	putResp, err := app.api.PutPostWithResponse(ctx, postID, post)
	if err != nil || putResp.StatusCode() >= 400 {
		http.Error(w, "Failed to save note", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *App) handlePostTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	postID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	resp, err := app.api.GetPostWithResponse(ctx, postID)
	if err != nil || resp.JSON200 == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	post := *resp.JSON200

	switch r.Method {
	case http.MethodPost:
		tagName := r.FormValue("q")
		if tagName != "" {
			post.Tags = append(post.Tags, tagName)
			putResp, err := app.api.PutPostWithResponse(ctx, postID, post)
			if err != nil || putResp.StatusCode() >= 400 {
				http.Error(w, "Failed to add tag", http.StatusInternalServerError)
				return
			}
		}
	case http.MethodDelete:
		tagToRemove := r.PathValue("tag")
		newTags := []types.TagName{}
		for _, t := range post.Tags {
			if t != tagToRemove {
				newTags = append(newTags, t)
			}
		}
		post.Tags = newTags
		putResp, err := app.api.PutPostWithResponse(ctx, postID, post)
		if err != nil || putResp.StatusCode() >= 400 {
			http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
			return
		}
	}

	// Re-fetch to get updated tags
	reResp, err := app.api.GetPostWithResponse(ctx, postID)
	if err != nil || reResp.JSON200 == nil {
		http.Error(w, "Failed to reload post", http.StatusInternalServerError)
		return
	}
	app.renderTemplate(w, r, "post-tags", PostData{Post: *reResp.JSON200})
}

func (app *App) handleTagSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	postIDStr := r.URL.Query().Get("post")
	exclude := r.URL.Query().Get("exclude")

	// Load existing post tags to exclude from suggestions
	excludeTags := map[string]bool{}
	if postIDStr != "" {
		if postID, err := uuid.Parse(postIDStr); err == nil {
			if resp, err := app.api.GetPostWithResponse(ctx, postID); err == nil && resp.JSON200 != nil {
				for _, t := range resp.JSON200.Tags {
					excludeTags[t] = true
				}
			}
		}
	}
	if exclude != "" {
		excludeTags[exclude] = true
	}

	limit := 1000
	params := &client.GetTagsParams{Limit: &limit}
	resp, err := app.api.GetTagsWithResponse(ctx, params)
	if err != nil || resp.JSON200 == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if resp.JSON200.Items != nil {
		for _, tag := range *resp.JSON200.Items {
			if excludeTags[tag.Name] {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(tag.Name), strings.ToLower(q)) {
				continue
			}
			_, _ = fmt.Fprintf(w, "<option value=%q>", tag.Name)
		}
	}
}

// uploadConflict records a conflict response from a single file upload.
type uploadConflict struct {
	Filename string
	Body     []byte
}

func (app *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isXHR := r.Header.Get("X-Requested-With") == "XMLHttpRequest"

	if r.Method == http.MethodGet {
		app.renderTemplate(w, r, "upload", nil)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if isXHR {
			app.respondJSON(w, http.StatusBadRequest, map[string]any{"error": "Failed to parse form"})
		} else {
			app.renderTemplate(w, r, "upload", map[string]any{"Error": "Failed to parse form"})
		}
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		if isXHR {
			app.respondJSON(w, http.StatusBadRequest, map[string]any{"error": "No files provided"})
		} else {
			app.renderTemplate(w, r, "upload", map[string]any{"Error": "No files provided"})
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

		resp, err := app.api.UploadPostWithBodyWithResponse(ctx, forceParam, contentType, bytes.NewReader(data))
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
			app.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": strings.Join(errors, "; ")})
		} else if len(errors) > 0 {
			app.respondJSON(w, http.StatusOK, map[string]any{
				"id":    lastPostID,
				"error": strings.Join(errors, "; "),
			})
		} else {
			app.respondJSON(w, http.StatusOK, map[string]any{"id": lastPostID})
		}
		return
	}

	totalErrors := len(errors)
	if totalErrors == len(files) {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": strings.Join(errors, "; ")})
		return
	}

	if totalErrors > 0 {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": "Some uploads failed: " + strings.Join(errors, "; ")})
		return
	}

	if len(files) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/posts/%s", lastPostID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *App) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data) //nolint:errchkjson // encoding response to HTTP writer, error is not actionable
}

func (app *App) handleTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all tags (paginate through all pages)
	var errs []string
	allTags := []types.Tag{}
	var cursor *string
	for {
		limit := 1000
		params := &client.GetTagsParams{Limit: &limit, Cursor: cursor}
		resp, err := app.api.GetTagsWithResponse(ctx, params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %v", err))
			break
		}
		if resp.StatusCode() >= 400 {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %s", resp.Body))
			break
		}
		if resp.JSON200 != nil && resp.JSON200.Items != nil {
			allTags = append(allTags, *resp.JSON200.Items...)
		}
		if resp.JSON200 == nil || resp.JSON200.Cursor == nil || *resp.JSON200.Cursor == "" {
			break
		}
		cursor = resp.JSON200.Cursor
	}

	// Fetch categories for color map
	catLimit := 1000
	catResp, err := app.api.GetTagCategoriesWithResponse(ctx, &client.GetTagCategoriesParams{Limit: &catLimit})
	if err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}
	colorMap := map[string]string{}
	if catResp != nil && catResp.JSON200 != nil && catResp.JSON200.Items != nil {
		for _, c := range *catResp.JSON200.Items {
			colorMap[c.Name] = c.Color
		}
	}

	app.renderTemplate(w, r, "tags", TagsData{Tags: allTags, CategoryColors: colorMap, Error: strings.Join(errs, "; ")})
}

func (app *App) handleTagEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == newResourceName

	// Fetch categories for dropdown
	catLimit := 1000
	catResp, err := app.api.GetTagCategoriesWithResponse(ctx, &client.GetTagCategoriesParams{Limit: &catLimit})
	var catErr string
	if err != nil {
		catErr = fmt.Sprintf("Failed to load categories: %v", err)
	}
	cats := []types.TagCategory{}
	if catResp != nil && catResp.JSON200 != nil && catResp.JSON200.Items != nil {
		cats = *catResp.JSON200.Items
	}

	switch r.Method {
	case http.MethodGet:
		tag := types.Tag{}
		var errs []string
		if catErr != "" {
			errs = append(errs, catErr)
		}
		if !isNew {
			resp, err := app.api.GetTagWithResponse(ctx, name)
			if err != nil {
				errs = append(errs, fmt.Sprintf("Failed to load tag: %v", err))
			} else if resp.JSON200 != nil {
				tag = *resp.JSON200
			} else {
				errs = append(errs, fmt.Sprintf("Failed to load tag: %s", resp.Body))
			}
		}
		var aliases []string
		if tag.Aliases != nil {
			aliases = *tag.Aliases
		}
		app.renderTemplate(w, r, "tag_edit", TagEditData{
			Tag:         tag,
			Aliases:     aliases,
			Categories:  cats,
			CurrentName: name,
			IsNew:       isNew,
			Error:       strings.Join(errs, "; "),
		})

	case http.MethodPost:
		newName := r.FormValue("name")
		description := r.FormValue("description")
		category := r.FormValue("category")
		aliasesRaw := r.FormValue("aliases")

		var aliases []string
		for a := range strings.SplitSeq(aliasesRaw, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				aliases = append(aliases, a)
			}
		}

		tag := types.Tag{
			Name:        newName,
			Description: description,
			Aliases:     &aliases,
		}
		if category != "" {
			tag.Category = &category
		}

		urlName := name
		if isNew {
			urlName = newName
		}
		resp, err := app.api.PutTagWithResponse(ctx, urlName, tag)
		if err != nil || resp.StatusCode() >= 400 {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Failed to save tag: %v", err)
			} else {
				errMsg = fmt.Sprintf("Failed to save tag: %s", resp.Body)
			}
			app.renderTemplate(w, r, "tag_edit", TagEditData{
				Tag:         tag,
				Aliases:     aliases,
				Categories:  cats,
				CurrentName: name,
				IsNew:       isNew,
				Error:       errMsg,
			})
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)

	case http.MethodDelete:
		resp, err := app.api.DeleteTagWithResponse(ctx, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete tag: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete tag: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

func (app *App) handleTagConvertToAlias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceName := r.PathValue("name")
	targetName := r.FormValue("target")

	if targetName == "" || sourceName == targetName {
		http.Error(w, "Invalid target tag", http.StatusBadRequest)
		return
	}

	resp, err := app.api.ConvertTagToAliasWithResponse(ctx, sourceName, client.ConvertTagToAliasJSONRequestBody{
		Target: targetName,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert tag to alias: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode() >= 400 {
		http.Error(w, fmt.Sprintf("Failed to convert tag to alias: %s", resp.Body), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tags/"+targetName, http.StatusSeeOther)
}

func (app *App) handleTagCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var errs []string

	catLimit := 1000
	resp, err := app.api.GetTagCategoriesWithResponse(ctx, &client.GetTagCategoriesParams{Limit: &catLimit})
	if err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}
	cats := []types.TagCategory{}
	if resp != nil && resp.JSON200 != nil && resp.JSON200.Items != nil {
		cats = *resp.JSON200.Items
	}

	// Tag counts are now provided server-side via TagCategory.TagCount.
	tagCounts := map[string]int{}
	for _, c := range cats {
		if c.TagCount != nil {
			tagCounts[c.Name] = *c.TagCount
		}
	}

	app.renderTemplate(w, r, "tag_categories", TagCategoriesData{Categories: cats, TagCounts: tagCounts, Error: strings.Join(errs, "; ")})
}

func (app *App) handleTagCategoryEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == newResourceName

	switch r.Method {
	case http.MethodGet:
		cat := types.TagCategory{Color: "#888888"}
		var editErr string
		if !isNew {
			resp, err := app.api.GetTagCategoryWithResponse(ctx, name)
			if err != nil {
				editErr = fmt.Sprintf("Failed to load category: %v", err)
			} else if resp.JSON200 != nil {
				cat = *resp.JSON200
			} else {
				editErr = fmt.Sprintf("Failed to load category: %s", resp.Body)
			}
		}
		app.renderTemplate(w, r, "tag_category_edit", TagCategoryEditData{
			Category:    cat,
			CurrentName: name,
			IsNew:       isNew,
			Error:       editErr,
		})

	case http.MethodPost:
		newName := r.FormValue("name")
		description := r.FormValue("description")
		color := r.FormValue("color")

		cat := types.TagCategory{
			Name:        newName,
			Description: description,
			Color:       color,
		}

		urlName := name
		if isNew {
			urlName = newName
		}
		resp, err := app.api.PutTagCategoryWithResponse(ctx, urlName, cat)
		if err != nil || resp.StatusCode() >= 400 {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Failed to save category: %v", err)
			} else {
				errMsg = fmt.Sprintf("Failed to save category: %s", resp.Body)
			}
			app.renderTemplate(w, r, "tag_category_edit", TagCategoryEditData{
				Category:    cat,
				CurrentName: name,
				IsNew:       isNew,
				Error:       errMsg,
			})
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)

	case http.MethodDelete:
		resp, err := app.api.DeleteTagCategoryWithResponse(ctx, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete category: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete category: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)
	}
}

func (app *App) handleNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		// Create new note
		resp, err := app.api.CreateNoteWithResponse(ctx, client.CreateNoteJSONRequestBody{
			Title: "New Note",
		})
		if err != nil || resp.StatusCode() >= 400 {
			notesResp, _ := app.api.GetNotesWithResponse(ctx)
			notes := []types.Note{}
			if notesResp != nil && notesResp.JSON200 != nil && notesResp.JSON200.Items != nil {
				notes = *notesResp.JSON200.Items
			}
			errMsg := "Failed to create note"
			if err != nil {
				errMsg = fmt.Sprintf("Failed to create note: %v", err)
			}
			app.renderTemplate(w, r, "notes", NotesData{Notes: notes, Error: errMsg})
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/notes/%s", resp.JSON201.ID), http.StatusSeeOther)
		return
	}

	resp, err := app.api.GetNotesWithResponse(ctx)
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
	app.renderTemplate(w, r, "notes", NotesData{Notes: notes, Error: loadErr})
}

func (app *App) handleNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		noteID, err := uuid.Parse(id)
		if err != nil {
			app.renderTemplate(w, r, "note", NoteData{Error: fmt.Sprintf("Invalid note ID: %v", err)})
			return
		}
		resp, err := app.api.GetNoteWithResponse(ctx, noteID)
		if err != nil || resp.JSON200 == nil {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Note not found: %v", err)
			} else {
				errMsg = fmt.Sprintf("Note not found: %s", resp.Body)
			}
			app.renderTemplate(w, r, "note", NoteData{Error: errMsg})
			return
		}
		note := *resp.JSON200
		rendered := renderMarkdown(note.Content)
		isNew := note.Content == ""
		app.renderTemplate(w, r, "note", NoteData{Note: note, RenderedContent: rendered, IsNew: isNew})

	case http.MethodPut:
		noteID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "Invalid note ID", http.StatusBadRequest)
			return
		}
		resp, err := app.api.PutNoteWithResponse(ctx, noteID, client.PutNoteJSONRequestBody{
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
		resp, err := app.api.DeleteNoteWithResponse(ctx, noteID)
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

func (app *App) handleMedia(w http.ResponseWriter, r *http.Request) {
	// /media/{key...} → proxy to API /media/{key...}
	path := strings.TrimPrefix(r.URL.Path, "/media/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	resp, err := app.media.getRaw(r.Context(), "/media/"+path)
	if err != nil {
		http.Error(w, "Failed to fetch media", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	copyMediaResponse(w, resp)
}
