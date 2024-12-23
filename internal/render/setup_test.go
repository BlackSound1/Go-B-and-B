package render

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/alexedwards/scs/v2"
)

var session *scs.SessionManager
var testApp config.AppConfig

// TestMain sets up the testing environment and runs the tests. It is the
// entrypoint for the testing framework.
func TestMain(m *testing.M) {

	// Change to true when in production
	testApp.InProduction = false

	// Define loggers. The | is a bitwise OR, so all flags get set to 1 integer value
	testApp.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Lets us store Reservations in the session
	gob.Register(models.Reservation{})

	// Create session info
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false // Can be false for testing purposes

	// Associate session with testApp config
	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

// Create a response writer
type myWriter struct{}

// Header returns an empty http.Header map to satisfy the http.ResponseWriter interface.
func (tw *myWriter) Header() http.Header {
	var h http.Header
	return h
}

// WriteHeader exists to satisfy the http.ResponseWriter interface
func (tw *myWriter) WriteHeader(i int) {}

// Write exists to satisfy the http.ResponseWriter interface
func (tw *myWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}
