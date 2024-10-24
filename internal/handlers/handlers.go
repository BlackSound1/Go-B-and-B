package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/driver"
	"github.com/BlackSound1/Go-B-and-B/internal/forms"
	"github.com/BlackSound1/Go-B-and-B/internal/helpers"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/BlackSound1/Go-B-and-B/internal/render"
	"github.com/BlackSound1/Go-B-and-B/internal/repository"
	"github.com/BlackSound1/Go-B-and-B/internal/repository/dbrepo"
	"github.com/go-chi/chi"
)

// Repo is the repository used by the handlers
var Repo *Repository

// Use the Reposiory design pattern
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Handler function to handle HTTP requests
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the about page handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// Reservation makes a reservation
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	// Get the reservation from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can't get reservation from session"))
		return
	}

	// Get the room by ID
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Add the room name to the reservation
	res.Room.RoomName = room.RoomName

	// Add the reservation to the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// Format the dates correctly
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	// Create a string map for the dates
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	// Add the reservation data to the template
	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil), // Have access to form first time it's rendered
		Data:      data,
		StringMap: stringMap,
	})
}

// PostReservation saves a reservation to the DB
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	// Get the reservation from the session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can't get reservation from session"))
		return
	}

	// Try to parse the form
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get form values
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

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

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Write info to DB
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Create model
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	// Write info to DB
	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Take user to the Reservation Summary page
	m.App.Session.Put(r.Context(), "reservation", reservation) // Save to session
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals displays the General's Quarters room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors displays the Major's Suite room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability displays the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability checks for availability, and if there is availability,
// saves the reservation data to the session
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	// Get the form values
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// Convert the dates to time
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Search for availbility in all rooms
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		// No rooms available. Flash message and redirect
		m.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Add room data to template
	data := make(map[string]interface{})
	data["rooms"] = rooms

	// Add data to session
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// jsonRespose defines what a JSON response for availability is
type jsonResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	// Get form data
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	// Convert dates to time
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	// Check availability
	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)

	// Create JSON response
	resp := jsonResponse{
		Ok:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	// Convert response to JSON
	out, err := json.MarshalIndent(resp, "", "\t")

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Add json header
	w.Header().Set("Content-Type", "application/json")

	// Write json response
	w.Write(out)
}

// Contact displays the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	// The .(models.Reservation) is a type assertion
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)

	// If there is no reservation, redirect to home page
	if !ok {
		m.App.ErrorLog.Println("Can't get reservation from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

		return
	}

	// Remove the reservation from the session
	m.App.Session.Remove(r.Context(), "reservation")

	// Create a map to pass to the template
	data := make(map[string]interface{})
	data["reservation"] = reservation

	// Format the dates correctly
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	// Create a string map for the dates
	strigMap := make(map[string]string)
	strigMap["start_date"] = sd
	strigMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: strigMap,
	})
}

// ChooseRoom displays list of availabile rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL (/choose-room/{id})
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Get the reservation from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}

	// Add the room to the reservation
	res.RoomID = roomID

	// Add the reservation to the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// Redirect to make reservation
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// BookRoom takes URL parameters, builds a sessional variable, and takes user to make reservation screen
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {

	// Get the date from the URL query
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	// Get room name
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Convert dates
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	// Build a Reservation
	var res models.Reservation

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate
	res.Room.RoomName = room.RoomName

	// Add the reservation to the session
	m.App.Session.Put(r.Context(), "reservation", res)

	// Redirect to make reservation page
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}
