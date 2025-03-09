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

func TestSnippetView(t *testing.T) {
	// Create a new instance of our application struct with a mocked logger
	app := newTestApplication(t)

	// Create a test server
	ts := newTestServer(t, app.routes())
	defer ts.server.Close()

	// Test cases
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/2",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Negative ID",
			urlPath:  "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/view/1.23",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)

			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}
