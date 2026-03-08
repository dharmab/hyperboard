package api

import (
	"net/http"
)

// BasicAuthMiddleware returns a middleware that enforces HTTP Basic Auth.
// Any username is accepted; the password must match adminPassword.
// The paths in exemptPaths are allowed without authentication.
func BasicAuthMiddleware(adminPassword string, exemptPaths ...string) func(http.Handler) http.Handler {
	exempt := make(map[string]bool, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exempt[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			_, password, ok := r.BasicAuth()
			if !ok || password != adminPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="hyperboard"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
