package handlers

import (
	"net/http"

	"github.com/BlackSound1/Go-B-and-B/pkg/config"
	"github.com/BlackSound1/Go-B-and-B/pkg/models"
	"github.com/BlackSound1/Go-B-and-B/pkg/render"
)

// Use the Reposiory design pattern
type Repository struct {
	App *config.AppConfig
}

// Repo is the repository used by the handlers
var Repo *Repository

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{App: a}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Handler function to handle HTTP requests
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr // Grab remote IP address

	// Store the IP address in the session
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	// Get the remote IP address from the session
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	// Add it to the string map to show on the page
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, "about.page.tmpl", &models.TemplateData{StringMap: stringMap})
}
