package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.tomcat.net/internal/models"
	"snippetbox.tomcat.net/ui"
)

// templateData holds data to be used when rendering HTML templates. It includes:
// - CurrentYear: The current year for displaying copyright information.
// - Snippet: A single snippet object for displaying individual snippets.
// - Snippets: A slice of snippet objects for displaying lists of snippets.
// - Form: A generic type to hold form data for processing and validation.
// - Flash: A string to display temporary messages to the user.
// - IsAuthenticated: A boolen to verify if it is authenticated
type templateData struct {
	CurrentYear     int // The current year for copyright information.
	Snippet         models.Snippet
	Snippets        []models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string // a CSRFToken field
}

// newTemplateCache initializes a template cache by parsing all HTML templates from the ui/html directory.
// It performs the following steps:
// 1. Finds all page templates in the ui/html/pages/ directory.
// 2. Creates a new template set for each page found.
// 3. Registers custom template functions to extend template capabilities.
// 4. Parses the base template and all partial templates.
// 5. Adds the specific page template to the set.
// 6. Stores the parsed templates in a map for efficient lookup during rendering.
//
// Returns:
// - map[string]*template.Template: A map of template names to parsed templates, allowing quick access by name.
// - error: Any error encountered during the template parsing process.
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize an empty map to store the parsed templates
	cache := map[string]*template.Template{}

	// Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded
	// filesystem which match the pattern `html/pages/*.html", This essentially`
	// gives us a slice of all the 'page' templates for the application, just
	// like before
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	// Iterate over each page file found
	for _, page := range pages {
		// Extract the base name of the file (filename without path)
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we
		// want to parse.
		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		// Use ParseFS() instead of ParseFile() to parse the template files
		// from the ui.Files embedded filesystem.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Store the parsed template in the cache with its name as the key
		cache[name] = ts
	}

	return cache, nil
}

// humanDate converts a time.Time value to a human-readable string format.
// This function is used to display dates in a user-friendly format throughout the application.
// The format used is "02 Jan 2006 at 15:04", which represents day, month, year, and time.
//
// Parameters:
// - t: time.Time - The time value to be formatted.
//
// Returns:
// - string: The formatted date string in the specified format.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// functions is a template.FuncMap that defines custom functions available within HTML templates.
// These functions extend the capabilities of Go's template engine, allowing for more dynamic and formatted output.
// The map includes:
// - "humanDate": A function to format dates in a human-readable way.
var functions = template.FuncMap{
	"humanDate": humanDate,
}
