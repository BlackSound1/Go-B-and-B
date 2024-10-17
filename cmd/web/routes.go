package main

import (
	"net/http"

	"github.com/BlackSound1/Go-B-and-B/pkg/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes() http.Handler {

	// Create new multiplexer
	mux := chi.NewRouter()

	// Use the Recoverer middleware to recover from panics more gracefully
	mux.Use(middleware.Recoverer)
	// mux.Use(WriteToConsole)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux
}
