package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/declanlin/snippetbox/internal/models"
	"github.com/declanlin/snippetbox/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	// Fetch a slice of the 10 most recently created snippets.
	snippets, err := app.snippets.Latest()

	// If there is an error in fetching the slice, log a server error and return.
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Initialize a new templateData struct to store the slice of snippets.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Render the templates code associated with the specified template page.
	app.render(w, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// ParamsFromContext() pulls the URL parameters from a request context, or returns nil if none are present
	params := httprouter.ParamsFromContext(r.Context())

	// Parse the "id" parameter from the http.Params.
	id, err := strconv.Atoi(params.ByName("id"))

	// If there is an error parsing the string id as an integer, or the parsed id is less than 1, we will consider
	// the resource to not exist.
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// Query the database for a snippet with the specified ID. Remember that we have specially returned a custom
	// ErrNoRecord error from the Get function for a snippet. We will want to check this, and handle it by returning
	// an HTTP 404 Not Found response, as opposed to a server error.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Initialize a new templateData struct to store the snippet.
	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Render the template code associated with the specified template page.
	app.render(w, http.StatusOK, "view.tmpl", data)
}

// Define a struct to represent the form data and validation errors for the form fields.
type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
	validator.Validator
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Initialize a new templateData struct to store additional resources for the template execution.
	data := app.newTemplateData(r)

	// Set the default value for the expiry time to be 365 days.

	// Without the code below, the server would crash when a user first visits the "/snippet/create" route.
	// This is because the application attempts to render the create.tmpl template, but since the value of
	// the Form field in the template data returned by newTemplateData() is initially nil, it crashes when
	// it attempts to evaluate a template tag such as {{with .Form.FieldErrors.title}}.
	data.Form = snippetCreateForm{
		Expires: 365,
	}

	// Render the template code associated with the specified template page.
	app.render(w, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	// r.ParseForm() adds any data in the POST request bodies to the r.PostForm map.
	// This works in the same way for PUT and PATCH requests.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Retrieve the value of the expired field (number of days as an integer) from the r.Postform map.
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Initialize a snippetCreateForm struct to store the form data and any validation errors.
	// Initialize a map to hold any validation errors on the form data.
	form := snippetCreateForm{
		Title:   r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
		Expires: expires,
	}

	// Check that the title is not blank and not more than 100 characters in length.
	form.CheckField(form.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(form.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")

	// Check that the content is not blank.
	form.CheckField(form.NotBlank(form.Content), "content", "This field cannot be blank")

	// Check that the expires value matches one of the permitted values (1, 7, 365).
	form.CheckField(form.PermittedInt(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7, or 365")

	// If there are any validation errors in the form data, dump them into a plain HTTP response and return from the handler.
	if !form.Valid() {
		// Initialize a new templateData struct to store additional resources for the template execution.
		data := app.newTemplateData(r)

		// Pass the snippetCreateForm instance as dynamic data in the Form field.
		data.Form = form

		// Re-render the create.tmpl template in the case of any validation errors.
		// Use the HTTP 422 Unprocessable Entity when sending the response to indicate that their was a form data validation error.
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)

		return
	}

	// Using the parsed values for the client form data, insert a new user into the database using these provided values.
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// After inserting a new user into the database, redirect the user to the viewing page for the snippet they just created.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
