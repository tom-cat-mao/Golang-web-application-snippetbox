package main

import (
	"html/template"
	"path/filepath"
	"time"

	"snippetbox.tomcat.net/internal/models"
)

// templateData holds data to be used when rendering HTML templates. It includes:
// - CurrentYear: The current year for displaying copyright information.
// - Snippet: A single snippet object for displaying individual snippets.
// - Snippets: A slice of snippet objects for displaying lists of snippets.
// - Form: A generic type to hold form data for processing and validation.
// - Flash: A string to display temporary messages to the user.
type templateData struct {
	CurrentYear int // The current year for copyright information.
	Snippet     models.Snippet
	Snippets    []models.Snippet
	Form        any
	Flash       string
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

	// Use filepath.Glob to find all .html files in the ./ui/html/pages/ directory
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	// Iterate over each page file found
	for _, page := range pages {
		// Extract the base name of the file (filename without path)
		name := filepath.Base(page)

		// Create a new template object with the extracted filename and register custom functions
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Add all partial templates to the template set
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		// Add the specific page template to the template set
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Store the parsed template in the cache with its name as the key
		cache[name] = ts

		// Collect a list of template files to include in this page template
		// files := []string{
		// 	"./ui/html/base.html",         // Base layout for the application
		// 	"./ui/html/partials/nav.html", // Navigation partial
		// 	page,                          // Specific page template
		// }

		// ts, err := template.ParseFiles(files...)
		// if err != nil {
		// 	return nil, err
		// }
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
