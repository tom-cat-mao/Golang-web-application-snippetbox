package main

import (
	"html/template"
	"path/filepath"
	"time"

	"snippetbox.tomcat.net/internal/models"
)

// templateData holds data to be used when rendering HTML templates.
type templateData struct {
	CurrentYear int // The current year for copyright information.
	Snippet     models.Snippet
	Snippets    []models.Snippet
}

// newTemplateCache reads all HTML files from the UI directory and parses them into a cache of *template.Template objects.
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

// humanDate formats a time.Time object to a user-friendly string.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// functions is a template.FuncMap containing custom template functions.
var functions = template.FuncMap{
	"humanDate": humanDate,
}
