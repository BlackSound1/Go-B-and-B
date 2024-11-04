package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/BlackSound1/Go-B-and-B/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates"
var functions = template.FuncMap{
	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

// TestMain sets up the testing environment and runs the tests. It is the
// entrypoint for the testing framework.
func TestMain(m *testing.M) {

	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Cannot load .env file")
		return
	}

	// Change to true when in production
	app.InProduction = false

	// Lets us store models in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// Define loggers. The | is a bitwise OR, so all flags get set to 1 integer value
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Create session info
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction // HTTPS

	// Associate session with app config
	app.Session = session

	// Set up dummy mail channel
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan
	defer close(app.MailChan)
	listenForMail()

	// Create template cache and associate it with app config
	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	// Set settings for config
	app.TemplateCache = tc
	app.UseCache = true // To prevent call to CreateTemplateCache() in RenderTemplate()

	// Create new repo and associate it with app config
	repo := NewTestRepo(&app)

	// Gives handlers package access to app config
	NewHandlers(repo)

	// Gives render package access to app config
	render.NewRenderer(&app)

	// Run the tests
	os.Exit(m.Run())
}

// listenForMail starts a goroutine to listen for messages on the app's mail channel.
// This is just a dummy implementation for testing
func listenForMail() {
	go func() {
		for {
			<-app.MailChan
		}
	}()
}

// getRoutes returns the chi multiplexer with routes set up for application testing
func getRoutes() http.Handler {

	// Create new multiplexer
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer) // Recoverer middleware to recover from panics more gracefully
	// mux.Use(NoSurf)               // NoSurf middleware to prevent CSRF attacks on POST requests
	mux.Use(SessionLoad)

	// Set up routes
	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/contact", Repo.Contact)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Post("/search-availability-json", Repo.AvailabilityJSON)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Get("/admin/dashboard", Repo.AdminDashboard)

	mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
	mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
	mux.Get("/admin/reservations-calendar", Repo.AdminReservationCalendar)
	mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationCalendar)
	mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservation)
	mux.Post("/admin/reservations/{src}/{id}", Repo.AdminPostShowReservation)
	mux.Get("/admin/process-reservation/{src}/{id}/do", Repo.AdminProcessReservation)
	mux.Get("/admin/delete-reservation/{src}/{id}/do", Repo.AdminDeleteReservation)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

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

// CreateTestTemplateCache creates a map of template names to template sets for testing purposes.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	// myCache := make(map[string]*template.Template)
	myCache := map[string]*template.Template{} // Same as above

	// Get all of the files named *.page.tmpl from ./templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))

	if err != nil {
		return myCache, err
	}

	// Range through all files ending with *.page.tmpl
	for _, page := range pages {
		// Get the file name
		name := filepath.Base(page)

		// Create a template with the name of the page
		// Associate the template with the name of the page
		templateSet, err := template.New(name).Funcs(functions).ParseFiles(page)

		if err != nil {
			return myCache, err
		}

		// Get the layout
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

		if err != nil {
			return myCache, err
		}

		// If there is a layout
		if len(matches) > 0 {
			templateSet, err = templateSet.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))

			if err != nil {
				return myCache, err
			}
		}

		// Add the templateSet to the map
		myCache[name] = templateSet
	}

	return myCache, nil
}
