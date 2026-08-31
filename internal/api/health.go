package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
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
		zerolog.Ctx(ctx).Error().Err(err).Str("dependency", "database").Msg("readiness check failed")
		respondWithError(w, http.StatusServiceUnavailable, "Service is not ready")
		return
	}

	if err := s.mediaStore.Ping(ctx); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Str("dependency", "object_store").Msg("readiness check failed")
		respondWithError(w, http.StatusServiceUnavailable, "Service is not ready")
		return
	}

	respond(w, http.StatusOK, "OK")
}
