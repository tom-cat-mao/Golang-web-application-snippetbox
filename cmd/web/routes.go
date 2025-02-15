package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Create a new HTTP request multiplexer.
	mux := http.NewServeMux()

	// Set up a file server to serve static files from the `./ui/static/` directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Handle requests for static resources and strip the "/static" prefix from the URL path.
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Define routes and associate them with handler functions.

	// Route for the home page.
	// This route handles GET requests to the root path ("/") and is associated with the `home` method of the application's `application` struct.
	mux.HandleFunc("GET /{$}", app.home)

	// This route handles GET requests to "/snippet/view/{id}" where "{id}" is a dynamic segment representing the snippet's unique identifier. It is associated with the `snippetView` method of the application's `application` struct.
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)

	// This route handles GET requests to "/snippet/create" and is associated with the `snippetCreate` method of the application's `application` struct.
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)

	// This route handles POST requests to "/snippet/create" and is associated with the `snippetCreatePost` method of the application's `application` struct.
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	// Create a middleware chain using alice that includes:
	// 1. recoverPanic: Recovers from any panic that occurs during the processing of a request.
	// 2. logRequest: Logs details about each HTTP request.
	// 3. commonHeaders: Adds security-related headers to outgoing responses.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	// Apply the middleware chain to the multiplexer.
	return standard.Then(mux)
}
