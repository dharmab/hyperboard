package api

import (
	"fmt"
	"net/http"

	"encoding/json"
)

func respond(w http.ResponseWriter, code int, body any) {
	var b []byte
	var err error
	if body != nil {
		b, err = json.Marshal(body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to marshal response")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, err = w.Write(b)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to write response")
			return
		}
	} else {
		respondWithError(w, http.StatusNotFound, "Not found")
	}
}

func respondWithError(w http.ResponseWriter, code int, message string, args ...any) {
	e := Error{Message: fmt.Sprintf(message, args...)}
	var b []byte
	b, err := json.Marshal(e)
	if err != nil {
		http.Error(w, e.Message, code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, e.Message, code)
		return
	}
}
