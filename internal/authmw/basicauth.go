package authmw

import (
	"net/http"
	"strings"
)

// AdminUsername is the conventional username used for HTTP Basic Auth.
const AdminUsername = "admin"

// BasicAuthMiddleware returns a middleware that enforces HTTP Basic Auth.
// Any username is accepted; the password must match adminPassword.
// The paths in exemptPaths are allowed without authentication. Paths
// ending in "/" are treated as prefixes (e.g. "/media/" exempts all
// paths starting with "/media/").
func BasicAuthMiddleware(adminPassword string, exemptPaths ...string) func(http.Handler) http.Handler {
	var exemptPrefixes []string
	exemptExact := make(map[string]bool)
	for _, p := range exemptPaths {
		if strings.HasSuffix(p, "/") {
			exemptPrefixes = append(exemptPrefixes, p)
		} else {
			exemptExact[p] = true
		}
	}
	isExempt := func(path string) bool {
		if exemptExact[path] {
			return true
		}
		for _, prefix := range exemptPrefixes {
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		return false
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isExempt(r.URL.Path) {
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
