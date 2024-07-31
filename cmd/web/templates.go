package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/declanlin/snippetbox/internal/models"
	"github.com/declanlin/snippetbox/ui"
)

// Define a type templateData which stores additional information that will be passed to ExecuteTemplate().
// This data will be accessed by the HTML templates and used to render the necessary page(s) for a route.
type templateData struct {
	CurrentYear     int
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

// Converts a Go time.Time object to a human-readable string.
func humanDate(t time.Time) string {

	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Map the names of template functions onto their implementations to be executed by a template.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize an empty cache.
	// This cache will operate in memory to store the template sets for each HTML page we our serving.
	// It maps the base element of each HTML page path to its template set.

	cache := map[string]*template.Template{}

	// Retrieve the name of all files in the ui.Files embedded filesystem matching the specified glob pattern
	// as a slice of strings.
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Iterate over each of the pages being served.
	for _, page := range pages {
		// Extract the base element of the page path as the name of the template set.
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we want to parse.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Use ParseFS() instead of ParseFiles() to parse the template files from the
		// ui.Files embedded filesystem into a template set.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		// Cache a mapping of the HTML page path's base element to its template set.

		cache[name] = ts
	}

	// Return the template cache with no errors.
	return cache, nil
}
