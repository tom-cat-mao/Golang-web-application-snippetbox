package main

import (
	"io"
	"log/slog"
	"net/http"
	"testing"

	"snippetbox.tomcat.net/internal/assert"
)

func TestPing(t *testing.T) {
	// Create a new instance of our application struct.
	// We initialize it with a logger that discards output (io.Discard)
	// to prevent test logs from cluttering the test output.
	app := &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Create a new test server which uses our application routes.
	// The test server automatically handles HTTPS requests and provides
	// a Client that can make requests to it.
	ts := newTestServer(t, app.routes())
	defer ts.server.Close()

	// Make a GET request to the /ping endpoint using our test server.
	// We then check that the response status code is 200 and the body is "OK".
	code, _, body := ts.get(t, "/ping")
	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}
