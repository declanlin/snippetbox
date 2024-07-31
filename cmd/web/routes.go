package main

import (
	"net/http"

	"github.com/declanlin/snippetbox/ui"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Create a new router to which we will attach middleware, attach handlers to routes, and return to the main function.
	router := httprouter.New()

	// Configure the handler on our router which is to be called when no matching route is found for the specified route.
	// The router will be configured to use our custom error logger (see main.go and helpers.go).
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Take the ui.Files embedded filesystem from the ui package and convert it to an http.FS type so that
	// it satisfies the http.FileSystem interface. Then pass that to the http.FileServer() function to create
	// the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))

	// Our static files are contained in the "static" folder of the ui.Files embedded filesystem.
	// For example, our CSS stylesheet is located at "static/css/main.css"
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// Configure the middleware chain specific to our dynamic application routes.

	// LoadAndSave provides middleware which automatically loads and saves session data for the current request,
	// and communicates the session token to and from the client in a cookie.
	// It checks each incoming request for a session cookie, and if the session cookie is present, it
	// retrieves the corresponding session data from the database (while also checking that your session has not
	// expired), and then adds the session data to the request context to be used in your handlers.
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// Configure the route for the home page.
	// alice.ThenFunc() returns an http.Handler.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))

	// Configure the route for viewing a snippet with a specified ID.
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))

	// Configure the user-related routes.
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protect routes using our custom authentication middleware.
	protected := dynamic.Append(app.requireAuthentication)

	// Configure the route for viewing the form for creating a new snippet via an HTTP GET request.
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	// Configure the route for create a new snippet via an HTTP POST request.
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// Configure the standard middleware chain for the router, which requests and responses will pass through as they
	// are handled by the server.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return the middleware chain followed by the router.
	return standard.Then(router)
}
