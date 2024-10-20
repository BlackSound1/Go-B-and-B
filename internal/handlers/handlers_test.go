package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"post-search-availability", "/search-availability", "POST", []postData{
		{key: "start", value: "2022-01-01"},
		{key: "end", value: "2022-01-02"},
	}, http.StatusOK},
	{"post-search-availability-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2022-01-01"},
		{key: "end", value: "2022-01-02"},
	}, http.StatusOK},
	{"make-reservation-post", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Smith"},
		{key: "email", value: "example@example.com"},
		{key: "phone", value: "555-555-5555"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes() // Get the routes to test

	testServer := httptest.NewTLSServer(routes) // Create a new test server

	defer testServer.Close() // Close the test server whenever done

	// Loop through the tests
	for _, test := range theTests {

		if test.method == "GET" {
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
		} else {
			// Test all POST requests

			values := url.Values{}

			// Populate values with everything necessary to make a POST
			for _, x := range test.params {
				values.Add(x.key, x.value)
			}

			baseURL := testServer.URL

			// Use the test server client to send a POST request and get the response
			resp, err := testServer.Client().PostForm(baseURL+test.url, values)

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
}
