package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/handlers"
	"github.com/BlackSound1/Go-B-and-B/internal/helpers"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/BlackSound1/Go-B-and-B/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

// App config available throughout main package
var app config.AppConfig

// To manage sessions. Available throughout main package
var session *scs.SessionManager

func main() {

	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting server on port", portNumber)

	// Create new server
	serv := createNewServer()

	// Start server
	err = serv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}

// createNewServer creates a new HTTP server with the specified address and handler.
func createNewServer() *http.Server {
	return &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
}

// run initializes the app config, template cache, and session info.
func run() error {

	// Change to true when in production
	app.InProduction = false

	// Define loggers. The | is a bitwise OR, so all flags get set to 1 integer value
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Lets us store Reservations in the session
	gob.Register(models.Reservation{})

	// Create session info
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction // HTTPS

	// Associate session with app config
	app.Session = session

	// Create template cache and associate it with app config
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Set settings for config
	app.TemplateCache = tc
	app.UseCache = false

	// Create new repo and associate it with app config
	repo := handlers.NewRepo(&app)

	// Gives handlers package access to app config
	handlers.NewHandlers(repo)

	// Gives render package access to app config
	render.NewTemplates(&app)

	// Create helpers
	helpers.NewHelpers(&app)

	return nil
}
