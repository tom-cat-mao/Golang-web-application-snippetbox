package main

import (
	"fmt"
	"net/http"
)

// commonHeaders middleware sets various security-related headers on outgoing HTTP responses.
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content-Security-Policy header to enhance security by restricting the sources of content that can be loaded on the page.
		// This policy:
		//   - Allows loading of content only from the same origin ('self')
		//   - Permits styles from the same origin and fonts.googleapis.com
		//   - Allows fonts from fonts.gstatic.com
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com",
		)

		// Set Referrer-Policy header to control the amount of referrer information sent with navigation requests.
		// The 'origin-when-cross-origin' policy:
		//   - Sends the full URL when navigating within the same origin
		//   - Sends only the origin when navigating to a different origin
		w.Header().Set(
			"Referrer-Policy",
			"origin-when-cross-origin",
		)

		// Set X-Content-Type-Options header to prevent MIME type sniffing attacks.
		// The 'nosniff' value instructs the browser to respect the declared content type and not to sniff the content.
		w.Header().Set(
			"X-Content-Type-Options",
			"nosniff",
		)

		// Set X-Frame-Options header to prevent clickjacking attacks.
		// The 'deny' value disallows the page from being framed, enhancing security by preventing embedding in iframes.
		w.Header().Set(
			"X-Frame-Options",
			"deny",
		)

		// Set X-XSS-Protection header to disable the browser's built-in XSS filters.
		// The '0' value is used for compatibility with older browsers, ensuring consistent behavior across different environments.
		w.Header().Set(
			"X-XSS-Protection",
			"0",
		)

		// Set Server header to mask the Go server version, enhancing security and privacy.
		// By setting it to 'Go', we obscure the exact version of the server, reducing the attack surface.
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
		// Extract details from the incoming request for logging purposes.
		// The following variables are captured:
		//   - ip: The IP address of the client making the request
		//   - proto: The protocol used for the request (e.g., "HTTP/1.1")
		//   - method: The HTTP method used (e.g., "GET", "POST")
		//   - uri: The request URI path and query string
		var (
			ip     = r.RemoteAddr       // IP address of the client making the request
			proto  = r.Proto            // Protocol used for the request (e.g., "HTTP/1.1")
			method = r.Method           // HTTP method used (e.g., "GET", "POST")
			uri    = r.URL.RequestURI() // Request URI path and query string
		)

		// Log details of the received request using the application's logger.
		// This log entry includes:
		//   - The client's IP address
		//   - The protocol used
		//   - The HTTP method
		//   - The request URI
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
				// Set the Connection header to 'close' to ensure the connection is closed immediately after sending the response.
				// This is done to prevent further requests on the same connection after a panic occurs.
				w.Header().Set("Connection", "close")

				// Log the error using the application's logger to record the panic details.
				// The error is wrapped in a new error to provide context about the panic.
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		// Call the next handler in the chain with the modified response writer and request
		next.ServeHTTP(w, r)
	})
}

// use middleware to control if someone try to create a script without authentication
// if access to the create page without authentication
// it will be redirected to the login page
// outherwise go to the next handler and clean the cache
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: nostore" so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache)
		w.Header().Add("Cache-Control", "nostore")

		// And call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
