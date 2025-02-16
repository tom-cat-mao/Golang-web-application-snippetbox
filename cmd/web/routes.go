package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Create a new HTTP request multiplexer (router) that will match incoming
	// requests against registered routes and dispatch them to the appropriate
	// handler functions.
	mux := http.NewServeMux()

	// Set up a file server to serve static files (CSS, JS, images) from the
	// ./ui/static/ directory. The file server will handle requests for static
	// resources like /static/css/main.css.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Register the file server to handle GET requests starting with /static/
	// and strip the /static prefix before serving the files.
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Create a middleware chain for routes that require session management.
	// The sessionManager.LoadAndSave middleware automatically loads and saves
	// session data for each request.
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Register application routes with their corresponding handler functions:

	// Home page route - handles GET requests to the root URL ("/")
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))

	// Snippet view route - handles GET requests to view individual snippets
	// The {id} is a dynamic URL parameter containing the snippet ID
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))

	// Snippet creation form route - handles GET requests to display the
	// snippet creation form
	mux.Handle("GET /snippet/create", dynamic.ThenFunc(app.snippetCreate))

	// Snippet creation submission route - handles POST requests to process
	// the snippet creation form submission
	mux.Handle("POST /snippet/create", dynamic.ThenFunc(app.snippetCreatePost))

	// Create a standard middleware chain that will be applied to all requests:
	// 1. recoverPanic - recovers from panics and returns a 500 error
	// 2. logRequest - logs details about each incoming request
	// 3. commonHeaders - adds security-related headers to responses
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	// Apply the standard middleware chain to the multiplexer and return
	// the configured router.
	return standard.Then(mux)
}
