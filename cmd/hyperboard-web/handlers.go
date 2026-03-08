package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/dharmab/hyperboard/pkg/types"
)

func (app *App) handleGallery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "recent"
	}
	search := r.URL.Query().Get("search")
	cursor := r.URL.Query().Get("cursor")

	q := url.Values{}
	q.Set("limit", "24")
	q.Set("sort", sort)
	if search != "" {
		q.Set("search", search)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}

	var resp postsResponse
	if err := app.api.getWithQuery(ctx, "/api/v1/posts", q, &resp); err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	posts := []types.Post{}
	if resp.Items != nil {
		posts = *resp.Items
	}

	nextCursor := ""
	if resp.Cursor != nil {
		nextCursor = *resp.Cursor
	}

	data := GalleryData{
		Posts:      posts,
		NextCursor: nextCursor,
		Sort:       sort,
		Search:     search,
	}

	// HTMX partial request (infinite scroll)
	if r.Header.Get("HX-Request") == "true" && cursor != "" {
		app.renderTemplate(w, r, "gallery-items", data)
		return
	}

	app.renderTemplate(w, r, "gallery", data)
}

func (app *App) handlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		var post types.Post
		if err := app.api.get(ctx, "/api/v1/posts/"+id, &post); err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		app.renderTemplate(w, r, "post", PostData{
			Post:    post,
			IsVideo: strings.HasPrefix(post.MimeType, "video/"),
		})

	case http.MethodDelete:
		_, _ = app.api.delete(ctx, "/api/v1/posts/"+id)
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
	_, _ = app.api.put(ctx, "/api/v1/posts/"+id, post, nil)
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
		tagName := r.FormValue("tag-input")
		if tagName != "" {
			post.Tags = append(post.Tags, tagName)
			_, _ = app.api.put(ctx, "/api/v1/posts/"+id, post, nil)
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
		_, _ = app.api.put(ctx, "/api/v1/posts/"+id, post, nil)
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
			_, _ = fmt.Fprintf(w, "<option value=%q>", tag.Name)
		}
	}
}

func (app *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodGet {
		app.renderTemplate(w, r, "upload", nil)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": "Failed to parse form"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": "No file provided"})
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": "Failed to read file"})
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var post types.Post
	statusCode, err := app.api.uploadFile(ctx, data, contentType, &post)
	if err != nil || statusCode >= 400 {
		app.renderTemplate(w, r, "upload", map[string]any{"Error": "Upload failed"})
		return
	}

	// Apply tags if provided
	tagsStr := r.FormValue("tags")
	noteStr := r.FormValue("note")
	if tagsStr != "" || noteStr != "" {
		post.Tags = strings.Fields(tagsStr)
		post.Note = noteStr
		_, _ = app.api.put(ctx, fmt.Sprintf("/api/v1/posts/%s", post.ID), post, nil)
	}

	http.Redirect(w, r, fmt.Sprintf("/posts/%s", post.ID), http.StatusSeeOther)
}

func (app *App) handleTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all tags (paginate through all pages)
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
	_ = app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &catResp)
	colorMap := map[string]string{}
	if catResp.Items != nil {
		for _, c := range *catResp.Items {
			colorMap[c.Name] = c.Color
		}
	}

	app.renderTemplate(w, r, "tags", TagsData{Tags: allTags, CategoryColors: colorMap})
}

func (app *App) handleTagEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == "_new"

	// Fetch categories for dropdown
	var catResp tagCategoriesResponse
	_ = app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &catResp)
	cats := []types.TagCategory{}
	if catResp.Items != nil {
		cats = *catResp.Items
	}

	switch r.Method {
	case http.MethodGet:
		tag := types.Tag{}
		if !isNew {
			_ = app.api.get(ctx, "/api/v1/tags/"+name, &tag)
		}
		app.renderTemplate(w, r, "tag_edit", TagEditData{
			Tag:         tag,
			Categories:  cats,
			CurrentName: name,
			IsNew:       isNew,
		})

	case http.MethodPost:
		newName := r.FormValue("name")
		description := r.FormValue("description")
		category := r.FormValue("category")

		tag := types.Tag{
			Name:        newName,
			Description: description,
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
				Categories:  cats,
				CurrentName: name,
				IsNew:       isNew,
				Error:       "Failed to save tag",
			})
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)

	case http.MethodDelete:
		_, _ = app.api.delete(ctx, "/api/v1/tags/"+name)
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

func (app *App) handleTagCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var resp tagCategoriesResponse
	_ = app.api.getWithQuery(ctx, "/api/v1/tagCategories", url.Values{"limit": {"1000"}}, &resp)
	cats := []types.TagCategory{}
	if resp.Items != nil {
		cats = *resp.Items
	}
	app.renderTemplate(w, r, "tag_categories", TagCategoriesData{Categories: cats})
}

func (app *App) handleTagCategoryEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == "_new"

	switch r.Method {
	case http.MethodGet:
		cat := types.TagCategory{Color: "#888888"}
		if !isNew {
			_ = app.api.get(ctx, "/api/v1/tagCategories/"+name, &cat)
		}
		app.renderTemplate(w, r, "tag_category_edit", TagCategoryEditData{
			Category:    cat,
			CurrentName: name,
			IsNew:       isNew,
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
			app.renderTemplate(w, r, "tag_category_edit", TagCategoryEditData{
				Category:    cat,
				CurrentName: name,
				IsNew:       isNew,
				Error:       "Failed to save category",
			})
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)

	case http.MethodDelete:
		_, _ = app.api.delete(ctx, "/api/v1/tagCategories/"+name)
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)
	}
}

func (app *App) handleNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == http.MethodPost {
		// Create new note
		note := types.Note{Title: "New Note"}
		var created types.Note
		_, _ = app.api.post(ctx, "/api/v1/notes", note, &created)
		http.Redirect(w, r, fmt.Sprintf("/notes/%s", created.ID), http.StatusSeeOther)
		return
	}

	var resp notesResponse
	_ = app.api.get(ctx, "/api/v1/notes", &resp)
	notes := []types.Note{}
	if resp.Items != nil {
		notes = *resp.Items
	}
	app.renderTemplate(w, r, "notes", NotesData{Notes: notes})
}

func (app *App) handleNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		if id == "_new" {
			note := types.Note{Title: "New Note"}
			var created types.Note
			_, _ = app.api.post(ctx, "/api/v1/notes", note, &created)
			http.Redirect(w, r, fmt.Sprintf("/notes/%s", created.ID), http.StatusSeeOther)
			return
		}
		var note types.Note
		if err := app.api.get(ctx, "/api/v1/notes/"+id, &note); err != nil {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		rendered := renderMarkdown(note.Content)
		app.renderTemplate(w, r, "note", NoteData{Note: note, RenderedContent: string(rendered)})

	case http.MethodPut:
		var note types.Note
		_ = app.api.get(ctx, "/api/v1/notes/"+id, &note)
		note.Title = r.FormValue("title")
		note.Content = r.FormValue("content")
		_, _ = app.api.put(ctx, "/api/v1/notes/"+id, note, nil)
		// Return rendered markdown for HTMX swap
		rendered := renderMarkdown(note.Content)
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<div id="note-view" class="note-content mt-2">%s</div>`, rendered)

	case http.MethodDelete:
		_, _ = app.api.delete(ctx, "/api/v1/notes/"+id)
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
	}
}
