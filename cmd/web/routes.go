package main

import (
	"net/http"

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

	// Configure the route for the home page.
	router.HandlerFunc(http.MethodGet, "/", app.home)

	// Configure the route for viewing a snippet with a specified ID.
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", app.snippetView)

	// Configure the route for viewing the form for creating a new snippet via an HTTP GET request.
	router.HandlerFunc(http.MethodGet, "/snippet/create", app.snippetCreate)

	// Configure the route for create a new snippet via an HTTP POST request.
	router.HandlerFunc(http.MethodPost, "/snippet/create", app.snippetCreatePost)

	// Configure the middleware chain for the router, which requests and responses will pass through as they
	// are handled by the server.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return the middleware chain followed by the router.
	return standard.Then(router)
}
