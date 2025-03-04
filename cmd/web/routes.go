package main

import (
	"net/http"

	"github.com/justinas/alice"
	"snippetbox.tomcat.net/ui"
)

func (app *application) routes() http.Handler {
	// Create a new ServeMux, which is an HTTP request multiplexer. This will
	// match incoming requests against registered routes and dispatch them to the
	// appropriate handler function.
	mux := http.NewServeMux()

	// Create a file server which serves static files (CSS, JS, images etc.) from the ./ui/static/ directory.
	// Notice the "ui/static" path is relative to the current working directory.
	// fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Use the http.FileServerFS() function to create a HTTP handler which
	// serves the embedded files in ui.Files. It's important to note that our
	// static files are contained in the "static" folder of the ui.Files
	// embedded filesystem.
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Unprotected application routes using the "dynamic" middleware chain.
	// Create a middleware chain containing the session management middleware.
	// Specifically, this will:
	// - LoadAndSave session data for the current request.
	// - noSurf function middleware to protect from CSRD attack
	// - app.authenticate check if it is authenticated
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// Register the dynamic routes (those that require session management) using
	// the dynamic middleware chain.

	// Home page
	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))

	// View a specific snippet
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))

	// User signup routes
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))

	// User login routes
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected (authenticated-only) application routes, using a new "protected"
	// middleware chain which includes the requireAuthentication middleware.
	protected := dynamic.Append(app.requireAuthentication)

	// Create a new snippet form
	mux.Handle("GET /snippet/create", protected.ThenFunc(app.snippetCreate))

	// Post a new snippet
	mux.Handle("POST /snippet/create", protected.ThenFunc(app.snippetCreatePost))

	// User logout route
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))

	// Create a middleware chain containing our 'standard' middleware
	// which will be applied to every request our application receives.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	// Wrap the servemux with the standard middleware chain. So any HTTP
	// requests coming in will be subject to the middleware chain before being
	// passed to the servemux.
	return standard.Then(mux)
}
