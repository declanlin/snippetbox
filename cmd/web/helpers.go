package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	// Generated the formatted text for the provided server error and the debugging stack trace for the
	// call sequence which produced that error.
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	// Log the server error using our custom error logger.
	app.errorLog.Output(2, trace)

	// Send a generic HTTP 500 Internal Server Error response to the client.
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	// Send an HTTP response associated with the specified status code to the client.
	http.Error(w, http.StatusText(status), status)
}

// Wrapper around clientError helper for the particular case in which we want to return an
// HTTP 400 Not Found response to the client.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// Function used to initialize a new templateData struct. As of now, all values are zeroed beside CurrentYear.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
	}
}

// Function used to help render a page being served at the client.
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the template set for the specified page.
	ts, ok := app.templateCache[page]

	// If the requested page does not exist and our handler does not properly respond to this situation,
	// indicate that a server error has occurred.
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// Instead of writing the template straight to the http.ResponseWriter, write it to a byte buffer first.
	// If there is an error in executing the template, we can call the serverError() helper and return, instead of
	// writing the response to the http.ResponseWriter.
	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// If the template is executed and written to the buffer without errors, proceed to setting the response header
	// and writing the contents of the buffer to the http.ResponseWriter.
	w.WriteHeader(status)
	buf.WriteTo(w)
}
