// Package main defines the HTTP handlers for the SnippetBox web application.
//
// This package contains handler functions that manage the core functionality of the application,
// including snippet management, user authentication, and session management. Each handler:
//   - Processes incoming HTTP requests
//   - Validates and sanitizes input data
//   - Interacts with the application's business logic and data models
//   - Renders HTML templates or redirects to other routes
//   - Manages user sessions and authentication
//   - Handles errors and logs them for debugging and monitoring
//
// The handlers adhere to RESTful principles, promoting a clear separation of concerns.
// They follow a consistent processing pattern:
//  1. Parse and validate request input (parameters, form data, etc.)
//  2. Execute business logic, often involving database operations
//  3. Handle errors, providing appropriate feedback to the user
//  4. Construct and send HTTP responses, including HTML, JSON, or redirects
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

// userSignupForm represents the form data and validation rules for user registration.
// It handles form data binding, validation, and error reporting for the signup process.
//
// Fields:
//   - Name: string - User's full name (form:"name")
//     Validation rules:
//   - Required: Must not be blank
//   - Email: string - User's email address (form:"email")
//     Validation rules:
//   - Required: Must not be blank
//   - Must be valid email format
//   - Password: string - User's password (form:"password")
//     Validation rules:
//   - Required: Must not be blank
//   - Minimum length: 8 characters
//   - Validator: validator.Validator - Embedded validator for error management (form:"-")
//     Purpose: Tracks and reports validation errors across form fields
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

// accountPasswordUpdateForm represents the form data and validation rules for updating the account password.
// It handles form data binding, validation, and error reporting for the password update process.
//
// Fields:
//   - CurrentPassword: string - User's current password (form:"current_password")
//     Validation rules:
//   - Required: Must not be blank
//   - NewPassword: string - User's new password (form:"new_password")
//     Validation rules:
//   - Required: Must not be blank
//   - Minimum length: 8 characters
//   - NewPasswordConfirmation: string - Confirmation of the new password (form:"new_password_confirmation")
//     Validation rules:
//   - Required: Must match the new password
//   - Validator: validator.Validator - Embedded validator for error management (form:"-")
//     Purpose: Tracks and reports validation errors across form fields
type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"current_password"`
	NewPassword             string `form:"new_password"`
	NewPasswordConfirmation string `form:"new_password_confirmation"`
	validator.Validator     `form:"-"`
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
	// Extract and convert the snippet ID from the URL parameter
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
// It processes login form submissions by:
//   - Parsing and validating form data
//   - Authenticating user credentials
//   - Managing user sessions
//   - Handling redirects after successful login
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Processing Flow:
// 1. Decode and parse form data from request body
// 2. Validate email and password fields:
//   - Email must be non-blank and valid format
//   - Password must be non-blank
//
// 3. If validation fails:
//   - Re-render login form with error messages (HTTP 422)
//
// 4. Authenticate user credentials against database
// 5. If authentication fails:
//   - Re-render login form with generic error message (HTTP 422)
//
// 6. If authentication succeeds:
//   - Renew session token for security
//   - Store authenticated user ID in session
//   - Redirect to either:
//   - Original requested path (if available)
//   - Home page (default)
//
// Error Handling:
// - Invalid form data: HTTP 400 Bad Request
// - Validation errors: HTTP 422 Unprocessable Entity
// - Authentication errors: HTTP 422 Unprocessable Entity
// - Session errors: HTTP 500 Internal Server Error
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

	// Use PopString to retrieve the path and remove it from the session atomically.
	// It returns the string value and a bool indicating if the key existed and held a string.
	path := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")

	// Check if the value existed in the session AND was a string.
	// The check for path != "" might be slightly redundant if PopString guarantees
	// ok=true only for non-empty strings it finds, but it's safe to include.
	if path != "" {
		// Value existed and was a non-empty string.
		// No need to call Remove() here, PopString already did it.
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}

	// If the path wasn't in the session, or wasn't a string, or was empty,
	// redirect to the default page (e.g., account view)
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}

// userLogoutPost handles POST requests to logout the current user.
// It performs the following operations:
// - Renews the session token to prevent session fixation attacks. This is important because:
//   - Session fixation attacks occur when an attacker fixes a user's session ID before they log in.
//   - By renewing the session token upon logout, we ensure that the session ID used during the logged-in state is invalidated.
//   - This prevents an attacker from using the old session ID to gain unauthorized access after the user logs out.
//
// - Removes the 'authenticatedUserID' from the session
// - Sets a flash message indicating successful logout
// - Redirects the user to the home page ('/')
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Renew the session token
// 2. Remove 'authenticatedUserID' from the session
// 3. Set a flash message
// 4. Redirect to the home page ('/')
//
// Error Handling:
// - Session errors during token renewal or user ID removal: 500 Internal Server Error
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Logout the user.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// about handles GET requests to the /about endpoint.
// It performs the following operations:
// - Creates a new template data structure populated with common data
// - Renders the about.html template with a 200 OK status
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Create new template data using application's newTemplateData method
// 2. Render the about.html template with the prepared data
//
// Template:
// - Uses ui/html/pages/about.html template
// - Displays static content about the application
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "about.html", data)
}

// ping handles GET requests to the /ping endpoint.
// It's a simple health check endpoint that:
// - Returns a 200 OK status
// - Writes "OK" as the response body
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// 1. Write "OK" as the response body
// 2. Implicitly returns 200 OK status
//
// Usage:
// - Used for health checks and monitoring
// - Verifies that the application is running and responding to requests
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// accountView handles GET requests to display the authenticated user's account information.
// It performs the following operations:
// 1. Retrieves the authenticated user's ID from the session
// 2. Fetches the user's details from the database using the user ID
// 3. Handles potential errors:
//   - Redirects to login page if user is not found
//   - Returns server error for other database errors
//
// 4. Prepares template data with the user's information
// 5. Renders the account.html template with the user's data
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// - Get user ID from session
// - Retrieve user from database
// - Handle errors appropriately
// - Prepare template data
// - Render account page
//
// Template:
// - Uses ui/html/pages/account.html template
// - Displays user's name, email, and account creation date
func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user's ID from the session
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Retrieve the user's details from the database
	user, err := app.users.Get(userID)
	if err != nil {
		// Handle case where user is not found
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			// Handle other database errors
			app.serverError(w, r, err)
		}
		return
	}

	// Prepare template data and add the user's information
	data := app.newTemplateData(r)
	data.User = user

	// Render the account page template
	app.render(w, r, http.StatusOK, "account.html", data)
}

// accountPasswordUpdate handles GET requests to display the password update form.
// It performs the following operations:
// 1. Initializes template data
// 2. Prepares an empty accountPasswordUpdateForm struct
// 3. Renders the password update form template
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//
// Flow:
// - Create template data
// - Initialize empty form struct
// - Render "password.html" template
//
// Template:
// - Uses ui/html/pages/password.html template
// - Displays form fields for current password, new password, and password confirmation
func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}

	app.render(w, r, http.StatusOK, "password.html", data)
}

// accountPasswordUpdatePost handles POST requests to update a user's password.
// It performs the following operations:
// 1. Retrieves the authenticated user's ID from the session
// 2. Decodes and validates the password update form data
// 3. Validates form fields:
//   - Current password: Must not be blank
//   - New password: Must not be blank and at least 8 characters
//   - Password confirmation: Must match new password
//
// 4. If validation fails, re-renders form with error messages
// 5. Attempts to update password in database
// 6. Handles incorrect current password case
// 7. Sets success flash message and redirects to account view on success
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request and form data
//
// Error Handling:
//   - Invalid form data: 400 Bad Request
//   - Validation errors: 422 Unprocessable Entity with form errors
//   - Incorrect current password: 422 Unprocessable Entity with error
//   - Database errors: 500 Internal Server Error
//
// Returns:
//   - Success: 303 See Other redirect to /account/view with flash message
//   - Error: Appropriate error response based on error type
func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	// Initialize form struct and get authenticated user ID
	var form accountPasswordUpdateForm
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Decode form data from POST request
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate all form fields
	form.CheckField(validator.NotBlank(form.CurrentPassword), "current_password", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "new_password", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "new_password_confirmation", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "new_password", "This field must be at least 8 characters long")
	form.CheckField(form.NewPassword == form.NewPasswordConfirmation, "new_password_confirmation", "Passwords do not match")

	// If validation fails, re-render form with error messages
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "password.html", data)
		return
	}

	// Attempt to update password in database
	err = app.users.PasswordUpdate(userID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		// Handle incorrect current password case
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("current_password", "Current password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "password.html", data)
		} else {
			// Handle other database errors
			app.serverError(w, r, err)
		}
		return
	}

	// Set success flash message and redirect to account view
	app.sessionManager.Put(r.Context(), "flash", "Password updated successfully")
	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}
