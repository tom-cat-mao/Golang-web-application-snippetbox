package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"snippetbox.tomcat.net/internal/models/mocks"
)

// newTestApplication initializes an application instance for testing, injecting
// mocked dependencies (SnippetModel and UserModel) to isolate tests from the real database layer.
func newTestApplication(t *testing.T) *application {
	// Initialize template cache to avoid repeated parsing of templates during tests
	// This helps prevent 'template not found' errors and improves test performance
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	// Initialize form decoder to handle form submissions in tests
	// Required for parsing URL-encoded form data from POST requests
	formDecoder := form.NewDecoder()

	// Configure session manager with test-appropriate settings
	// Uses secure cookies and a 12-hour lifetime to match production-like behavior
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		snippets:       &mocks.SnippetModel{}, // Now compatible via interface
		users:          &mocks.UserModel{},    // Now compatible via interface
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

// Define a custom testServer type which embeds a httptest.Server instance.
type testServer struct {
	server *httptest.Server
}

// Create a newTestServer helper which initializes and returns a new httptest.Server
// of our custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	// Initialize a new httptest.Server instance. We use the httptest.NewTLSServer()
	// function, which is similar to httptest.NewServer() but also implicitly
	// creates a TLS certificate and key that are used when serving HTTPS requests.
	ts := httptest.NewTLSServer(h)

	// Initialize a new cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client. Any response cookies will
	// now be stored and sent with subsequent requests when using this client
	ts.Client().Jar = jar

	// Disable redirect following. This means our client will not follow
	// redirects from the test server.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// Implement a get() method on our custom testServer type. This makes
// a GET request to a given url path on the test server, and returns the response
// body, status code and server error.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.server.Client().Get(ts.server.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

// Create a postForm method for sending POST requests to the test server. The
// final parameter to this method is a url.Values object which can contain any
// form data that you want to send in the request body
func (ts *testServer) postForm(t *testing.T, usrlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.server.Client().PostForm(ts.server.URL+usrlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body from the test server
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	// Return the response status, headers and body
	return rs.StatusCode, rs.Header, string(body)
}
