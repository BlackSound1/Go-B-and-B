package main

import (
	"testing"
)

// TestRun tests the run function which sets up the app
// config and starts the web server.
func TestRun(t *testing.T) {
	_, err := run()

	if err != nil {
		t.Error("Failed run()")
	}
}

func TestCreateNewServer(t *testing.T) {
	serv := createNewServer()

	// Test server is not nil
	if serv == nil {
		t.Errorf("expected server to be not nil, but got nil")
		return
	}

	// Test server address is set to portNumber
	if serv.Addr != portNumber {
		t.Errorf("expected server address to be %s, but got %s", portNumber, serv.Addr)
	}

	// Test server handler is not nil
	if serv.Handler == nil {
		t.Errorf("expected server handler to be not nil, but got nil")
	}
}
