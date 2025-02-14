package main

import "net/http"

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
