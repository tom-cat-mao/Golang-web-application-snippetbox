package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.tomcat.net/internal/models"
)

// Handles requests to display the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Retrieve the 5 most recent snippets from the database.
	snippets, err := app.snippets.Latest()
	if err != nil {
		// If an error occurs, call the serverError method to handle it and return.
		app.serverError(w, r, err)
		return
	}

	// Create a new template data struct with default values.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Render the "home.html" template with the provided data.
	app.render(w, r, http.StatusOK, "home.html", data)
}

// Handles requests to view a specific snippet.
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		// If an error occurs or the ID is invalid, return a 404 Not Found response.
		http.NotFound(w, r)
		return
	}

	// Retrieve the snippet from the database based on the provided ID.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// If the error indicates that no record was found, return a 404 Not Found response.
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			// For other types of errors, call the serverError method to handle them and return.
			app.serverError(w, r, err)
		}
		return
	}

	// Create a new template data struct with default values.
	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Render the "view.html" template with the provided data.
	app.render(w, r, http.StatusOK, "view.html", data)
}

// Handles requests to display the create snippet page.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data struct with default values.
	data := app.newTemplateData(r)

	// Render the "create.html" template with the provided data.
	app.render(w, r, http.StatusOK, "create.html", data)
}

// Handles POST requests to create a new snippet.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	title := "0 snail"
	content := "0 snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n- Kobayashi Issa"
	expires := 7

	// Insert the new snippet into the database.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		// If an error occurs, call the serverError method to handle it and return.
		app.serverError(w, r, err)
		return
	}

	// Redirect the user to the view page of the newly created snippet using a 303 See Other status code.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
