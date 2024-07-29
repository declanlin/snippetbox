package main

import (
	"fmt"
	"net/http"
)

// A middleware which can be attached to a router to automatically add HTTP security headers to every response,
// inline with the current OWASP guidance.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CSP headers are used to restrict where the resources for your web page (e.g. Javascript, images, fonts, etc.)
		// are allowed to be loaded from.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		// Referrer-Policy is used to control what information is included in the Referrer header when a user navigates
		// away from your web page. We have the value set to "origin-when-cross-origin", which means the full URL will be
		// included for same-origin requests, but for all other requests information like the URL path and query string
		// values will be stripped out.
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")

		// X-Content-Type-Options: nosniff instructs browsers to not MIME-type sniff the contenttype of the response,
		// which in turn helps to prevent content-sniffing attacks.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: deny is used to help prevent clickjacking attacks in older browsers that
		// don’t support CSP headers.
		w.Header().Set("X-Frame-Options", "deny")

		// X-XSS-Protection: 0 is used to disable the blocking of cross-site scripting attacks.
		// Previously it was good practice to set this header to X-XSS-Protection: 1; mode=block ,
		// but when you’re using CSP headers like we are the recommendation is to disable this
		// feature altogether.
		w.Header().Set("X-XSS-Protection", "0")

		// Proceed with handling the request, passing control to the next middleware or to the final handler.
		next.ServeHTTP(w, r)
	})
}

// A middleware which can be attached to a router to log information about incoming HTTP requests.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the formatted HTTP request information.
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		// Proceed with handling the request, passing control to the next middleware or to the final handler.
		next.ServeHTTP(w, r)
	})
}

// A middleware which can be attached to a router to recover from server-side panics.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The following deferred function will execute any time a panic occurs during the execution of next.ServeHTTP(w, r).
		// It will instruct the client to close their connection with the server and log an error message.
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		// Proceed with handling the request, passing control to the next middleware or to the final handler.
		next.ServeHTTP(w, r)
	})
}
