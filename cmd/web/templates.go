package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/declanlin/snippetbox/internal/models"
)

// Define a type templateData which stores additional information that will be passed to ExecuteTemplate().
// This data will be accessed by the HTML templates and used to render the necessary page(s) for a route.
type templateData struct {
	CurrentYear int
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
	Form        any
}

// Converts a Go time.Time object to a human-readable string.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
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

	// Retrieve the name of all files matching the specified glob pattern as a slice of strings.
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Iterate over each of the pages being served.
	for _, page := range pages {
		// Extract the base element of the page path as the name of the template set.
		name := filepath.Base(page)

		// Create a new template set. Register the function map to this new template set.
		// Parse the base.tmpl file for the current HTML page into the template set.
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		// Parse any of the HTML page partials that our web application is serving into the template set.
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		// Parse the .tmpl file for the current HTML page into the template set.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Cache a mapping of the HTML page path's base element to its template set.
		cache[name] = ts
	}

	// Return the template cache with no errors.
	return cache, nil
}
