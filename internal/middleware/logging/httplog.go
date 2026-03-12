package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/rs/zerolog/log"
)

// contextKey is a string type used for context value keys to avoid collisions.
type contextKey string

// requestIDKey is the context key for storing the request ID.
const requestIDKey contextKey = "requestID"

// statusWriter wraps http.ResponseWriter to capture the response status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status code and delegates to the underlying ResponseWriter.
func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// RequestLoggingMiddleware is HTTP middleware that assigns a request ID and logs each request.
func RequestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			requestID = hex.EncodeToString(b)
		}

		// Attach a request-scoped logger so all handlers get request_id in their logs.
		logger := log.Logger.With().Str("request_id", requestID).Logger()
		ctx := logger.WithContext(r.Context())
		ctx = context.WithValue(ctx, requestIDKey, requestID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-Id", requestID)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.RequestURI()).
			Int("status", sw.status).
			Str("content_type", r.Header.Get("Content-Type")).
			Int64("content_length", r.ContentLength).
			Msg("request")
	})
}

// RequestID returns the request ID from the context, or empty string if not set.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
