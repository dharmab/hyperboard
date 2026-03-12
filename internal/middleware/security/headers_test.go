package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	t.Parallel()

	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	tests := []struct {
		header string
		want   string
	}{
		{"Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; media-src 'self'"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			t.Parallel()
			got := rec.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("header %s = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}
