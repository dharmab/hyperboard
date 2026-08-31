package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		origin   string
		referer  string
		insecure bool
		want     int
	}{
		{name: "safe request without source", method: http.MethodGet, want: http.StatusNoContent},
		{name: "same HTTPS origin", method: http.MethodPost, origin: "https://hyperboard.example.com", want: http.StatusNoContent},
		{name: "same HTTP origin in local development", method: http.MethodPost, origin: "http://hyperboard.example.com", insecure: true, want: http.StatusNoContent},
		{name: "same-origin referer", method: http.MethodDelete, referer: "https://hyperboard.example.com/tags", want: http.StatusNoContent},
		{name: "sibling subdomain", method: http.MethodPost, origin: "https://evil.example.com", want: http.StatusForbidden},
		{name: "wrong scheme", method: http.MethodPost, origin: "http://hyperboard.example.com", want: http.StatusForbidden},
		{name: "missing source", method: http.MethodPost, want: http.StatusForbidden},
		{name: "malformed source", method: http.MethodPost, origin: "://bad", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := &app{cfg: &config{InsecureSessionCookie: tt.insecure}}
			handler := a.csrfMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			target := "https://hyperboard.example.com/test"
			if tt.insecure {
				target = "http://hyperboard.example.com/test"
			}
			req := httptest.NewRequestWithContext(t.Context(), tt.method, target, nil)
			req.Host = "hyperboard.example.com"
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Referer", tt.referer)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.want, recorder.Code)
		})
	}
}
