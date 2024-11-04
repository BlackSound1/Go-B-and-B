package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/BlackSound1/Go-B-and-B/internal/driver"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/go-chi/chi"
)

// Create a set of tests to run
var handlerTests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", "GET", http.StatusOK},
	{"about", "/about", "GET", http.StatusOK},
	{"contact", "/contact", "GET", http.StatusOK},
	{"gq", "/generals-quarters", "GET", http.StatusOK},
	{"ms", "/majors-suite", "GET", http.StatusOK},
	{"sa", "/search-availability", "GET", http.StatusOK},
	{"mr", "/make-reservation", "GET", http.StatusOK},
	{"non-existant-route", "/nothing-here", "GET", http.StatusNotFound},
	{"login", "/user/login", "GET", http.StatusOK},
	{"logout", "/user/logout", "GET", http.StatusOK},
	{"dashboard", "/admin/dashboard", "GET", http.StatusOK},
	{"reservation - new", "/admin/reservations-new", "GET", http.StatusOK},
	{"reservation - all", "/admin/reservations-all", "GET", http.StatusOK},
	{"reservation - show", "/admin/reservations/new/1/show", "GET", http.StatusOK},
	{"reservation-calendar", "/admin/reservations-calendar", "GET", http.StatusOK},
	{"reservation-calendar-with-params", "/admin/reservations-calendar?y=2020&m=2", "GET", http.StatusOK},
}

// TestHandlers tests all the routes in the application. It sends a GET request to
// each route and checks that the status code of the response is as expected.
func TestHandlers(t *testing.T) {
	routes := getRoutes() // Get the routes to test

	testServer := httptest.NewTLSServer(routes) // Create a new test server

	defer testServer.Close() // Close the test server whenever done

	for _, test := range handlerTests {
		baseURL := testServer.URL

		// Use the test server client to send a GET request and get the response
		resp, err := testServer.Client().Get(baseURL + test.url)
		if err != nil {
			t.Log("Error", err)
			t.Fatal(err)
		}

		// Check that the status code is as expected
		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}

// Create a set of tests to run
var reservationTests = []struct {
	name               string
	reservation        models.Reservation
	expectedStatusCode int
	expectedLocation   string
	expectedHTML       string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room:   models.Room{ID: 1, RoomName: "General's Quarters"},
		},
		expectedStatusCode: http.StatusOK,
		expectedHTML:       `action="/make-reservation"`,
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
	{
		name: "non-existent-room",
		reservation: models.Reservation{
			RoomID: 100,
			Room:   models.Room{ID: 100, RoomName: "General's Quarters"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
		expectedHTML:       "",
	},
}

// TestReservation tests the Reservation handler for various scenarios.
func TestReservation(t *testing.T) {
	for _, test := range reservationTests {
		req, _ := http.NewRequest("GET", "/make-reservation", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		recorder := httptest.NewRecorder()

		// If there is a room, put the reservation in the session
		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		handler := http.HandlerFunc(Repo.Reservation)
		handler.ServeHTTP(recorder, req)

		// Test status code
		if recorder.Code != test.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedStatusCode)
		}

		// Test expected location
		if test.expectedLocation != "" {
			// Get the URL from the test
			actualLocation, _ := recorder.Result().Location()

			if actualLocation.String() != test.expectedLocation {
				t.Errorf("%s returned wrong location: got %s, wanted %s", test.name, actualLocation.String(), test.expectedLocation)
			}
		}

		// Test expected HTML
		if test.expectedHTML != "" {
			// Read the response body
			HTML := recorder.Body.String()

			if !strings.Contains(HTML, test.expectedHTML) {
				t.Errorf("%s expected to find HTML %s but didn't", test.name, test.expectedHTML)
			}
		}
	}
}

// Create a set of tests to run
var postReservationTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/reservation-summary",
	},
	{
		name:                 "missing-post-body",
		postedData:           nil,
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "",
	},
	{
		name: "invalid-start-date",
		postedData: url.Values{
			"start_date": {"invalid"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-end-date",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"invalid"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-room-id",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"invalid"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "invalid-data",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"J"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1"},
		},
		expectedResponseCode: http.StatusOK,
		expectedHTML:         `action="/make-reservation"`,
		expectedLocation:     "",
	},
	{
		name: "DB-insert-fails-reservation",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"2"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
	{
		name: "DB-insert-fails-restriction",
		postedData: url.Values{
			"start_date": {"2050-01-01"},
			"end_date":   {"2050-01-02"},
			"first_name": {"John"},
			"last_name":  {"Smith"},
			"email":      {"john@smith.com"},
			"phone":      {"555-555-5555"},
			"room_id":    {"1000"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedHTML:         "",
		expectedLocation:     "/",
	},
}

// TestPostReservation tests the PostReservation handler for various scenarios.
func TestPostReservation(t *testing.T) {
	for _, test := range postReservationTests {
		var req *http.Request

		// If there is no posted data, want an empty request body
		if test.postedData != nil {
			req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(test.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/make-reservation", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostReservation)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedResponseCode {
			t.Errorf("%s expected code %d, but got %d", test.name, test.expectedResponseCode, recorder.Code)
		}

		// Check location
		if test.expectedLocation != "" {
			// Get the URL from the test
			actualLocation, _ := recorder.Result().Location()
			if actualLocation.String() != test.expectedLocation {
				t.Errorf("%s returned wrong location: got %s, wanted %s", test.name, actualLocation.String(), test.expectedLocation)
			}
		}

		// Check HTML
		if test.expectedHTML != "" {
			// Read the response body
			HTML := recorder.Body.String()
			if !strings.Contains(HTML, test.expectedHTML) {
				t.Errorf("%s expected to find HTML %s but didn't", test.name, test.expectedHTML)
			}
		}
	}
}

// TestNewRepo tests that the NewRepo function returns a Repository type
func TestNewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	// Check type of testRepo
	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *handlers.Repository", reflect.TypeOf(testRepo).String())
	}
}

// Create a set of tests to run
var postAvailabilityTests = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "rooms-not-available",
		postedData: url.Values{
			"start": {"2050-01-01"},
			"end":   {"2050-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "rooms-are-available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "empty-post-body",
		postedData:         url.Values{},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "start-date-wrong-format",
		postedData: url.Values{
			"start":   {"invalid"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "end-date-wrong-format",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"invalid"},
			"room_id": {"1"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name: "DB-query-fails",
		postedData: url.Values{
			"start": {"2060-01-01"},
			"end":   {"2060-01-02"},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}

// TestPostAvailability tests the PostAvailability handler for various scenarios.
func TestPostAvailability(t *testing.T) {
	for _, test := range postAvailabilityTests {
		req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(test.postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostAvailability)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedStatusCode {
			t.Errorf("%s expected code %d, but got %d", test.name, test.expectedStatusCode, recorder.Code)
		}
	}
}

// Create a set of tests to run
var availabilityJSONTests = []struct {
	name            string
	postedData      url.Values
	expectedOK      bool
	expectedMessage string
}{
	{
		name: "rooms-not-available",
		postedData: url.Values{
			"start":   {"2050-01-01"},
			"end":     {"2050-01-02"},
			"room_id": {"1"},
		},
		expectedOK: false,
	},
	{
		name: "rooms-available",
		postedData: url.Values{
			"start":   {"2040-01-01"},
			"end":     {"2040-01-02"},
			"room_id": {"1"},
		},
		expectedOK: true,
	},
	{
		name:            "empty-post-body",
		postedData:      nil,
		expectedOK:      false,
		expectedMessage: "Internal Server Error",
	},
	{
		name: "DB-query-fails",
		postedData: url.Values{
			"start":   {"2060-01-01"},
			"end":     {"2060-01-02"},
			"room_id": {"1"},
		},
		expectedOK:      false,
		expectedMessage: "Error querying database",
	},
}

// TestAvailabilityJSON tests the AvailabilityJSON handler for various scenarios.
func TestAvailabilityJSON(t *testing.T) {
	for _, test := range availabilityJSONTests {
		var req *http.Request

		if test.postedData != nil {
			req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(test.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/search-availability-json", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AvailabilityJSON)
		handler.ServeHTTP(recorder, req)

		var JSONResponse jsonResponse

		err := json.Unmarshal(recorder.Body.Bytes(), &JSONResponse)

		if err != nil {
			t.Errorf("%s failed to parse JSON", test.name)
		}

		// Check OK
		if JSONResponse.Ok != test.expectedOK {
			t.Errorf("%s: expected %v but got %v", test.name, test.expectedOK, JSONResponse.Ok)
		}
	}
}

// Create a set of tests to run
var reservationSummaryTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusOK,
		expectedLocation:   "",
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		url:                "/reservation-summary",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

// TestReservationSummary tests the ReservationSummary handler for various scenarios.
func TestReservationSummary(t *testing.T) {
	for _, test := range reservationSummaryTests {
		req, _ := http.NewRequest("GET", test.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		recorder := httptest.NewRecorder()

		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		handler := http.HandlerFunc(Repo.ReservationSummary)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedStatusCode)
		}

		// Check location
		if test.expectedLocation != "" {
			actualLocation, _ := recorder.Result().Location()

			if actualLocation.String() != test.expectedLocation {
				t.Errorf("failed %s: expected location %s but got location %s", test.name, test.expectedLocation, actualLocation.String())
			}
		}
	}
}

// Create a set of tests to run
var chooseRoomTests = []struct {
	name               string
	reservation        models.Reservation
	url                string
	expectedStatusCode int
	expectedLocation   string
}{
	{
		name: "reservation-in-session",
		reservation: models.Reservation{
			RoomID: 1,
			Room: models.Room{
				ID:       1,
				RoomName: "General's Quarters",
			},
		},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/make-reservation",
	},
	{
		name:               "reservation-not-in-session",
		reservation:        models.Reservation{},
		url:                "/choose-room/1",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
	{
		name:               "malformed-url",
		reservation:        models.Reservation{},
		url:                "/choose-room/bad-url",
		expectedStatusCode: http.StatusSeeOther,
		expectedLocation:   "/",
	},
}

// TestChooseRoom tests the ChooseRoom handler for various scenarios.
func TestChooseRoom(t *testing.T) {
	for _, test := range chooseRoomTests {
		req, _ := http.NewRequest("GET", test.url, nil)
		ctx := getCtx(req)
		if test.name == "malformed-url" {
			ctx = addIdToChiContext(ctx, "bad-url")
		} else {
			ctx = addIdToChiContext(ctx, "1")
		}
		req = req.WithContext(ctx)

		req.RequestURI = test.url // Set RequestURI on request so we can get the ID from the URL

		recorder := httptest.NewRecorder()

		if test.reservation.RoomID > 0 {
			session.Put(ctx, "reservation", test.reservation)
		}

		handler := http.HandlerFunc(Repo.ChooseRoom)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedStatusCode)
		}

		// Check location
		if test.expectedLocation != "" {
			actualLocation, _ := recorder.Result().Location()

			if actualLocation.String() != test.expectedLocation {
				t.Errorf("failed %s: expected location %s but got location %s", test.name, test.expectedLocation, actualLocation.String())
			}
		}
	}
}

// Create a set of tests to run
var bookRoomTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{
		name:               "DB-works",
		url:                "/book-room?s=2050-01-01&e=2050-01-02&id=1",
		expectedStatusCode: http.StatusSeeOther,
	},
	{
		name:               "DB-fails",
		url:                "/book-room?s=2040-01-01&e=2040-01-02&id=4",
		expectedStatusCode: http.StatusSeeOther,
	},
}

// TestBookRoom tests the BookRoom handler for various scenarios.
func TestBookRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	for _, test := range bookRoomTests {
		req, _ := http.NewRequest("GET", test.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		recorder := httptest.NewRecorder()
		session.Put(ctx, "reservation", reservation)
		handler := http.HandlerFunc(Repo.BookRoom)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedStatusCode {
			t.Errorf("%s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedStatusCode)
		}
	}
}

// Struct for login tests
var loginTests = []struct {
	name               string
	email              string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{"valid-credentials", "asd@asd.asd", http.StatusSeeOther, "", "/"},
	{"invalid-credentials", "qwe@qwe.qwe", http.StatusSeeOther, "", "/user/login"},
	{"invalid-data", "a", http.StatusOK, `action="/user/login"`, ""},
}

// TestLogin tests the PostShowLogin handler.
func TestLogin(t *testing.T) {
	for _, test := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", test.email)
		postedData.Add("password", "password")

		req, _ := http.NewRequest("POST", "/user/login", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.PostShowLogin)
		handler.ServeHTTP(recorder, req)

		if recorder.Code != test.expectedStatusCode {
			t.Errorf("Test %s. PostShowLogin handler returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedStatusCode)
		}

		if test.expectedLocation != "" {
			actualLocation, _ := recorder.Result().Location()

			if actualLocation.String() != test.expectedLocation {
				t.Errorf("Test %s. PostShowLogin handler returned wrong response code: got %s, wanted %s", test.name, actualLocation.String(), test.expectedLocation)
			}
		}

		// Check expected values in HTML
		if test.expectedHTML != "" {
			html := recorder.Body.String()

			if !strings.Contains(html, test.expectedHTML) {
				t.Errorf("Test %s. PostShowLogin handler returned wrong HTML: got %q, wanted %q", test.name, html, test.expectedHTML)
			}
		}
	}
}

// Create a set of tests to run
var adminPostShowReservationsTests = []struct {
	name                 string
	url                  string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
}{
	{
		name: "valid-data-from-new",
		url:  "/admin/reservations/new/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"john@doe.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-new",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-all",
		url:  "/admin/reservations/all/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"john@doe.com"},
			"phone":      {"555-555-5555"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-all",
		expectedHTML:         "",
	},
	{
		name: "valid-data-from-cal",
		url:  "/admin/reservations/cal/1/show",
		postedData: url.Values{
			"first_name": {"John"},
			"last_name":  {"Doe"},
			"email":      {"john@doe.com"},
			"phone":      {"555-555-5555"},
			"year":       {"2022"},
			"month":      {"01"},
		},
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "/admin/reservations-calendar?y=2022&m=01",
		expectedHTML:         "",
	},
}

// TestAdminPostShowReservations tests the AdminPostShowReservation handler for various scenarios.
func TestAdminPostShowReservations(t *testing.T) {
	for _, test := range adminPostShowReservationsTests {
		var req *http.Request

		if test.postedData != nil {
			req, _ = http.NewRequest("POST", "/user/login", strings.NewReader(test.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations/new/1/show", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.RequestURI = test.url
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminPostShowReservation)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedResponseCode {
			t.Errorf("Test %s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedResponseCode)
		}

		// Check location
		if test.expectedLocation != "" {
			actualLocation, _ := recorder.Result().Location()

			if actualLocation.String() != test.expectedLocation {
				t.Errorf("Test %s expected a location of %s, but got %s", test.name, test.expectedLocation, actualLocation.String())
			}
		}

		// Check expected values in HTML
		if test.expectedHTML != "" {
			HTML := recorder.Body.String()

			if !strings.Contains(HTML, test.expectedHTML) {
				t.Errorf("Test %s expected to find %s, but didn't", test.name, test.expectedHTML)
			}
		}
	}
}

// Create a set of tests to run
var adminPostReservationCalendarTests = []struct {
	name                 string
	postedData           url.Values
	expectedResponseCode int
	expectedLocation     string
	expectedHTML         string
	blocks               int
	reservations         int
}{
	{
		name: "cal",
		postedData: url.Values{
			"year":  {time.Now().Format("2006")},
			"month": {time.Now().Format("01")},
			fmt.Sprintf("add_block_1_%s", time.Now().AddDate(0, 0, 2).Format("2006-01-2")): {"1"},
		},
		expectedResponseCode: http.StatusSeeOther,
	},
	{
		name:                 "cal-blocks",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		blocks:               1,
	},
	{
		name:                 "cal-reservation",
		postedData:           url.Values{},
		expectedResponseCode: http.StatusSeeOther,
		reservations:         1,
	},
}

// TestPostReservationCalendar tests the PostReservationCalendar handler for various scenarios.
func TestPostReservationCalendar(t *testing.T) {
	for _, test := range adminPostReservationCalendarTests {
		var req *http.Request

		if test.postedData != nil {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", strings.NewReader(test.postedData.Encode()))
		} else {
			req, _ = http.NewRequest("POST", "/admin/reservations-calendar", nil)
		}

		ctx := getCtx(req)
		req = req.WithContext(ctx)

		now := time.Now()
		bm := make(map[string]int)
		rm := make(map[string]int)

		currYear, currMonth, _ := now.Date()
		currLocation := now.Location()

		firstOfMonth := time.Date(currYear, currMonth, 1, 0, 0, 0, 0, currLocation)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		for d := firstOfMonth; !d.After(lastOfMonth); d = d.AddDate(0, 0, 1) {
			rm[d.Format("2006-01-2")] = 0
			bm[d.Format("2006-01-2")] = 0
		}

		if test.blocks > 0 {
			bm[firstOfMonth.Format("2006-01-2")] = test.blocks
		}

		if test.reservations > 0 {
			rm[lastOfMonth.Format("2006-01-2")] = test.reservations
		}

		session.Put(ctx, "block_map_1", bm)
		session.Put(ctx, "reservation_map_1", rm)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminPostReservationCalendar)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedResponseCode {
			t.Errorf("Test %s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedResponseCode)
		}
	}
}

// Create a set of tests to run
var adminProcessReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "process-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "process-reservation-back-to-calendar",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

// TestAdminProcessReservation tests the AdminProcessReservation handler for various scenarios.
func TestAdminProcessReservation(t *testing.T) {
	for _, test := range adminProcessReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", test.queryParams), nil)

		ctx := getCtx(req)
		req = req.WithContext(ctx)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(Repo.AdminProcessReservation)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedResponseCode {
			t.Errorf("Test %s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedResponseCode)
		}
	}
}

// Create a set of tests to run
var adminDeleteReservationTests = []struct {
	name                 string
	queryParams          string
	expectedResponseCode int
	expectedLocation     string
}{
	{
		name:                 "delete-reservation",
		queryParams:          "",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
	{
		name:                 "delete-reservation-back-to-calendar",
		queryParams:          "?y=2021&m=12",
		expectedResponseCode: http.StatusSeeOther,
		expectedLocation:     "",
	},
}

// TestAdminDeleteReservation tests the AdminDeleteReservation handler for various scenarios.
func TestAdminDeleteReservation(t *testing.T) {
	for _, test := range adminDeleteReservationTests {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/admin/process-reservation/cal/1/do%s", test.queryParams), nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(Repo.AdminDeleteReservation)
		handler.ServeHTTP(recorder, req)

		// Check status code
		if recorder.Code != test.expectedResponseCode {
			t.Errorf("Test %s returned wrong response code: got %d, wanted %d", test.name, recorder.Code, test.expectedResponseCode)
		}
	}
}

// addIdToChiContext adds an ID to the chi route context within the provided context.
// It returns a new context with the chi route context containing the ID as a URL parameter.
func addIdToChiContext(ctx context.Context, id string) context.Context {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", id)
	return context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
}

// getCtx retrieves the context from the session using the X-Session header. It
// creates a new context if there is an error loading the context from the session.
func getCtx(req *http.Request) context.Context {

	// Get the context from the session using X-Session header
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
