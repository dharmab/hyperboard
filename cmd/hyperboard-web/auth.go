package main

import (
	"net/http"
)

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		app.renderTemplate(w, r, "login.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if password != app.cfg.AdminPassword {
			app.renderTemplate(w, r, "login.html", map[string]any{"Error": "Invalid password"})
			return
		}
		setSessionCookie(w, app.cfg.SessionSecret)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

func (app *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *App) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !verifySession(app.cfg.SessionSecret, cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
