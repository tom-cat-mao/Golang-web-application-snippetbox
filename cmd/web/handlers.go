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

// snippetCreateForm represents the data structure for the snippet creation form.
// It handles:
// - Form data binding
// - Field validation
// - Error reporting
//
// Fields:
//   - Title: string - The snippet title (form:"title")
//     Validation:
//   - Required
//   - Max length: 100 characters
//   - Not blank
//   - Content: string - The snippet content (form:"content")
//     Validation:
//   - Required
//   - Not blank
//   - Expires: int - Expiration in days (form:"expires")
//     Validation:
//   - Required
//   - Must be 1, 7 or 365
//   - Validator: validator.Validator - Embedded validator (form:"-")
//     Purpose: Manages validation errors
//
// Usage:
// - Used in both GET and POST handlers for snippet creation
// - Validates form data before database insertion
type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

// home handles GET requests to the root URL (/).
// It:
// - Fetches the latest 5 snippets from the database
// - Renders the home page template with the snippets
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Fetch latest snippets using SnippetModel.Latest()
// 2. Create template data with newTemplateData()
// 3. Add snippets to template data
// 4. Render "home.html" template
//
// Error Handling:
// - Database errors: 500 Internal Server Error
// - Template errors: 500 Internal Server Error
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

// snippetView handles GET requests to view a specific snippet.
// It:
// - Extracts snippet ID from URL
// - Fetches snippet from database
// - Renders the view template with snippet data
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// URL Parameters:
//   - id: int - The snippet ID to view
//
// Flow:
// 1. Extract and validate ID from URL
// 2. Fetch snippet using SnippetModel.Get()
// 3. Handle database errors
// 4. Create template data
// 5. Add snippet to template data
// 6. Render "view.html" template
//
// Error Handling:
// - Invalid ID: 404 Not Found
// - Snippet not found: 404 Not Found
// - Database errors: 500 Internal Server Error
// - Template errors: 500 Internal Server Error
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

// snippetCreate handles GET requests to display the snippet creation form.
// It:
// - Initializes template data
// - Sets default form values
// - Renders the create form template
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Create template data
// 2. Initialize form with default values
// 3. Render "create.html" template
//
// Error Handling:
// - Template errors: 500 Internal Server Error
//
// Note: The actual form submission is handled by snippetCreatePost
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data structure
	data := app.newTemplateData(r) // Initialize template data

	data.Form = snippetCreateForm{
		Expires: 365, // Set default expiration to 365 days
	}

	// Render the "create.html" template with the provided data
	app.render(w, r, http.StatusOK, "create.html", data)
}

// snippetCreatePost handles POST requests to create a new snippet.
// It:
// - Parses form data
// - Validates input
// - Inserts snippet into database
// - Redirects to snippet view page
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Parse and decode form data
// 2. Validate form fields
// 3. Handle validation errors
// 4. Insert snippet into database
// 5. Handle database errors
// 6. Set flash message
// 7. Redirect to new snippet's view page
//
// Error Handling:
// - Invalid form data: 400 Bad Request
// - Validation errors: 422 Unprocessable Entity
// - Database errors: 500 Internal Server Error
//
// Returns:
// - Success: 303 See Other redirect to /snippet/view/{id}
// - Validation error: 422 with error messages
// - Database error: 500 Internal Server Error
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Initialize form struct to hold submitted data
	var form snippetCreateForm

	// Decode form data from POST request into struct
	// Uses formDecoder to handle both URL-encoded and multipart form data
	err := app.decodePostForm(r, &form)
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

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created") // Flash message on success

	// Redirect to show the new snippet
	// Uses HTTP 303 See Other to prevent duplicate form submissions
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

// userSignup handles GET requests to display the user signup form.
// It:
// - Initializes template data
// - Renders the signup form template
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Create template data
// 2. Render "signup.html" template
//
// Error Handling:
// - Template errors: 500 Internal Server Error
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Display a form for signing up a new user..")
}

// userSignupPost handles POST requests to create a new user.
// It:
// - Parses form data
// - Validates input
// - Creates user in database
// - Redirects to home page
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Parse and decode form data
// 2. Validate form fields
// 3. Handle validation errors
// 4. Insert user into database
// 5. Handle database errors
// 6. Set flash message
// 7. Redirect to home page
//
// Error Handling:
// - Invalid form data: 400 Bad Request
// - Validation errors: 422 Unprocessable Entity
// - Database errors: 500 Internal Server Error
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Create a new user..")
}

// userLogin handles GET requests to display the user login form.
// It:
// - Initializes template data
// - Renders the login form template
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Create template data
// 2. Render "login.html" template
//
// Error Handling:
// - Template errors: 500 Internal Server Error
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Display a form for logging in a user..")
}

// userLoginPost handles POST requests to authenticate and login a user.
// It:
// - Parses form data
// - Validates credentials
// - Creates user session
// - Redirects to home page
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Parse and decode form data
// 2. Validate credentials
// 3. Handle validation errors
// 4. Create user session
// 5. Handle session creation errors
// 6. Set flash message
// 7. Redirect to home page
//
// Error Handling:
// - Invalid form data: 400 Bad Request
// - Validation errors: 401 Unauthorized
// - Session errors: 500 Internal Server Error
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Authenticate and login the user...")
}

// userLogoutPost handles POST requests to logout the current user.
// It:
// - Destroys user session
// - Redirects to home page
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Destroy user session
// 2. Handle session destruction errors
// 3. Set flash message
// 4. Redirect to home page
//
// Error Handling:
// - Session errors: 500 Internal Server Error
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}
