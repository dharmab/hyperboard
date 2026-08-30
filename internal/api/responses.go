package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// respond writes a JSON response with the given status code and body.
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

// parameterBindingErrorHandler translates generated parameter binding failures into API errors.
func parameterBindingErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	log.Warn().Err(err).Msg("failed to bind request parameter")

	parameter := bindingErrorParameter(err)
	if parameter == "" {
		respondWithError(w, http.StatusBadRequest, "Request parameters are invalid")
		return
	}
	respondWithError(w, http.StatusBadRequest, "Request parameter %q is invalid", parameter)
}

func bindingErrorParameter(err error) string {
	if unescapedCookieError, ok := errors.AsType[*UnescapedCookieParamError](err); ok {
		return unescapedCookieError.ParamName
	}
	if unmarshalingError, ok := errors.AsType[*UnmarshalingParamError](err); ok {
		return unmarshalingError.ParamName
	}
	if requiredParameterError, ok := errors.AsType[*RequiredParamError](err); ok {
		return requiredParameterError.ParamName
	}
	if requiredHeaderError, ok := errors.AsType[*RequiredHeaderError](err); ok {
		return requiredHeaderError.ParamName
	}
	if invalidFormatError, ok := errors.AsType[*InvalidParamFormatError](err); ok {
		return invalidFormatError.ParamName
	}
	if tooManyValuesError, ok := errors.AsType[*TooManyValuesForParamError](err); ok {
		return tooManyValuesError.ParamName
	}
	return ""
}

// respondWithError writes a JSON error response with a formatted message.
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
