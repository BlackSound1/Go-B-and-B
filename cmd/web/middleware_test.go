package main

import (
	"net/http"
	"testing"
)

// TestNoSurf tests the NoSurf middleware function to ensure it returns
// a handler that implements the http.Handler interface when a dummy handler
// is passed in. It verifies that the returned handler is of the correct type.
func TestNoSurf(t *testing.T) {
	// Create a dummy Handler
	var myH myHandler

	// Create a NoSurf handler, passing in the dummy handler
	h := NoSurf(&myH)

	// Check that the handler is of type http.Handler
	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}

// TestSessionLoad tests the SessionLoad middleware function to ensure it returns
// a handler that implements the http.Handler interface when a dummy handler
// is passed in.
func TestSessionLoad(t *testing.T) {
	// Create a dummy Handler
	var myH myHandler

	// Create a SessionLoad handler, passing in the dummy handler
	h := SessionLoad(&myH)

	// Check that the handler is of type http.Handler
	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}
