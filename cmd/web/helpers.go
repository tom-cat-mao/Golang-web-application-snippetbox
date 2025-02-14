package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// serverError logs an error and sends a 500 Internal Server Error response.
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

// clientError sends a specified HTTP status code and its corresponding message to the response.
func (app *application) clientError(w http.ResponseWriter, status int) {
	// Send the specified HTTP status code and its text representation to the client.
	http.Error(w, http.StatusText(status), status)
}

// render is a helper method that renders an HTML template.
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

// newTemplateData creates a new instance of templateData with current year set.
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}
