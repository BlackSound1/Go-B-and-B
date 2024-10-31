package main

import (
	"net/http"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func routes(app *config.AppConfig) http.Handler {

	// Create new multiplexer
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer) // Recoverer middleware to recover from panics more gracefully
	mux.Use(NoSurf)               // NoSurf middleware to prevent CSRF attacks on POST requests
	mux.Use(SessionLoad)

	// Set up routes
	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)

	// All routes starting with /admin will be protected
	mux.Route("/admin", func(r chi.Router) {
		// Protect these routes
		r.Use(Auth)

		r.Get("/dashboard", handlers.Repo.AdminDashboard)

		r.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		r.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		r.Get("/reservations-calendar", handlers.Repo.AdminReservationCalendar)
		r.Get("/reservations/{src}/{id}", handlers.Repo.AdminShowReservation)
		r.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
		r.Get("/process-reservation/{src}/{id}", handlers.Repo.AdminProcessReservation)
		r.Get("/delete-reservation/{src}/{id}", handlers.Repo.AdminDeleteReservation)
	})

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
