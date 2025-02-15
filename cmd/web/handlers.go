// Package web implements the HTTP handlers for the SnippetBox application.
//
// This package contains all the request handlers that process incoming HTTP requests
// and generate appropriate responses. It handles:
// - Routing and request processing
// - Data validation and form handling
// - Template rendering and response generation
// - Error handling and logging
// - Middleware integration
//
// The package follows RESTful principles and implements clean separation of concerns.
// Each handler is responsible for a specific route and follows a consistent pattern:
// 1. Parse and validate input
// 2. Process business logic
// 3. Handle errors appropriately
// 4. Generate and return response
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.tomcat.net/internal/models"
	"snippetbox.tomcat.net/internal/validator"
)

// snippetCreateForm encapsulates all data and validation state for the snippet creation form.
// It handles form submission, validation, and error reporting.
//
// Fields:
//   - Title: The snippet title
//     Type: string
//     Tag: form:"title" (maps to HTML form field name)
//     Validation:
//   - Required field
//   - Maximum length: 100 characters
//   - Must not be empty
//   - Content: The snippet content
//     Type: string
//     Tag: form:"content" (maps to HTML form field name)
//     Validation:
//   - Required field
//   - Must not be empty
//   - Expires: Number of days until the snippet expires
//     Type: int
//     Tag: form:"expires" (maps to HTML form field name)
//     Validation:
//   - Required field
//   - Must be one of: 1, 7, or 365
//   - Validator: Embedded validator instance
//     Type: validator.Validator
//     Tag: form:"-" (excluded from form binding)
//     Purpose: Handles field validation and error collection
//
// The form follows these validation rules:
// 1. All fields are required
// 2. Title must be non-empty and ≤ 100 characters
// 3. Content must be non-empty
// 4. Expires must be 1, 7, or 365
// 5. Validation errors are collected in the Validator instance
type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

// home handles GET requests to the application's homepage (/).
// It retrieves the 5 most recent snippets from the database and renders them
// using the home page template.
//
// Security:
//   - No authentication required
//   - No sensitive data exposed
//
// Performance:
//   - Single database query for latest snippets
//   - Template rendering overhead
//
// Caching:
//   - No caching headers set
//   - Consider adding Cache-Control for static content
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// The handler follows these steps:
// 1. Fetch latest snippets from database
//   - Uses models.SnippetModel.Latest() method
//   - Retrieves maximum of 5 snippets
//
// 2. Create template data structure
//   - Initializes with default values
//   - Adds snippets to data.Snippets
//
// 3. Render home page template
//   - Uses application.render() helper
//   - Template: "home.html"
//   - Status code: 200 OK
//
// Error Handling:
//   - Database errors: Returns 500 Internal Server Error
//   - Template rendering errors: Returns 500 Internal Server Error
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest 5 snippets from the database
	snippets, err := app.snippets.Latest() // Fetch latest 5 snippets from database
	if err != nil {
		// If there's an error fetching snippets, return a 500 Internal Server Error
		app.serverError(w, r, err)
		return
	}

	// Create a template data structure with default values
	data := app.newTemplateData(r)
	data.Snippets = snippets // Add snippets to template data

	// Render the "home.html" template with the provided data
	app.render(w, r, http.StatusOK, "home.html", data)
}

// snippetView handles GET requests to view a specific snippet (/snippet/view/{id}).
// It accepts a snippet ID as a URL parameter, retrieves the corresponding snippet
// from the database, and renders it using the view template.
//
// Security:
//   - No authentication required
//   - No sensitive data exposed
//
// Performance:
//   - Single database query by primary key
//   - Template rendering overhead
//
// Caching:
//   - No caching headers set
//   - Consider adding Cache-Control for static content
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// URL Parameters:
//   - id: The unique identifier of the snippet to view
//
// The handler follows these steps:
// 1. Extract and validate snippet ID from URL
//   - Uses strconv.Atoi() for conversion
//   - Validates ID is positive integer
//
// 2. Retrieve snippet from database
//   - Uses models.SnippetModel.Get() method
//
// 3. Handle database errors
//   - If record not found: return 404
//   - If other error: return 500
//
// 4. Prepare template data
//   - Initializes with default values
//   - Adds snippet to data.Snippet
//
// 5. Render view template
//   - Uses application.render() helper
//   - Template: "view.html"
//   - Status code: 200 OK
//
// Error Handling:
//   - Invalid ID format: Returns 404 Not Found
//   - Snippet not found: Returns 404 Not Found
//   - Database errors: Returns 500 Internal Server Error
//   - Template rendering errors: Returns 500 Internal Server Error
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id")) // Convert URL parameter to integer
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
	data.Snippet = snippet // Add snippet to template data

	// Render the "view.html" template with the provided data
	app.render(w, r, http.StatusOK, "view.html", data)
}

// snippetCreate handles GET requests to display the snippet creation form (/snippet/create).
// It renders an HTML form that allows users to input new snippet information.
// This handler only displays the form - the actual creation is handled by
// snippetCreatePost.
//
// Security:
//   - No authentication required
//   - No sensitive data exposed
//
// Performance:
//   - No database queries
//   - Template rendering overhead
//
// Caching:
//   - No caching headers set
//   - Consider adding Cache-Control for static content
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// The handler follows these steps:
// 1. Create template data structure
//   - Initializes with default values
//   - Sets form with default expires value (365 days)
//
// 2. Render create form template
//   - Uses application.render() helper
//   - Template: "create.html"
//   - Status code: 200 OK
//
// Error Handling:
//   - Template rendering errors: Returns 500 Internal Server Error
//
// Note: This is a GET handler that displays the form. The actual snippet creation
// is handled by snippetCreatePost.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data structure
	data := app.newTemplateData(r) // Initialize template data

	data.Form = snippetCreateForm{
		Expires: 365, // Set default expiration to 365 days
	}

	// Render the "create.html" template with the provided data
	app.render(w, r, http.StatusOK, "create.html", data)
}

// snippetCreatePost handles POST requests to create a new snippet (/snippet/create).
// It processes the form submission, validates the input data, and creates
// a new snippet record in the database.
//
// Security:
//   - No authentication required
//   - No sensitive data exposed
//
// Performance:
//   - Single database insert operation
//   - Form parsing and validation overhead
//
// Caching:
//   - No caching headers set
//
// Concurrency:
//   - Safe for concurrent use
//   - Database handles concurrent inserts
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//
// Form Data:
//   - title: The snippet title
//     Type: string
//     Validation:
//   - Required field
//   - Maximum length: 100 characters
//   - Must not be empty
//   - content: The snippet content
//     Type: string
//     Validation:
//   - Required field
//   - Must not be empty
//   - expires: Number of days until the snippet expires
//     Type: int
//     Validation:
//   - Required field
//   - Must be one of: 1, 7, or 365
//
// The handler follows these steps:
// 1. Parse form data
//   - Uses r.ParseForm()
//   - Handles both URL-encoded and multipart form data
//
// 2. Validate required fields
//   - Title: not blank, max 100 chars
//   - Content: not blank
//   - Expires: must be 1, 7 or 365
//
// 3. Handle validation errors
//   - If errors exist, re-render form with error messages
//   - Status code: 422 Unprocessable Entity
//
// 4. Insert new snippet into database
//   - Uses models.SnippetModel.Insert()
//
// 5. Handle database errors
//   - Returns 500 Internal Server Error
//
// 6. Redirect to new snippet's view page
//   - Uses HTTP 303 See Other
//   - Prevents duplicate form submissions
//
// Error Handling:
//   - Invalid form data: Returns 400 Bad Request
//   - Validation errors: Returns 422 Unprocessable Entity
//   - Database errors: Returns 500 Internal Server Error
//   - Template rendering errors: Returns 500 Internal Server Error
//
// Returns:
//   - On success: HTTP 303 redirect to /snippet/view/{id}
//   - On validation error: HTTP 422 with error messages
//   - On database error: HTTP 500 Internal Server Error
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Parse form data from the request
	// Handles both URL-encoded and multipart form data
	err := r.ParseForm() // Parse form data from request
	if err != nil {
		// Return 400 Bad Request if form parsing fails
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Initialize form struct to hold submitted data
	var form snippetCreateForm

	// Decode form data from POST request into struct
	// Uses formDecoder to handle both URL-encoded and multipart form data
	err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		// Return 400 Bad Request if form decoding fails
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate form fields
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() { // Check if form validation failed
		data := app.newTemplateData(r)
		data.Form = form
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
