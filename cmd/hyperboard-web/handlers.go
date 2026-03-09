package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/rs/zerolog/log"
)

func (app *App) handlePosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")
	cursor := r.URL.Query().Get("cursor")

	q := url.Values{}
	q.Set("limit", "24")
	if search != "" {
		q.Set("search", search)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}

	var resp postsResponse
	var loadErr string
	if err := app.api.getWithQuery(ctx, "/api/v1/posts", q, &resp); err != nil {
		log.Error().Err(err).Str("search", search).Str("cursor", cursor).Msg("Failed to load posts")
		loadErr = fmt.Sprintf("Failed to load posts: %v", err)
	}

	posts := []types.Post{}
	if resp.Items != nil {
		posts = *resp.Items
	}

	nextCursor := ""
	if resp.Cursor != nil {
		nextCursor = *resp.Cursor
	}

	data := PostsData{
		Posts:      posts,
		NextCursor: nextCursor,
		Search:     search,
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
		var post types.Post
		if err := app.api.get(ctx, "/api/v1/posts/"+id, &post); err != nil {
			app.renderTemplate(w, r, "post", PostData{
				Error: fmt.Sprintf("Post not found: %v", err),
			})
			return
		}
		var fileSize int64
		if resp, err := app.api.head(ctx, "/media"+mediaPath(post.ContentUrl)); err == nil {
			_ = resp.Body.Close()
			fileSize = resp.ContentLength
		}
		var similarPosts []types.Post
		var similarResp postsResponse
		if err := app.api.getWithQuery(ctx, "/api/v1/posts/"+id+"/similar", url.Values{"limit": {"12"}}, &similarResp); err == nil {
			if similarResp.Items != nil {
				similarPosts = *similarResp.Items
			}
		}
		app.renderTemplate(w, r, "post", PostData{
			Post:         post,
			IsVideo:      strings.HasPrefix(post.MimeType, "video/"),
			FileSize:     fileSize,
			SimilarPosts: similarPosts,
		})

	case http.MethodDelete:
		if _, err := app.api.delete(ctx, "/api/v1/posts/"+id); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete post: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *App) handlePostNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	note := r.FormValue("note")

	var post types.Post
	if err := app.api.get(ctx, "/api/v1/posts/"+id, &post); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	post.Note = note
	if _, err := app.api.put(ctx, "/api/v1/posts/"+id, post, nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save note: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *App) handlePostTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	var post types.Post
	if err := app.api.get(ctx, "/api/v1/posts/"+id, &post); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodPost:
		tagName := r.FormValue("q")
		if tagName != "" {
			post.Tags = append(post.Tags, tagName)
			if _, err := app.api.put(ctx, "/api/v1/posts/"+id, post, nil); err != nil {
				http.Error(w, fmt.Sprintf("Failed to add tag: %v", err), http.StatusInternalServerError)
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
		if _, err := app.api.put(ctx, "/api/v1/posts/"+id, post, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove tag: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Re-fetch to get updated tags
	if err := app.api.get(ctx, "/api/v1/posts/"+id, &post); err != nil {
		http.Error(w, "Failed to reload post", http.StatusInternalServerError)
		return
	}
	app.renderTemplate(w, r, "post-tags", PostData{Post: post})
}

func (app *App) handleTagSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	postID := r.URL.Query().Get("post")
	exclude := r.URL.Query().Get("exclude")

	// Load existing post tags to exclude from suggestions
	excludeTags := map[string]bool{}
	if postID != "" {
		var post types.Post
		if err := app.api.get(ctx, "/api/v1/posts/"+postID, &post); err == nil {
			for _, t := range post.Tags {
				excludeTags[t] = true
			}
		}
	}
	if exclude != "" {
		excludeTags[exclude] = true
	}

	var resp tagsResponse
	query := url.Values{}
	query.Set("limit", "10")
	if q != "" {
		query.Set("search", q)
	}
	if err := app.api.getWithQuery(ctx, "/api/v1/tags", query, &resp); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if resp.Items != nil {
		for _, tag := range *resp.Items {
			if excludeTags[tag.Name] {
				continue
			}
			_, _ = fmt.Fprintf(w, "<option value=%q>", tag.Name)
		}
	}
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
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to open", header.Filename))
			continue
		}

		data, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to read", header.Filename))
			continue
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		force := r.FormValue("force") == "true"
		var post types.Post
		statusCode, respBody, err := app.api.uploadFile(ctx, data, contentType, force, &post)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", header.Filename, err))
			continue
		}
		if statusCode == http.StatusConflict {
			// Could be exact duplicate or similar posts found.
			// Pass through the API response body for the JS to handle.
			if isXHR {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write(respBody)
				return
			}
			var apiErr struct{ Message string }
			if json.Unmarshal(respBody, &apiErr) == nil {
				errors = append(errors, fmt.Sprintf("%s: %s", header.Filename, apiErr.Message))
			} else {
				errors = append(errors, fmt.Sprintf("%s: duplicate detected", header.Filename))
			}
			continue
		}
		if statusCode >= 400 {
			var apiErr struct{ Message string }
			if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
				errors = append(errors, fmt.Sprintf("%s: %s", header.Filename, apiErr.Message))
			} else {
				errors = append(errors, fmt.Sprintf("%s: HTTP %d", header.Filename, statusCode))
			}
			continue
		}
		lastPostID = post.ID
		log.Info().Str("id", post.ID.String()).Str("filename", header.Filename).Msg("uploaded post")
	}

	if isXHR {
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

	if len(errors) == len(files) {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": strings.Join(errors, "; ")})
		return
	}

	if len(errors) > 0 {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": fmt.Sprintf("Some uploads failed: %s", strings.Join(errors, "; "))})
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
	_ = json.NewEncoder(w).Encode(data)
}

func (app *App) handleTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all tags (paginate through all pages)
	var errs []string
	allTags := []types.Tag{}
	cursor := ""
	for {
		q := url.Values{}
		q.Set("limit", "1000")
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		var resp tagsResponse
		if err := app.api.getWithQuery(ctx, "/api/v1/tags", q, &resp); err != nil {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %v", err))
			break
		}
		if resp.Items != nil {
			allTags = append(allTags, *resp.Items...)
		}
		if resp.Cursor == nil || *resp.Cursor == "" {
			break
		}
		cursor = *resp.Cursor
	}

	// Fetch categories for color map
	var catResp tagCategoriesResponse
	if err := app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &catResp); err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}
	colorMap := map[string]string{}
	if catResp.Items != nil {
		for _, c := range *catResp.Items {
			colorMap[c.Name] = c.Color
		}
	}

	app.renderTemplate(w, r, "tags", TagsData{Tags: allTags, CategoryColors: colorMap, Error: strings.Join(errs, "; ")})
}

func (app *App) handleTagEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == "_new"

	// Fetch categories for dropdown
	var catResp tagCategoriesResponse
	var catErr string
	if err := app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &catResp); err != nil {
		catErr = fmt.Sprintf("Failed to load categories: %v", err)
	}
	cats := []types.TagCategory{}
	if catResp.Items != nil {
		cats = *catResp.Items
	}

	switch r.Method {
	case http.MethodGet:
		tag := types.Tag{}
		var errs []string
		if catErr != "" {
			errs = append(errs, catErr)
		}
		if !isNew {
			if err := app.api.get(ctx, "/api/v1/tags/"+name, &tag); err != nil {
				errs = append(errs, fmt.Sprintf("Failed to load tag: %v", err))
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
		_, err := app.api.put(ctx, "/api/v1/tags/"+urlName, tag, nil)
		if err != nil {
			app.renderTemplate(w, r, "tag_edit", TagEditData{
				Tag:         tag,
				Aliases:     aliases,
				Categories:  cats,
				CurrentName: name,
				IsNew:       isNew,
				Error:       fmt.Sprintf("Failed to save tag: %v", err),
			})
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)

	case http.MethodDelete:
		if _, err := app.api.delete(ctx, "/api/v1/tags/"+name); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete tag: %v", err), http.StatusInternalServerError)
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

	// Fetch the target tag (to get its current aliases)
	var targetTag types.Tag
	if err := app.api.get(ctx, "/api/v1/tags/"+targetName, &targetTag); err != nil {
		http.Error(w, fmt.Sprintf("Target tag not found: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch the source tag (to get its aliases)
	var sourceTag types.Tag
	if err := app.api.get(ctx, "/api/v1/tags/"+sourceName, &sourceTag); err != nil {
		http.Error(w, fmt.Sprintf("Source tag not found: %v", err), http.StatusBadRequest)
		return
	}

	// Find all posts tagged with the source tag
	var allPosts []types.Post
	cursor := ""
	for {
		q := url.Values{}
		q.Set("limit", "1000")
		q.Set("search", sourceName)
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		var resp postsResponse
		if err := app.api.getWithQuery(ctx, "/api/v1/posts", q, &resp); err != nil {
			break
		}
		if resp.Items != nil {
			allPosts = append(allPosts, *resp.Items...)
		}
		if resp.Cursor == nil || *resp.Cursor == "" {
			break
		}
		cursor = *resp.Cursor
	}

	// For each post, replace the source tag with the target tag
	for _, post := range allPosts {
		newTags := []types.TagName{}
		hasTarget := false
		for _, t := range post.Tags {
			if t == sourceName {
				continue
			}
			if t == targetName {
				hasTarget = true
			}
			newTags = append(newTags, t)
		}
		if !hasTarget {
			newTags = append(newTags, targetName)
		}
		post.Tags = newTags
		if _, err := app.api.put(ctx, fmt.Sprintf("/api/v1/posts/%s", post.ID), post, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to re-tag post %s: %v", post.ID, err), http.StatusInternalServerError)
			return
		}
	}

	// Build the new alias list for the target tag:
	// target's existing aliases + source's name + source's aliases
	var targetAliases []string
	if targetTag.Aliases != nil {
		targetAliases = append(targetAliases, *targetTag.Aliases...)
	}
	targetAliases = append(targetAliases, sourceName)
	if sourceTag.Aliases != nil {
		targetAliases = append(targetAliases, *sourceTag.Aliases...)
	}

	// Delete the source tag first (so its name is freed for use as an alias,
	// and its aliases are removed by CASCADE)
	if _, err := app.api.delete(ctx, "/api/v1/tags/"+sourceName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete source tag: %v", err), http.StatusInternalServerError)
		return
	}

	// Update target tag with the merged aliases
	targetTag.Aliases = &targetAliases
	if _, err := app.api.put(ctx, "/api/v1/tags/"+targetName, targetTag, nil); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update target tag aliases: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tags/"+targetName, http.StatusSeeOther)
}

func (app *App) handleTagCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var errs []string

	var resp tagCategoriesResponse
	if err := app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &resp); err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}
	cats := []types.TagCategory{}
	if resp.Items != nil {
		cats = *resp.Items
	}

	// Count tags per category
	tagCounts := map[string]int{}
	cursor := ""
	for {
		q := url.Values{}
		q.Set("limit", "1000")
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		var tagsResp tagsResponse
		if err := app.api.getWithQuery(ctx, "/api/v1/tags", q, &tagsResp); err != nil {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %v", err))
			break
		}
		if tagsResp.Items != nil {
			for _, t := range *tagsResp.Items {
				if t.Category != nil {
					tagCounts[*t.Category]++
				}
			}
		}
		if tagsResp.Cursor == nil || *tagsResp.Cursor == "" {
			break
		}
		cursor = *tagsResp.Cursor
	}

	app.renderTemplate(w, r, "tag_categories", TagCategoriesData{Categories: cats, TagCounts: tagCounts, Error: strings.Join(errs, "; ")})
}

func (app *App) handleTagCategoryEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == "_new"

	switch r.Method {
	case http.MethodGet:
		cat := types.TagCategory{Color: "#888888"}
		var editErr string
		if !isNew {
			if err := app.api.get(ctx, "/api/v1/tagCategories/"+name, &cat); err != nil {
				editErr = fmt.Sprintf("Failed to load category: %v", err)
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
		_, err := app.api.put(ctx, "/api/v1/tagCategories/"+urlName, cat, nil)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to save category: %v", err)
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
		if _, err := app.api.delete(ctx, "/api/v1/tagCategories/"+name); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete category: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)
	}
}

func (app *App) handleNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		// Create new note
		note := types.Note{Title: "New Note"}
		var created types.Note
		if _, err := app.api.post(ctx, "/api/v1/notes", note, &created); err != nil {
			var resp notesResponse
			_ = app.api.get(ctx, "/api/v1/notes", &resp)
			notes := []types.Note{}
			if resp.Items != nil {
				notes = *resp.Items
			}
			app.renderTemplate(w, r, "notes", NotesData{Notes: notes, Error: fmt.Sprintf("Failed to create note: %v", err)})
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/notes/%s", created.ID), http.StatusSeeOther)
		return
	}

	var resp notesResponse
	var loadErr string
	if err := app.api.get(ctx, "/api/v1/notes", &resp); err != nil {
		loadErr = fmt.Sprintf("Failed to load notes: %v", err)
	}
	notes := []types.Note{}
	if resp.Items != nil {
		notes = *resp.Items
	}
	app.renderTemplate(w, r, "notes", NotesData{Notes: notes, Error: loadErr})
}

func (app *App) handleNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		if id == "_new" {
			note := types.Note{Title: "New Note"}
			var created types.Note
			if _, err := app.api.post(ctx, "/api/v1/notes", note, &created); err != nil {
				app.renderTemplate(w, r, "note", NoteData{Error: fmt.Sprintf("Failed to create note: %v", err)})
				return
			}
			http.Redirect(w, r, fmt.Sprintf("/notes/%s", created.ID), http.StatusSeeOther)
			return
		}
		var note types.Note
		if err := app.api.get(ctx, "/api/v1/notes/"+id, &note); err != nil {
			app.renderTemplate(w, r, "note", NoteData{Error: fmt.Sprintf("Note not found: %v", err)})
			return
		}
		rendered := renderMarkdown(note.Content)
		isNew := note.Content == ""
		app.renderTemplate(w, r, "note", NoteData{Note: note, RenderedContent: rendered, IsNew: isNew})

	case http.MethodPut:
		var note types.Note
		if err := app.api.get(ctx, "/api/v1/notes/"+id, &note); err != nil {
			http.Error(w, fmt.Sprintf("Failed to load note: %v", err), http.StatusInternalServerError)
			return
		}
		note.Title = r.FormValue("title")
		note.Content = r.FormValue("content")
		if _, err := app.api.put(ctx, "/api/v1/notes/"+id, note, nil); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save note: %v", err), http.StatusInternalServerError)
			return
		}
		// Return rendered markdown for HTMX swap
		rendered := renderMarkdown(note.Content)
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<div id="note-view" class="note-content mt-2">%s</div>`, string(rendered))

	case http.MethodDelete:
		if _, err := app.api.delete(ctx, "/api/v1/notes/"+id); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete note: %v", err), http.StatusInternalServerError)
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

	resp, err := app.api.getRaw(r.Context(), "/media/"+path)
	if err != nil {
		http.Error(w, "Failed to fetch media", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Media not found", resp.StatusCode)
		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = io.Copy(w, resp.Body)
}
