package api

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, "OK")
}

func (s *Server) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Database is not ready: %v", err)
		return
	}

	respond(w, http.StatusOK, "OK")
}
