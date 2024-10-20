package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/forms"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/BlackSound1/Go-B-and-B/internal/render"
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

	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	// Get the remote IP address from the session
	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")

	// Add it to the string map to show on the page
	stringMap["remote_ip"] = remoteIP

	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{StringMap: stringMap})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// Create an empty reservation when page is first loaded
	var emptyReservation models.Reservation

	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil), // Have access to form first time it's rendered
		Data: data,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm() // Try to parse the form

	if err != nil {
		log.Println(err)
		return
	}

	// Get form values
	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	// Create form
	form := forms.New(r.PostForm)

	// Validate form for required fields
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	// If the form is invalid, redisplay the page with the form values
	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})

		return
	}

	// Take user to the Reservation Summary page
	m.App.Session.Put(r.Context(), "reservation", reservation) // Save to session
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	// Get the form values
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	w.Write([]byte("start: " + start + ", end: " + end))
}

// jsonRespose defines what a JSON response for availability is
type jsonResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	// Create JSON response
	resp := jsonResponse{
		Ok:      true,
		Message: "Available!",
	}

	// Convert response to JSON
	out, err := json.MarshalIndent(resp, "", "\t")

	if err != nil {
		log.Println(err)
	}

	// Add json header
	w.Header().Set("Content-Type", "application/json")

	// Write json response
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// The .(models.Reservation) is a type assertion
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	// If there is no reservation, redirect to home page
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

		return
	}

	// Remove the reservation from the session
	m.App.Session.Remove(r.Context(), "reservation")

	// Create a map to pass to the template
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
