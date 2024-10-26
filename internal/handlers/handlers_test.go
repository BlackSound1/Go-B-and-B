package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/BlackSound1/Go-B-and-B/internal/driver"
	"github.com/BlackSound1/Go-B-and-B/internal/models"
	"github.com/go-chi/chi"
)

// Create a struct to define what kinds of values we expect
type postData struct {
	key   string
	value string
}

// Create a set of tests to run
var theTests = []struct {
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

	// {"post-search-availability", "/search-availability", "POST", []postData{
	// 	{key: "start", value: "2022-01-01"},
	// 	{key: "end", value: "2022-01-02"},
	// }, http.StatusOK},
	// {"post-search-availability-json", "/search-availability-json", "POST", []postData{
	// 	{key: "start", value: "2022-01-01"},
	// 	{key: "end", value: "2022-01-02"},
	// }, http.StatusOK},
	// {"make-reservation-post", "/make-reservation", "POST", []postData{
	// 	{key: "first_name", value: "John"},
	// 	{key: "last_name", value: "Smith"},
	// 	{key: "email", value: "example@example.com"},
	// 	{key: "phone", value: "555-555-5555"},
	// }, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes() // Get the routes to test

	testServer := httptest.NewTLSServer(routes) // Create a new test server

	defer testServer.Close() // Close the test server whenever done

	// Loop through the tests
	for _, test := range theTests {

		// Test all GET requests

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

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)

	// Get the context
	ctx := getCtx(req)

	// Add context to the request
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	// Put the reservation into the session using the context
	session.Put(ctx, "reservation", reservation)

	// Take Reservation function and turn it into a handler
	handler := http.HandlerFunc(Repo.Reservation)

	// Call the handler. Response Recorder is the response writer
	handler.ServeHTTP(recorder, req)

	// Check that the response code is 200
	if recorder.Code != http.StatusOK {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusOK)
	}

	// Test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test with non-existant room
	req, _ = http.NewRequest("GET", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	recorder = httptest.NewRecorder()
	reservation.RoomID = 100                     // Make sure it does not exist
	session.Put(ctx, "reservation", reservation) // Put reservation into session
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("Reservation handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_PostReservation(t *testing.T) {
	reqBody := "start_date=2050-01-01&end_date=2050-01-02&email=John@Smith.com&phone=123456789&room_id=1"
	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusSeeOther)
	}

	// Test for missing post body
	req, _ = http.NewRequest("POST", "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// test for invalid start date
	reqBody = "start_date=invalid&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com&phone=123456789&room_id=1"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid start date: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid end date
	reqBody = "start_date=2050-01-01&end_date=invalid&first_name=John&last_name=Smith&email=john@smith.com&phone=123456789&room_id=1"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid end date: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid room id
	reqBody = "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com&phone=123456789&room_id=invalid"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for invalid room id: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid data
	reqBody = "start_date=2050-01-01&end_date=2050-01-02&first_name=J&last_name=Smith&email=john@smith.com&phone=123456789&room_id=1"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for invalid data: got %d, wanted %d", recorder.Code, http.StatusSeeOther)
	}

	// Test for failure to insert reservation into database
	reqBody = "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com&phone=123456789&room_id=2"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to fail inserting reservation: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test for failure to insert restriction into database
	reqBody = "start_date=2050-01-01&end_date=2050-01-02&first_name=John&last_name=Smith&email=john@smith.com&phone=123456789&room_id=1000"
	req, _ = http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostReservation)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler failed when trying to fail inserting reservation: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_NewRepo(t *testing.T) {
	var db driver.DB
	testRepo := NewRepo(&app, &db)

	// Check type of testRepo
	if reflect.TypeOf(testRepo).String() != "*handlers.Repository" {
		t.Errorf("Did not get correct type from NewRepo: got %s, wanted *handlers.Repository", reflect.TypeOf(testRepo).String())
	}
}

func TestRepository_PostAvailability(t *testing.T) {
	// Test case where rooms not available

	reqBody := "start=2050-01-01&end=2050-01-02"
	req, _ := http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusSeeOther)
	}

	// Test case where rooms are available

	reqBody = "start=2040-01-01&end=2040-01-02"
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusOK)
	}

	// Test empty post body

	req, _ = http.NewRequest("POST", "/search-availability", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test bad start date

	reqBody = "start=invalid&end=2050-01-02"
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test bad end date

	reqBody = "start=2040-01-01&end=invalid"
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test bad DB query

	reqBody = "start=2060-01-01&end=2060-01-02" // Should cause bad db call
	req, _ = http.NewRequest("POST", "/search-availability", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.PostAvailability)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostAvailability handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_AvailabilityJSON(t *testing.T) {
	// Test case where rooms not available

	// Create request body
	reqBody := "start=2050-01-01&end=2050-01-02&room_id=1"

	// Create request
	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	// Get the context with session
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	// Set the request header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create response recorder
	recorder := httptest.NewRecorder()

	// Make the handler an http.HandlerFunc
	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	// Make the request to the handler
	handler.ServeHTTP(recorder, req)

	// Since no rooms available, expect http.StatusSeeOther.
	// This time, want to parse JSON and get the expected response
	var j jsonResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// Since start date > 2049-12-31, expect no availability
	if j.Ok {
		t.Error("Got availability when none was expected in AvailabilityJSON")
	}

	// Test case where rooms are available

	reqBody = "start=2040-01-01&end=2040-01-02&room_id=1"
	req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.AvailabilityJSON)
	handler.ServeHTTP(recorder, req)
	err = json.Unmarshal(recorder.Body.Bytes(), &j)
	if err != nil {
		t.Error("failed to parse json!")
	}

	// Since start date < 2049-12-31, expect availability
	if !j.Ok {
		t.Error("Got no availability when some was expected in AvailabilityJSON")
	}

	// Test case where no request body

	// req, _ = http.NewRequest("POST", "/search-availability-json", nil)
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// recorder = httptest.NewRecorder()
	// handler = http.HandlerFunc(Repo.AvailabilityJSON)
	// handler.ServeHTTP(recorder, req)

	// err = json.Unmarshal(recorder.Body.Bytes(), &j)
	// if err != nil {
	// 	t.Error("failed to parse json!")
	// }

	// if j.Ok || j.Message != "Internal server error" {
	// 	t.Error("Got availability when request body was empty")
	// }

	// Test case where there is a database error

	// reqBody = "start=2060-01-01"
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "end=2060-01-02")
	// reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")
	// req, _ = http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))
	// ctx = getCtx(req)
	// req = req.WithContext(ctx)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// recorder = httptest.NewRecorder()
	// handler = http.HandlerFunc(Repo.AvailabilityJSON)
	// handler.ServeHTTP(recorder, req)
	// err = json.Unmarshal(recorder.Body.Bytes(), &j)
	// if err != nil {
	// 	t.Error("failed to parse json!")
	// }

	// // since we specified a start date < 2049-12-31, we expect availability
	// if j.Ok || j.Message != "Error querying database" {
	// 	t.Error("Got availability when simulating database error")
	// }

}

func TestRepository_ReservationSummary(t *testing.T) {

	// Test case where the reservation is in the session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/reservation-summary", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusOK)
	}

	// Test case where reservation is not in session
	req, _ = http.NewRequest("GET", "/reservation-summary", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ReservationSummary)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ReservationSummary handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_ChooseRoom(t *testing.T) {
	// Test case where reservation is in the session
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/choose-room/1", nil)
	ctx := getCtx(req)
	ctx = addIdToChiContext(ctx, "1")
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"
	recorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusSeeOther)
	}

	// Test case where reservation is not in the session
	req, _ = http.NewRequest("GET", "/choose-room/1", nil)
	ctx = getCtx(req)
	ctx = addIdToChiContext(ctx, "1")
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/1"
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}

	// Test case where URL paramter is missing
	req, _ = http.NewRequest("GET", "/choose-room/fish", nil)
	ctx = getCtx(req)
	ctx = addIdToChiContext(ctx, "fish")
	req = req.WithContext(ctx)
	req.RequestURI = "/choose-room/fish"
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("ChooseRoom handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
	}
}

func TestRepository_BookRoom(t *testing.T) {
	// Test case where the database works

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/book-room?s=2050-01-01&e=2050-01-02&id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	recorder := httptest.NewRecorder()
	session.Put(ctx, "reservation", reservation)
	handler := http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusSeeOther)
	}

	// Test case where the database doesn't work

	req, _ = http.NewRequest("GET", "/book-room?s=2040-01-01&e=2040-01-02&id=4", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	recorder = httptest.NewRecorder()
	handler = http.HandlerFunc(Repo.BookRoom)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Errorf("BookRoom handler returned wrong response code: got %d, wanted %d", recorder.Code, http.StatusTemporaryRedirect)
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
