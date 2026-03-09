package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func respond(w http.ResponseWriter, code int, body any) {
	if body == nil {
		w.WriteHeader(code)
		return
	}
	b, err := json.Marshal(body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to marshal response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(b); err != nil {
		log.Error().Err(err).Msg("failed to write response body")
	}
}

func respondWithError(w http.ResponseWriter, code int, message string, args ...any) {
	e := Error{Message: fmt.Sprintf(message, args...)}
	b, err := json.Marshal(e)
	if err != nil {
		http.Error(w, e.Message, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if _, err = w.Write(b); err != nil {
		log.Error().Err(err).Msg("failed to write error response body")
	}
}
