// Package main defines the HTTP handlers for the SnippetBox web application.
//
// This package contains handler functions that process incoming HTTP requests,
// interact with application logic, and render appropriate HTTP responses.
// Handlers are responsible for:
//   - Request routing and processing
//   - Input data validation and form handling
//   - Rendering HTML templates
//   - Managing user sessions and authentication
//   - Error handling and logging
//
// The handlers in this package adhere to RESTful principles and promote a
// separation of concerns. Each handler focuses on a specific route and follows
// a consistent request processing pattern:
//  1. Parse and validate request input (parameters, form data, etc.).
//  2. Execute the necessary business logic (e.g., database interactions).
//  3. Handle any errors that occur during processing.
//  4. Construct and send the appropriate HTTP response (e.g., HTML, redirects).
package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.tomcat.net/internal/models"
	"snippetbox.tomcat.net/internal/validator"
)

// snippetCreateForm represents the data structure for the snippet creation form
// used in the snippet creation process.
// It handles form data binding, field validation, and error reporting.
//
// Fields:
//   - Title: string - The snippet title (form:"title")
//     Validation rules:
//   - Required: Must not be empty.
//   - Max length: 100 characters.
//   - Not blank: Must contain non-whitespace characters.
//   - Content: string - The snippet content (form:"content")
//     Validation rules:
//   - Required: Must not be empty.
//   - Not blank: Must contain non-whitespace characters.
//   - Expires: int - Expiration in days (form:"expires")
//     Validation rules:
//   - Required: Must not be empty.
//   - Must be 1, 7, or 365.
//   - Validator: validator.Validator - Embedded validator (form:"-")
//     Purpose: Manages validation errors.
//
// Usage:
//   - Used in both GET and POST handlers for snippet creation.
//   - Validates form data before database insertion.
type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// userLoginForm struct represents the form data for user login.
// It includes fields for email and password, along with validation.
//
// Fields:
//   - Email: string - User's email address (form:"email")
//     Validation rules:
//   - Required: Must not be empty.
//   - Must be in valid email format.
//   - Password: string - User's password (form:"password")
//     Validation rules:
//   - Required: Must not be empty.
//   - Validator: validator.Validator - Embedded validator for form validation (form:"-")
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// home handles GET requests to the root URL (/).
//
// It fetches the latest 5 snippets from the database and renders the home page
// template with the snippets.
//
// Parameters:
//   - w: http.ResponseWriter - the HTTP response writer.
//   - r: *http.Request - the HTTP request.
//
// Flow:
//  1. Fetch latest snippets using SnippetModel.Latest().
//  2. Create template data with newTemplateData().
//  3. Add snippets to template data.
//  4. Render "home.html" template.
//
// Error Handling:
//   - Database errors: 500 Internal Server Error.
//   - Template errors: 500 Internal Server Error.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest 5 snippets from the database.
	snippets, err := app.snippets.Latest() // Fetch latest 5 snippets from database.
	if err != nil {
		// If there's an error fetching snippets, return a 500 Internal Server Error.
		app.serverError(w, r, err)
		return
	}

	// Create a template data structure with default values.
	data := app.newTemplateData(r)
	data.Snippets = snippets // Add snippets to template data.

	// Render the "home.html" template with the provided data.
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
// - Initializes template data with a default expiration of 365 days.
// - Renders the create form template.
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response.
//   - r: *http.Request - Contains the incoming HTTP request.
//
// Flow:
// 1. Create template data.
// 2. Set default form values (Expires to 365 days).
// 3. Render "create.html" template.
//
// Error Handling:
// - Template errors: 500 Internal Server Error.
//
// Note: The actual form submission is handled by snippetCreatePost.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data structure.
	data := app.newTemplateData(r) // Initialize template data.

	data.Form = snippetCreateForm{
		Expires: 365, // Set default expiration to 365 days.
	}

	// Render the "create.html" template with the provided data.
	app.render(w, r, http.StatusOK, "create.html", data)
}

// snippetCreatePost handles POST requests to create a new snippet.
// It:
// - Parses form data from the request.
// - Validates the parsed form data.
// - Inserts the snippet into the database if the data is valid.
// - Redirects to the snippet view page on success.
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response.
//   - r: *http.Request - Contains the incoming HTTP request.
//
// Flow:
// 1. Parse and decode form data.
// 2. Validate form fields.
// 3. Handle validation errors by re-rendering the form with error messages.
// 4. Insert snippet into database.
// 5. Handle database errors.
// 6. Set flash message to indicate successful creation.
// 7. Redirect to new snippet's view page.
//
// Error Handling:
// - Invalid form data: 400 Bad Request.
// - Validation errors: 422 Unprocessable Entity.
// - Database errors: 500 Internal Server Error.
//
// Returns:
// - Success: 303 See Other redirect to /snippet/view/{id}.
// - Validation error: 422 with error messages.
// - Database error: 500 Internal Server Error.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Initialize form struct to hold submitted data.
	var form snippetCreateForm

	// Decode form data from POST request into struct.
	// Uses formDecoder to handle both URL-encoded and multipart form data.
	err := app.decodePostForm(r, &form)
	if err != nil {
		// Return 400 Bad Request if form decoding fails.
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate form fields.
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() { // Check if form validation failed.
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	// Insert the new snippet into the database.
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		// Return 500 Internal Server Error if database insertion fails.
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created") // Flash message on success.

	// Redirect to show the new snippet.
	// Uses HTTP 303 See Other to prevent duplicate form submissions.
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
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}

// userSignupPost handles POST requests to create a new user.
// It processes the signup form submission, validates the input data,
// creates a new user account, and handles the signup workflow.
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request and form data
//
// Flow:
// 1. Decode POST form data into userSignupForm struct
// 2. Validate form fields:
//   - Name: Must not be blank
//   - Email: Must be valid email format
//   - Password: Must be at least 8 characters
//
// 3. If validation fails, re-render form with error messages
// 4. Attempt to create new user in database
// 5. Handle potential duplicate email addresses
// 6. Set success flash message
// 7. Redirect to login page
//
// Error Handling:
//   - Invalid form data: 400 Bad Request
//   - Validation errors: 422 Unprocessable Entity with form errors
//   - Duplicate email: 422 Unprocessable Entity with email error
//   - Database errors: 500 Internal Server Error
//
// Returns:
//   - Success: 303 See Other redirect to /user/login
//   - Error: Appropriate error response based on error type
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Initialize form struct to hold submitted data
	var form userSignupForm

	// Decode form data from POST request into struct
	err := app.decodePostForm(r, &form)
	if err != nil {
		// Return 400 Bad Request if form decoding fails
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate all form fields
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	// If validation fails, re-render the signup form with error messages
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	// Attempt to create a new user record in the database
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			// If email already exists, add an error message and re-render the form
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			// For any other database error, return 500 Internal Server Error
			app.serverError(w, r, err)
		}
		return
	}

	// Add a success flash message to be displayed on the login page
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	// Redirect to the login page with status 303 See Other
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// userLogin handles GET requests to display the user login form.
//
// It initializes template data and renders the login form template.
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
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}

// userLoginPost handles POST requests to authenticate and login a user.
// It:
// - Parses form data from the request body.
// - Validates the user's email address and password.
// - Authenticates the user against the database.
// - Creates a new session for the user.
// - Redirects the user to the home page.
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response.
//   - r: *http.Request - Contains the incoming HTTP request.
//
// Flow:
// 1. Decode the form data from the request body.
// 2. Validate the form data.
// 3. If validation fails, re-render the login form with error messages.
// 4. Authenticate the user against the database.
// 5. If authentication fails, re-render the login form with an error message.
// 6. Create a new session for the user.
// 7. Redirect the user to the home page.
//
// Error Handling:
// - Invalid form data: 400 Bad Request.
// - Validation errors: 401 Unauthorized.
// - Authentication errors: 401 Unauthorized.
// - Session errors: 500 Internal Server Error.
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
	var form userLoginForm

	err := app.decodePostForm(r, &form)

	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
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
	// Logout the user.
	fmt.Fprintln(w, "Logout the user...")
}
