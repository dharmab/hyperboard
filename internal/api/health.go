package api

import (
	"context"
	"net/http"
	"time"
)

// GetHealth handles liveness probe requests.
func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, "OK")
}

// GetReadiness handles readiness probe requests by checking connectivity to the database and object store.
func (s *Server) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.sqlStore.Ping(ctx); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Database is not ready: %v", err)
		return
	}

	if err := s.mediaStore.Ping(ctx); err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Object store is not ready: %v", err)
		return
	}

	respond(w, http.StatusOK, "OK")
}
