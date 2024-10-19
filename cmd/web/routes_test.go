package main

import (
	"testing"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/go-chi/chi"
)

// TestRoutes tests the routes function, which sets up the routes for the web server.
// It verifies that the returned type is *chi.Mux.
func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing
	default:
		t.Errorf("type is not *chi.Mux, but is %T", v)
	}
}
