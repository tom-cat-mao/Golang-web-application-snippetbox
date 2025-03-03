package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Create a new ServeMux, which is an HTTP request multiplexer. This will
	// match incoming requests against registered routes and dispatch them to the
	// appropriate handler function.
	mux := http.NewServeMux()

	// Create a file server which serves static files (CSS, JS, images etc.) from the ./ui/static/ directory.
	// Notice the "ui/static" path is relative to the current working directory.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Register the file server to handle GET requests that start with "/static/".
	// For matching requests, strip the "/static" prefix before the file server looks for
	// the file to serve.
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// Unprotected application routes using the "dynamic" middleware chain.
	// Create a middleware chain containing the session management middleware.
	// Specifically, this will:
	// - LoadAndSave session data for the current request.
	// - noSurf function middleware to protect from CSRD attack
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)

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
