package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
)

// serverError handles internal server errors by:
// - Logging the error details including method, URI and stack trace
// - Sending a 500 Internal Server Error response to the client
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request details
//   - err: error - The error that occurred
//
// This function is typically called when:
// - Database operations fail
// - Template rendering fails
// - Unexpected errors occur during request processing
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	// Log the error with additional details such as HTTP method and URI, along with a stack trace.
	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)

	// Send a 500 Internal Server Error response to the client.
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError sends an HTTP error response with the specified status code.
// It is used for client-side errors (4xx status codes) such as:
// - 400 Bad Request
// - 404 Not Found
// - 422 Unprocessable Entity
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - status: int - The HTTP status code to send (must be 4xx)
//
// Note: For server errors (5xx), use serverError instead
func (app *application) clientError(w http.ResponseWriter, status int) {
	// Send the specified HTTP status code and its text representation to the client.
	http.Error(w, http.StatusText(status), status)
}

// render handles template rendering with proper error handling and status code management.
// It:
// - Retrieves the template from the cache
// - Executes the template with provided data
// - Handles template execution errors
// - Sets the appropriate HTTP status code
//
// Parameters:
//   - w: http.ResponseWriter - Used to write the HTTP response
//   - r: *http.Request - Contains the incoming HTTP request
//   - status: int - HTTP status code to set (e.g., 200 OK)
//   - page: string - Name of the template to render
//   - data: interface{} - Data to pass to the template
//
// Error Handling:
// - If template doesn't exist: calls serverError
// - If template execution fails: calls serverError
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data interface{}) {
	// Retrieve the appropriate template from the cache based on the provided page name
	ts, ok := app.templateCache[page]

	// Check if the template exists in the cache
	if !ok {
		// If the template does not exist, log an error and call serverError with a custom message
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	// Execute the template with the provided data and render it into the buffer
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Set the HTTP status code in the response header
	w.WriteHeader(status)

	// Write the rendered template output to the response body
	buf.WriteTo(w)
}

// newTemplateData creates and initializes a templateData struct with:
// - Current year for copyright information
// - Flash messages from session
//
// Parameters:
//   - r: *http.Request - Contains the incoming HTTP request
//
// Returns:
//   - templateData: Initialized template data structure
//
// This function is called at the start of each handler to create
// the base data structure that will be passed to templates
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
		Flash:       app.sessionManager.PopString(r.Context(), "flash"),
	}
}

// decodePostForm handles form data decoding with proper error handling.
// It:
// - Parses the form data from the request
// - Decodes the form data into the destination struct
// - Handles invalid decoder errors
//
// Parameters:
//   - r: *http.Request - Contains the incoming HTTP request
//   - dst: any - Destination struct to decode form data into
//
// Returns:
//   - error: Any error that occurred during decoding
//
// Note: This function uses the formDecoder instance from the application
// to handle both URL-encoded and multipart form data
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}
