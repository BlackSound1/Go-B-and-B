package main

import (
	"net/http"
	"os"
	"testing"
)

// TestMain sets up the testing environment and runs the tests. It is the
// entrypoint for the testing framework.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// myHandler is a dummy handler for testing
type myHandler struct{}

// ServeHTTP only exists to satisfy the http.Handler interface
func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
