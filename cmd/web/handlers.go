// Package web implements the HTTP handlers for the SnippetBox application.
//
// This package contains all the request handlers that process incoming HTTP requests
// and generate appropriate responses. It handles routing, request processing,
// data validation, and template rendering for the web application.
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"snippetbox.tomcat.net/internal/models"
)

// snippetCreateForm encapsulates all data and validation state for the snippet creation form.
// It handles form submission, validation, and error reporting.
//
// Fields:
//   - Title: The snippet title (required, max 100 characters)
//     Validation: Must be non-empty and ≤ 100 characters
//   - Content: The snippet content (required)
//     Validation: Must be non-empty
//   - Expires: Number of days until the snippet expires
//     Validation: Must be 1, 7 or 365
//   - FieldErrors: Map containing validation error messages for each field
//     Key: Field name (e.g., "title", "content", "expires")
//     Value: Error message for the field
type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

// home handles GET requests to the application's homepage (/).
// It retrieves the 5 most recent snippets from the database and renders them
// using the home page template.
//
// Security: No authentication required
// Caching: No caching headers set
// Performance: Database query for latest snippets
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

// snippetView handles GET requests to view a specific snippet (/snippet/view/{id}).
// It accepts a snippet ID as a URL parameter, retrieves the corresponding snippet
// from the database, and renders it using the view template.
//
// Security: No authentication required
// Caching: No caching headers set
// Performance: Single database query by primary key
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

// snippetCreate handles GET requests to display the snippet creation form (/snippet/create).
// It renders an HTML form that allows users to input new snippet information.
// This handler only displays the form - the actual creation is handled by
// snippetCreatePost.
//
// Security: No authentication required
// Caching: No caching headers set
// Performance: No database queries
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

	data.Form = snippetCreateForm{
		Expires: 365,
	}

	// Render the "create.html" template with the provided data
	app.render(w, r, http.StatusOK, "create.html", data)
}

// snippetCreatePost handles POST requests to create a new snippet (/snippet/create).
// It processes the form submission, validates the input data, and creates
// a new snippet record in the database.
//
// Security: No authentication required
// Caching: No caching headers set
// Performance: Single database insert operation
// Concurrency: Safe for concurrent use
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
// 1. Parse form data
// 2. Validate required fields:
//   - title: cannot be blank, max 100 characters
//   - content: cannot be blank
//   - expires: must be 1, 7 or 365
//
// 3. Collect field validation errors in fieldErrors map
// 4. If validation errors exist, return error messages
// 5. Insert the new snippet into the database
// 6. Redirect to the view page for the new snippet
//
// Error handling:
//   - Invalid form data: Returns 400 Bad Request with field error messages
//   - Database insertion errors: Returns 500 Internal Server Error
//   - Successful creation: Redirects to /snippet/view/{id} with 303 See Other
//
// Returns:
//   - On success: HTTP 303 redirect to new snippet's view page
//   - On validation error: HTTP 400 with field error messages
//   - On database error: HTTP 500 Internal Server Error
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Parse form data from the request
	// Handles both URL-encoded and multipart form data
	err := r.ParseForm()
	if err != nil {
		// Return 400 Bad Request if form parsing fails
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Get and convert expires value from form
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		// Return 400 Bad Request if expires value is invalid
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Create form struct with submitted values
	form := snippetCreateForm{
		Title:       r.PostFormValue("title"),
		Content:     r.PostFormValue("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}

	// Validate title field
	if strings.TrimSpace(form.Title) == "" {
		// Add error if title is empty
		form.FieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		// Add error if title exceeds 100 characters
		form.FieldErrors["title"] = "This field cannot be more than 100 characters long"
	}

	// Validate content field
	if strings.TrimSpace(form.Content) == "" {
		// Add error if content is empty
		form.FieldErrors["content"] = "This field cannot be blank"
	}

	// Validate expires field
	if expires != 1 && expires != 7 && expires != 365 {
		// Add error if expires value is invalid
		form.FieldErrors["expires"] = "This field must equal 1, 7 or 365"
	}

	// Check if there are any validation errors
	if len(form.FieldErrors) > 0 {
		// Prepare template data with form and errors
		data := app.newTemplateData(r)
		data.Form = form
		// Render create template with 422 Unprocessable Entity status
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	// Insert the new snippet into the database
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		// Return 500 Internal Server Error if database insertion fails
		app.serverError(w, r, err)
		return
	}

	// Redirect to the new snippet's view page
	// Uses HTTP 303 See Other to prevent duplicate form submissions
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
