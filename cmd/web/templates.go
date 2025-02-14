package main

import (
	"html/template"
	"path/filepath"

	"snippetbox.tomcat.net/internal/models"
)

type templateData struct {
	Snippet  models.Snippet
	Snippets []models.Snippet
}

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

		// Collect a list of template files to include in this page template
		files := []string{
			"./ui/html/base.html",         // Base layout for the application
			"./ui/html/partials/nav.html", // Navigation partial
			page,                          // Specific page template
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
