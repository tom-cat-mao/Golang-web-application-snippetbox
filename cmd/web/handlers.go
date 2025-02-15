// Package main implements the web handlers for the SnippetBox application.
// It contains HTTP handlers that process web requests and generate responses.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.tomcat.net/internal/models"
)

// home handles the display of the homepage.
// It retrieves the 5 most recent snippets from the database and renders them using
// the home page template.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// The handler follows these steps:
// 1. Fetches latest snippets from the database
// 2. Creates template data with the snippets
// 3. Renders the home page template with the data
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest 5 snippets from the database
	snippets, err := app.snippets.Latest()
	if err != nil {
		// If there's an error fetching snippets, return a 500 Internal Server Error
		app.serverError(w, r, err)
		return
	}

	// Create a template data structure with default values
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Render the "home.html" template with the provided data
	app.render(w, r, http.StatusOK, "home.html", data)
}

// snippetView handles requests to view a specific snippet.
// It accepts a snippet ID as a URL parameter and displays the corresponding snippet.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// URL Parameters:
//   - id: The unique identifier of the snippet to view
//
// The handler follows these steps:
// 1. Extracts and validates the snippet ID from the URL
// 2. Retrieves the snippet from the database
// 3. If snippet not found, returns 404 Not Found
// 4. If found, renders the snippet using the view template
//
// Error handling:
//   - Invalid ID format: Returns 404 Not Found
//   - Snippet not found: Returns 404 Not Found
//   - Database errors: Returns 500 Internal Server Error
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		// If the id parameter is invalid or conversion fails, return 404 Not Found
		http.NotFound(w, r)
		return
	}

	// Get the snippet record from the database using the provided id
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// If the snippet is not found, return 404 Not Found
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			// For other errors, return 500 Internal Server Error
			app.serverError(w, r, err)
		}
		return
	}

	// Create a new template data structure and set the snippet
	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Render the "view.html" template with the provided data
	app.render(w, r, http.StatusOK, "view.html", data)
}

// snippetCreate handles requests to display the create snippet form.
// It renders an HTML form that allows users to input new snippet information.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// The handler follows these steps:
// 1. Creates a new template data structure
// 2. Renders the create form template
//
// This is a GET handler that displays the form. The actual snippet creation
// is handled by snippetCreatePost.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data structure
	data := app.newTemplateData(r)

	// Render the "create.html" template with the provided data
	app.render(w, r, http.StatusOK, "create.html", data)
}

// snippetCreatePost handles POST requests to create a new snippet.
// It processes the form submission, validates the data, and creates a new snippet
// in the database.
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// Form Data:
//   - title: The snippet title (required, max length 100)
//   - content: The snippet content (required, min length 10)
//   - expires: Number of days until the snippet expires (required, must be 1, 7 or 365)
//
// The handler follows these steps:
// 1. Parse form data and validate required fields
// 2. Convert expires value to integer
// 3. Validate expires value is one of allowed values (1, 7, 365)
// 4. Insert the new snippet into the database
// 5. Redirect to the view page for the new snippet
//
// Error handling:
//   - Invalid form data: Returns 400 Bad Request
//   - Database insertion errors: Returns 500 Internal Server Error
//   - Successful creation: Redirects to /snippet/view/{id} with 303 See Other
//
// Returns:
//   - On success: HTTP 303 redirect to new snippet's view page
//   - On error: Appropriate HTTP error status code
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Set the title, content, and expiration time for the new snippet
	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Insert the new snippet into the database
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		// If insertion fails, return 500 Internal Server Error
		app.serverError(w, r, err)
		return
	}

	// Redirect to the new snippet's view page with 303 See Other status code
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
