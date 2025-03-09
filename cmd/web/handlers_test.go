package main

import (
	"html"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"testing"

	"snippetbox.tomcat.net/internal/assert"
)

// Define a regular expression which captures the CSRF token value from
// the HTML for our user signup page.
var csrfTokenRX = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)

// extractCSRFToken is a helper function that extracts a CSRF token from
// the given HTML body using a regular expression. It takes a testing.T
// instance and the HTML body as input, and returns the extracted token
// as a string. If no token is found, the test will be marked as failed
// and execution will stop via t.Fatal().
func extractCSRFToken(t *testing.T, body string) string {
	// Use the FindStringSubmatch method to extract the token from the HTML body
	// Note that this returns an array with the entire matched pattern in the
	// first position, and the values of any captured data in the subsequent
	// positions.
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(matches[1])
}

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

// TestUserSignup tests the user signup handler.
// It verifies that the signup form is returned with a valid CSRF token,
// which is required for form submission. The test:
// 1. Creates a new test application and server
// 2. Makes GET request to /user/signup
// 3. Extracts CSRF token from response body
// 4. Logs token for debugging purposes (but doesn't validate it here)
func TestUserSignup(t *testing.T) {
	// Create the application struct containing our mocked dependencies and set
	// up the test server for running an end-to-end test.
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.server.Close()

	// Make a GET /user/signup request and then extract the CSRF token from the
	// response body
	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	// Log the CSRF token value in our test output using the t.Logf() function
	// The t.Logf() function works in the same way as fmt.Printf(), but writes
	// the provided message to the test output.
	t.Logf("CSRF token is: %q", csrfToken)
}
