package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/declanlin/snippetbox/internal/models"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// Define a structure which stores application-specific dependencies for the execution of server-side operations.
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

// Define a function which wraps sql.Open() and returns a sql.DB connection pool for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	// Open a connection to the database with the specified database driver name and DSN.
	db, err := sql.Open("mysql", dsn)

	// Return a nil database pointer if there is an issue connecting to the database.
	if err != nil {
		return nil, err
	}

	// Verify that the connection to the database is still alive.
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Return the connection pool to the main function without any errors.
	return db, nil
}

func main() {
	// flag.String() defines a string flag with the specified name, default value, and usage string.
	// flag.String() returns the address of a string variable which stores the value of the flag.
	addr := flag.String("addr", ":4000", "HTTP Network Address")

	// The DSN string for the snippetbox MYSQL database.
	dsn := flag.String("dsn", "web:Pipluppy2003!@/snippetbox?parseTime=true", "MYSQL Data Source Name")

	// Note: The following SQL statements can be used to create a new database for snippetbox along with
	// a table for snippet objects.

	// -- Create a new UTF-8 `snippetbox` database.
	// CREATE DATABASE snippetbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
	// -- Switch to using the `snippetbox` database.
	// USE snippetbox;

	// -- Create a `snippets` table.
	// CREATE TABLE snippets (
	// id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
	// title VARCHAR(100) NOT NULL,
	// content TEXT NOT NULL,
	// created DATETIME NOT NULL,
	// expires DATETIME NOT NULL
	// );

	// -- Add an index on the created column.
	// CREATE INDEX idx_snippets_created ON snippets(created);

	// After all flags are defined, call flag.Parse() to parse the command line into the defined flags.
	flag.Parse()

	// Define custom error and info loggers for our web application.
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ltime|log.Ldate|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime|log.Ldate)

	// Create a connection pool for the database with the specified DSN, assuming that we have a supported driver
	// for the database.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Defer a call to db.Close() to ensure that the connection pool is closed before the main() function call exits,
	// in the event that a panic occurs.
	defer db.Close()

	// Create a new template cache for the pages we are serving.
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Create a new instance of a *form.Decoder type to be used for decoding HTML form data.
	formDecoder := form.NewDecoder()

	// Create a new instance of a *scs.SessionManager to be used as a session manager for stateful HTTP transactions.
	sessionManager := scs.New()
	// Configure the session manager to use the MYSQL database as the session store, and set a lifetime of 12 hours (so
	// that sessions expire automatically 12 hours after their creation).
	// mysqlstore.New() returns a new MYSQLstore instance with a background cleanup goroutine that runs every 5 minutes
	// to remove expired session data.
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Create an instance of the application structure to store application-specific dependencies for
	// the execution of server-side operations.
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we want the server to use.
	// The only thing we are changing in our case is the curve preferences value, so that only
	// elliptic curves with assembly implementations are used. We are selectively choosing to ignore all
	// other curves beside the ones specified below due to the fact that they are very CPU intensive. Omitting them
	// helps ensure that our server will perform well under heavy load.

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
	}

	// Create an instance of an HTTP server which our application will run on.
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Print an information log to the standard output stream indicating that the server is about to be started.
	infoLog.Printf("Starting server on %s", *addr)

	// ListenAndServe() listens on the TCP network address srv.Addr and then calls Serve() to handle requests
	// on incoming connections.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	// If there is an error listening on the network, log the error. Fatal() is equivalent to errorLog.Println()
	// followed by a call to os.Exit(1).
	errorLog.Fatal(err)
}
