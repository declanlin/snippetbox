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
// Include struct tags which tell the decoder how to store the value from the HTML form data.
// The struct tag `form:"-"` tells the decoder to completely ignore a field during decoding.

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
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

	// Declare a new empty instance of a snippetCreateForm struct to store the form data and a validator.
	var form snippetCreateForm

	// Decode the relevant values from the HTML form into the snippetCreateForm struct.
	err := app.decodePostForm(r, &form)
	if err != nil {
		// The client entered form data that was not valid.
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Check that the title is not blank and not more than 100 characters in length.
	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")

	// Check that the content is not blank.
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")

	// Check that the expires value matches one of the permitted values (1, 7, 365).
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7, or 365")

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

	// Use the Put() function to add a string value and corresponding key to the session data.
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	// After inserting a new user into the database, redirect the user to the viewing page for the snippet they just created.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// Render and display the signup form for the client.
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {

	// Initialize a new templateData struct to store additional resources for the template execution.
	data := app.newTemplateData(r)

	// Intialize the data.Form field as a zeroed userSignupForm instance.
	data.Form = userSignupForm{}

	// Render the template for the signup.tmpl template.
	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Declare a zeroed instance of the userSignupForm struct to store form data and access the validator.
	var form userSignupForm

	// Decode the HTML form data into the userSignupForm struct.
	// Return an HTTP 400 Response if the user attempts to sign up with data that cannot be decoded.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form fields.

	// Check that the name and email are not blank.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")

	// Check that the email address is in a valid format.
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")

	// Check that the password is not blank and at least 8 characters long.
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	// If there are any validation errors in the form data, dump them into a plain HTTP response and return from the handler.
	if !form.Valid() {
		// Initialize a new templateData struct to store additional resources for the template execution.
		data := app.newTemplateData(r)

		// Pass the userSignupForm instance as dynamic data in the Form field.
		data.Form = form

		// Re-render the singup.tmpl template in the case of any validation errors.
		// Use the HTTP 422 Unprocessable Entity when sending the response to indicate that their was a form data validation error.
		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)

		return
	}

	// Attempt to create a new user in the database.
	// If there is a duplicate email error, add an error message to the form and redisplay it.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Add a confirmation flash message to the session confirming their signup worked.
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	// Redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Declare a zeroed instance of the userLoginForm struct to store form data and access the validator.
	var form userLoginForm

	// Decode the HTML form data into the userLoginForm struct.
	// Return an HTTP 400 Response if the user attempts to log in with data that cannot be decoded.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the login form data.
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	// If any of the form data is not valid, re-render the login page with the errors (stored in the session data).
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
	}

	// Authenticate the user credentials. If the credentials are invalid, add a generic non-field error message
	// and re-display the login page.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Incorrect email or password")

			// Re-display the login page after modifying the form in the template data.
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusOK, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Use the RenewToken() method on the current session to change the session ID.
	// It's good practice to generate a new session ID when the authentication state or privilege level changes
	// for the user, e.g. login and logout operations.
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the ID of the current user to the session so that they are considered "logged in".
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Redirect the logged in user to the snippet create page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session ID to change the session ID.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Remove the authenticatedUserID from the session data so that the user is "logged out".
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Add a flash message indicating that the user has been successfully logged out.
	app.sessionManager.Put(r.Context(), "flash", "You have been logged out successfully!")

	// Redirect the user to the application homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
