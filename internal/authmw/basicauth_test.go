package authmw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasicAuthMiddleware(t *testing.T) {
	t.Parallel()
	const password = "secret123"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := BasicAuthMiddleware(password, "/healthz", "/media/")

	tests := []struct {
		name       string
		path       string
		user       string
		pass       string
		setAuth    bool
		expectCode int
	}{
		{
			name:       "exempt exact path",
			path:       "/healthz",
			expectCode: http.StatusOK,
		},
		{
			name:       "exempt prefix path",
			path:       "/media/posts/123/content.webp",
			expectCode: http.StatusOK,
		},
		{
			name:       "valid auth",
			path:       "/api/posts",
			user:       "admin",
			pass:       password,
			setAuth:    true,
			expectCode: http.StatusOK,
		},
		{
			name:       "invalid password",
			path:       "/api/posts",
			user:       "admin",
			pass:       "wrong",
			setAuth:    true,
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "missing auth",
			path:       "/api/posts",
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "any username accepted",
			path:       "/api/posts",
			user:       "anyuser",
			pass:       password,
			setAuth:    true,
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			if tt.setAuth {
				req.SetBasicAuth(tt.user, tt.pass)
			}
			w := httptest.NewRecorder()
			middleware(handler).ServeHTTP(w, req)
			if w.Code != tt.expectCode {
				t.Errorf("status = %d, want %d", w.Code, tt.expectCode)
			}
		})
	}
}
