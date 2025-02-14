package main

import "net/http"

func (app *application) routes() http.Handler {
	// Create a new HTTP request multiplexer.
	mux := http.NewServeMux()

	// Set up a file server to serve static files from the `./ui/static/` directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Handle requests for static resources and strip the "/static" prefix from the URL path.
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Define routes and associate them with handler functions.

	// Route for the home page.
	mux.HandleFunc("GET /{$}", app.home)

	// Route to view a specific snippet by its ID.
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)

	// Route to display the form for creating a new snippet.
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)

	// Route to handle POST requests for submitting the new snippet form.
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	// Apply the commonHeaders middleware to all routes.
	return commonHeaders(mux)
}
