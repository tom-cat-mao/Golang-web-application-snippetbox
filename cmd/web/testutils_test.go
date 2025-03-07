package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

// Create a newTestApplication helper which returns an instance of out
// application struct containing mocked dependencies.
func newTestApplication(t *testing.T) *application {
	return &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
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
