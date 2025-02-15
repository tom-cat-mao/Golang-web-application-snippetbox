package main

import (
	"fmt"
	"net/http"
)

// commonHeaders middleware sets various security-related headers on outgoing HTTP responses.
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Security-Policy header for security.
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		)

		// Set Referrer-Policy header to control how much referrer information is included in navigation requests.
		w.Header().Set(
			"Referrer-Policy",
			"origin-when-cross-origin",
		)

		// Set X-Content-Type-Options to prevent MIME type sniffing attacks.
		w.Header().Set(
			"X-Content-Type-Options",
			"nosniff",
		)

		// Set X-Frame-Options to prevent clickjacking attacks by disallowing the page from being framed.
		w.Header().Set(
			"X-Frame-Options",
			"deny",
		)

		// Set X-XSS-Protection to disable browser's built-in XSS filters for compatibility with older browsers.
		w.Header().Set(
			"X-XSS-Protection",
			"0",
		)

		// Set Server header to hide the Go server version, enhancing security and privacy by not revealing the exact technology stack.
		w.Header().Set(
			"Server",
			"Go",
		)

		// Call the next handler in the chain with the modified response writer and request.
		next.ServeHTTP(w, r)
	})
}

// logRequest middleware logs details about each HTTP request received by the server.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract details from the incoming request
		var (
			ip     = r.RemoteAddr       // IP address of the client making the request
			proto  = r.Proto            // Protocol used for the request (e.g., "HTTP/1.1")
			method = r.Method           // HTTP method used (e.g., "GET", "POST")
			uri    = r.URL.RequestURI() // Request URI path and query string
		)

		// Log details of the received request
		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		// Call the next handler in the chain with the modified response writer and request
		next.ServeHTTP(w, r)
	})
}

// recoverPanic middleware recovers from any panic that occurs during the processing of a request.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Set the connection to close immediately after sending the response
				w.Header().Set("Connection", "close")

				// Log the error using the application's logger
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		// Call the next handler in the chain with the modified response writer and request
		next.ServeHTTP(w, r)
	})
}
