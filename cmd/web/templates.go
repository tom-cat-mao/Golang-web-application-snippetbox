package main

import (
	"html/template"
	"path/filepath"

	"snippetbox.tomcat.net/internal/models"
)

type templateData struct {
	CurrentYear int
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

		// Create a new template set with the base.html as the layout
		ts, err := template.ParseFiles("./ui/html/base.html")
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
