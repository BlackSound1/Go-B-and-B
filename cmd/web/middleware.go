package main

import (
	"net/http"

	"github.com/justinas/nosurf"
)

// NoSurf creates a nosurf CSRF token.
// Adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session on every request.
// Communicates session token to and from client in a cookie
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}
