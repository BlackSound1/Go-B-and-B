package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/driver"
	"github.com/BlackSound1/Go-B-and-B/internal/handlers"
	"github.com/BlackSound1/Go-B-and-B/internal/helpers"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/BlackSound1/Go-B-and-B/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
)

const portNumber = ":8080"

// App config available throughout main package
var app config.AppConfig

// To manage sessions. Available throughout main package
var session *scs.SessionManager

func main() {

	db, err := run(".env")
	if err != nil {
		log.Fatal(err)
	}

	// Needed to do this outside of run() because we need to close the db connection
	// when main loop is finished, not when run() returns
	defer db.SQL.Close()

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
func run(envFile string) (*driver.DB, error) {

	// Load .env file
	err := godotenv.Load(envFile)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	// Change to true when in production
	app.InProduction = false

	// Define loggers. The | is a bitwise OR, so all flags get set to 1 integer value
	app.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Lets us store models in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// Create session info
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction // HTTPS

	// Associate session with app config
	app.Session = session

	// Connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL(os.Getenv("DB_STRING"))

	if err != nil {
		log.Fatal("Cannot connect to database!")
	}

	log.Println("Connected to database")

	// Create template cache and associate it with app config
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Set settings for config
	app.TemplateCache = tc
	app.UseCache = false

	// Create new repo and associate it with app config and db
	repo := handlers.NewRepo(&app, db)

	// Gives handlers package access to app config
	handlers.NewHandlers(repo)

	// Gives render package access to app config
	render.NewRenderer(&app)

	// Create helpers
	helpers.NewHelpers(&app)

	return db, nil
}
