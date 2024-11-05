package helpers

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/joho/godotenv"
)

var app *config.AppConfig

// NewHelpers sets the config for the helpers package
func NewHelpers(a *config.AppConfig) {
	app = a
}

// ClientError logs the error status and sends a human-readable error response
func ClientError(w http.ResponseWriter, status int) {
	// Log error
	app.InfoLog.Println("Client error with status of: ", status)

	// Send human-readable error response
	http.Error(w, http.StatusText(status), status)
}

// ServerError logs the error and sends a 500 status code back to the user.
// Use this when we can't recover from an error, and want to show the user
// a generic error page.
func ServerError(w http.ResponseWriter, err error) {
	// Define the stack trace of the error
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	// Log error
	app.ErrorLog.Println(trace)

	// Send a 500 Internal Server Error response
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// isAuthenticated checks if the user is authenticated, or not.
// The user is authenticated if there is an entry in their session
// for "user_id".
func IsAuthenticated(r *http.Request) bool {
	return app.Session.Exists(r.Context(), "user_id")
}

// getAllDotEnv reads all the environment variables from the given
// .env file and puts them into a map.
func GetAllDotEnv(envfile string) map[string]any {
	godotenv.Load(envfile)

	connStr := os.Getenv("DB_STRING")
	prod, _ := strconv.ParseBool(os.Getenv("PROD"))
	useCache, _ := strconv.ParseBool(os.Getenv("USE_TEMPLATE_CACHE"))

	return map[string]any{
		"DATABASE_URL":       connStr,
		"PROD":               prod,
		"USE_TEMPLATE_CACHE": useCache,
	}
}
