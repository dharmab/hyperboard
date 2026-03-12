package web

import (
	"net/http"
)

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.renderTemplate(w, r, "login.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if password != a.cfg.AdminPassword {
			a.renderTemplate(w, r, "login.html", map[string]any{"Error": "Invalid password"})
			return
		}
		setSessionCookie(w, a.cfg.SessionSecret)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *app) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !verifySession(a.cfg.SessionSecret, cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
