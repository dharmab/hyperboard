package web

import (
	"net/http"
	"strings"
)

func (a *app) handleMedia(w http.ResponseWriter, r *http.Request) {
	// /media/{key...} → proxy to API /media/{key...}
	path := strings.TrimPrefix(r.URL.Path, "/media/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	resp, err := a.media.getRaw(r.Context(), "/media/"+path)
	if err != nil {
		http.Error(w, "Failed to fetch media", http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	copyMediaResponse(w, resp)
}
