package web

import (
	"net/http"
	"net/url"
)

const (
	maxFormBody   int64 = 1 << 20 // 1MB for regular form submissions
	maxUploadBody int64 = 8 << 30 // 8GB for file uploads
)

// maxBody wraps a handler to limit the request body size.
func maxBody(limit int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next(w, r)
	}
}

// registerRoutes registers all route handlers on the given ServeMux.
func (a *app) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handlePosts)
	mux.HandleFunc("/posts-partial", a.handlePosts)
	mux.HandleFunc("GET /posts/{id}/content", func(w http.ResponseWriter, r *http.Request) {
		id := url.PathEscape(r.PathValue("id"))
		resp, err := a.media.getRaw(r.Context(), "/api/v1/posts/"+id+"/content")
		if err != nil {
			http.Error(w, "Failed to fetch media", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		copyMediaResponse(w, resp)
	})
	mux.HandleFunc("GET /posts/{id}/download", func(w http.ResponseWriter, r *http.Request) {
		id := url.PathEscape(r.PathValue("id"))
		resp, err := a.media.getRaw(r.Context(), "/api/v1/posts/"+id+"/content/download")
		if err != nil {
			http.Error(w, "Failed to fetch media", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		copyMediaResponse(w, resp)
	})
	mux.HandleFunc("GET /posts/{id}/thumbnail", func(w http.ResponseWriter, r *http.Request) {
		id := url.PathEscape(r.PathValue("id"))
		resp, err := a.media.getRaw(r.Context(), "/api/v1/posts/"+id+"/thumbnail")
		if err != nil {
			http.Error(w, "Failed to fetch media", http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		copyMediaResponse(w, resp)
	})
	mux.HandleFunc("/posts/{id}", maxBody(maxFormBody, a.handlePost))
	mux.HandleFunc("GET /search.json", a.handleSearchJSON)
	mux.HandleFunc("/posts/{id}/note", maxBody(maxFormBody, a.handlePostNote))
	mux.HandleFunc("/posts/{id}/tags", maxBody(maxFormBody, a.handlePostTags))
	mux.HandleFunc("POST /posts/{id}/regenerate-thumbnail", a.handleRegenerateThumbnail)
	mux.HandleFunc("/posts/{id}/tags/{tag}", maxBody(maxFormBody, a.handlePostTags))
	mux.HandleFunc("/tag-suggestions", a.handleTagSuggestions)
	mux.HandleFunc("/upload", maxBody(maxUploadBody, a.handleUpload))
	mux.HandleFunc("/tags", a.handleTags)
	mux.HandleFunc("POST /tags/{name}/convert-to-alias", maxBody(maxFormBody, a.handleTagConvertToAlias))
	mux.HandleFunc("/tags/{name}", maxBody(maxFormBody, a.handleTagEdit))
	mux.HandleFunc("/tag-categories", a.handleTagCategories)
	mux.HandleFunc("/tag-categories/{name}", maxBody(maxFormBody, a.handleTagCategoryEdit))
	mux.HandleFunc("/notes", maxBody(maxFormBody, a.handleNotes))
	mux.HandleFunc("/notes/{id}", maxBody(maxFormBody, a.handleNote))
}
