package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/config"
	"github.com/BlackSound1/Go-B-and-B/internal/driver"
	"github.com/BlackSound1/Go-B-and-B/internal/forms"
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

// NewTestRepo creates a new test repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
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
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get the room by ID
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	// 2020-01-01 -- 01/02 03:04:05PM '06 -0700

	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get the room by ID
	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
		Room:      room,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		http.Error(w, "form is not valid", http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert reservation into database!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't insert room restriction!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Send email to guest

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@here.com",
		Subject:  "Reservation Confirmation",
		Template: "guest_email_confirmation.html",
		Content: fmt.Sprintf(
			`
				<tr>
					<td>%s %s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s to %s</td>
				</tr>
			`,
			reservation.FirstName,
			reservation.LastName,
			reservation.Email,
			reservation.Phone,
			reservation.Room.RoomName,
			reservation.StartDate.Format("2006-01-02"),
			reservation.EndDate.Format("2006-01-02"),
		),
	}

	m.App.MailChan <- msg

	// Send email to property owner

	msg = models.MailData{
		To:       "me@here.com",
		From:     "me@here.com",
		Subject:  "Reservation Confirmation (Owner)",
		Template: "owner_email_confirmation.html",
		Content: fmt.Sprintf(
			`
				<tr>
					<td>%s %s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s to %s</td>
				</tr>
			`,
			reservation.FirstName,
			reservation.LastName,
			reservation.Email,
			reservation.Phone,
			reservation.Room.RoomName,
			reservation.StartDate.Format("2006-01-02"),
			reservation.EndDate.Format("2006-01-02"),
		),
	}

	m.App.MailChan <- msg

	m.App.Session.Put(r.Context(), "reservation", reservation)
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
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get the form values
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// Convert the dates to time
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't get parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Search for availbility in all rooms
	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't search for availability")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		// No rooms available
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
	// Parse request body
	err := r.ParseForm()
	if err != nil {
		// Can't parse form. Return appropriate json
		resp := jsonResponse{
			Ok:      false,
			Message: "Internal Server Error",
		}

		out, _ := json.MarshalIndent(resp, "", "\t")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	// Get form data
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")
	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	// Convert dates to time
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	// Check availability
	available, err := m.DB.SearchAvailabilityByDatesByRoomID(startDate, endDate, roomID)
	if err != nil {
		// There was a problem connecting to the database
		resp := jsonResponse{
			Ok:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "\t")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	// Create JSON response
	resp := jsonResponse{
		Ok:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	// Convert response to JSON
	out, _ := json.MarshalIndent(resp, "", "\t")

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
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get the reservation from the session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		m.App.Session.Put(r.Context(), "error", "Can't get room from database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

// ShowLogin displays the login page
func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// PostShowLogin handles the login form submission.
func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	// Prevents Session Fixation Attack. Best to renew token at each login and logout
	_ = m.App.Session.RenewToken(r.Context())

	// Parse the form
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	// Get form data
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)      // Create new form
	form.Required("email", "password") // Set certain fields as required
	form.IsEmail("email")              // Check if email is valid

	if !form.Valid() {
		// If the form is not valid, render the login page with the form data
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	// Try to authenticate user
	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		// If not authenticated, redirect and add error to session
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}

	// Add user to session, flash success message, and redirect
	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout logs the user out of the system
// It destroys the session, renews the CSRF token, and redirects to the login screen
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Destroy whole session
	_ = m.App.Session.Destroy(r.Context())

	// Renew token
	_ = m.App.Session.RenewToken(r.Context())

	// Redirect to login screen
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-new-reservations.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-all-reservations.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminReservationCalendar(w http.ResponseWriter, r *http.Request) {

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{})
}

