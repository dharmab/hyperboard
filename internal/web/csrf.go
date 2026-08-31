package web

import (
	"net/http"
	"net/url"
	"strings"
)

// csrfMiddleware rejects unsafe requests that do not identify this exact web
// origin. It runs after session authentication, so the public login flow is
// intentionally unaffected.
func (a *app) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		source := r.Header.Get("Origin")
		allowPath := false
		if source == "" {
			source = r.Header.Get("Referer")
			allowPath = true
		}
		if !a.isSameOrigin(r, source, allowPath) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func (a *app) isSameOrigin(r *http.Request, source string, allowPath bool) bool {
	if source == "" || r.Host == "" {
		return false
	}
	u, err := url.Parse(source)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if !allowPath && (u.Path != "" && u.Path != "/" || u.RawQuery != "") {
		return false
	}

	requestScheme := "https"
	if r.TLS == nil && a.cfg.InsecureSessionCookie {
		requestScheme = "http"
	}
	return u.Scheme == requestScheme && strings.EqualFold(u.Host, r.Host)
}
