package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// handlePosts serves the posts listing page with search and infinite scroll support.
func (a *app) handlePosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := r.URL.Query().Get("search")
	if search == "" && !r.URL.Query().Has("search") {
		search = "sort:random"
	}
	cursor := r.URL.Query().Get("cursor")

	limit := 24
	params := &client.GetPostsParams{Limit: &limit}
	if search != "" {
		params.Search = &search
	}
	if cursor != "" {
		params.Cursor = &cursor
	}

	resp, err := a.api.GetPostsWithResponse(ctx, params)
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

	data := postsData{
		Posts:      posts,
		NextCursor: nextCursor,
		Search:     search,
		TagFilters: a.cfg.TagFilters,
		Error:      loadErr,
	}

	// HTMX partial request (search or infinite scroll)
	if r.Header.Get("HX-Request") == "true" {
		if loadErr != "" {
			http.Error(w, loadErr, http.StatusInternalServerError)
			return
		}
		a.renderTemplate(w, r, "posts-items", data)
		return
	}

	a.renderTemplate(w, r, "posts", data)
}

// handlePost serves the single post view and handles post deletion.
func (a *app) handlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	switch r.Method {
	case http.MethodGet:
		postID, err := uuid.Parse(id)
		if err != nil {
			a.renderTemplate(w, r, "post", postData{Error: fmt.Sprintf("Invalid post ID: %v", err)})
			return
		}
		resp, err := a.api.GetPostWithResponse(ctx, postID)
		if err != nil || resp.JSON200 == nil {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Post not found: %v", err)
			} else {
				errMsg = fmt.Sprintf("Post not found: %s", resp.Body)
			}
			a.renderTemplate(w, r, "post", postData{Error: errMsg})
			return
		}
		post := *resp.JSON200

		var fileSize int64
		if headResp, err := a.media.head(ctx, "/media"+mediaPath(post.ContentUrl)); err == nil {
			_ = headResp.Body.Close()
			fileSize = headResp.ContentLength
		}

		isVideo := strings.HasPrefix(post.MimeType, "video/")

		var similarPosts []types.Post
		if !isVideo {
			similarLimit := 12
			if similarResp, err := a.api.GetSimilarPostsWithResponse(ctx, postID, &client.GetSimilarPostsParams{Limit: &similarLimit}); err == nil && similarResp.JSON200 != nil {
				if similarResp.JSON200.Items != nil {
					similarPosts = *similarResp.JSON200.Items
				}
			}
		}

		a.renderTemplate(w, r, "post", postData{
			Post:         post,
			IsVideo:      isVideo,
			FileSize:     fileSize,
			SimilarPosts: similarPosts,
		})

	case http.MethodDelete:
		postID, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid post ID: %v", err), http.StatusBadRequest)
			return
		}
		resp, err := a.api.DeletePostWithResponse(ctx, postID)
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

func (a *app) handlePostNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	note := r.FormValue("note")

	postID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	resp, err := a.api.GetPostWithResponse(ctx, postID)
	if err != nil || resp.JSON200 == nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	post := *resp.JSON200
	post.Note = note
	putResp, err := a.api.PutPostWithResponse(ctx, postID, post)
	if err != nil || putResp.StatusCode() >= 400 {
		http.Error(w, "Failed to save note", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) handlePostTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	postID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	resp, err := a.api.GetPostWithResponse(ctx, postID)
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
			putResp, err := a.api.PutPostWithResponse(ctx, postID, post)
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
		putResp, err := a.api.PutPostWithResponse(ctx, postID, post)
		if err != nil || putResp.StatusCode() >= 400 {
			http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
			return
		}
	}

	// Re-fetch to get updated tags
	reResp, err := a.api.GetPostWithResponse(ctx, postID)
	if err != nil || reResp.JSON200 == nil {
		http.Error(w, "Failed to reload post", http.StatusInternalServerError)
		return
	}
	a.renderTemplate(w, r, "post-tags", postData{Post: *reResp.JSON200})
}

func (a *app) handleRegenerateThumbnail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	postID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	resp, err := a.api.RegeneratePostThumbnailWithResponse(ctx, postID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to regenerate thumbnail: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode() >= 400 {
		http.Error(w, fmt.Sprintf("Failed to regenerate thumbnail: %s", resp.Body), resp.StatusCode())
		return
	}

	http.Redirect(w, r, "/posts/"+id, http.StatusSeeOther)
}

func (a *app) handleTagSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	postIDStr := r.URL.Query().Get("post")
	exclude := r.URL.Query().Get("exclude")

	// Load existing post tags to exclude from suggestions
	excludeTags := map[string]bool{}
	if postIDStr != "" {
		if postID, err := uuid.Parse(postIDStr); err == nil {
			if resp, err := a.api.GetPostWithResponse(ctx, postID); err == nil && resp.JSON200 != nil {
				for _, t := range resp.JSON200.Tags {
					excludeTags[t] = true
				}
			}
		}
	}
	if exclude != "" {
		excludeTags[exclude] = true
	}

	// Paginate through all tags
	var allTags []types.Tag
	var cursor *string
	for {
		limit := 1000
		params := &client.GetTagsParams{Limit: &limit, Cursor: cursor}
		resp, err := a.api.GetTagsWithResponse(ctx, params)
		if err != nil || resp.JSON200 == nil {
			break
		}
		if resp.JSON200.Items != nil {
			allTags = append(allTags, *resp.JSON200.Items...)
		}
		if resp.JSON200.Cursor == nil || *resp.JSON200.Cursor == "" {
			break
		}
		cursor = resp.JSON200.Cursor
	}

	w.Header().Set("Content-Type", "text/html")
	for _, tag := range allTags {
		if excludeTags[tag.Name] {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(tag.Name), strings.ToLower(q)) {
			continue
		}
		_, _ = fmt.Fprintf(w, "<option value=%q>", tag.Name)
	}
}
