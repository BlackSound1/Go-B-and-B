package render

import (
	"encoding/gob"
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

func TestMain(m *testing.M) {

	// Change to true when in production
	testApp.InProduction = false

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
