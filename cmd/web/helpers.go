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

// serverError handles internal server errors by logging the error details and sending a 500 response.
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - err: error that occurred
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

// clientError sends a specified HTTP status code and its corresponding message to the client.
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - status: HTTP status code to send
func (app *application) clientError(w http.ResponseWriter, status int) {
	// Send the specified HTTP status code and its text representation to the client.
	http.Error(w, http.StatusText(status), status)
}

// render is a helper method that renders an HTML template with the provided data.
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - status: HTTP status code to set
//   - page: template name to render
//   - data: template data to pass to the template
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

// newTemplateData creates and initializes a new templateData struct with the current year.
// Parameters:
//   - r: *http.Request containing the incoming HTTP request
//
// Returns:
//   - templateData: initialized template data structure
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}

// decodePostForm decodes form data from the request into the provided destination struct.
// Parameters:
//   - r: *http.Request containing the incoming HTTP request
//   - dst: destination struct to decode form data into
//
// Returns:
//   - error: if decoding fails
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
