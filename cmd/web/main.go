package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BlackSound1/Go-B-and-B/pkg/config"
	"github.com/BlackSound1/Go-B-and-B/pkg/handlers"
	"github.com/BlackSound1/Go-B-and-B/pkg/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

// App config available throughout main package
var app config.AppConfig

// To manage sessions. Available throughout main package
var session *scs.SessionManager

func main() {

	// Change to true when in production
	app.InProduction = false

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
		log.Fatal("cannot create template cache")
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

	fmt.Println("Starting server on port", portNumber)

	// Create new server
	serv := &http.Server{
		Addr:    portNumber,
		Handler: routes(),
	}

	// Start server
	err = serv.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
