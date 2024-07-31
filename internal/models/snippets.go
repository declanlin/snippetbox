package models

import (
	"database/sql"
	"errors"
	"time"
)

// Define a Snippet type to hold data for an individual Snippet.
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps an sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// Define a function that will insert a new snippet into the MYSQL database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// Generate an SQL statement for inserting a new snippet into the database.
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Use the Exec() method on the embedded connection pool to execute the SQL statement.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, nil
	}

	// Use the LastInsertID() from the Result interface on the result returned by Exec(), which returns
	// the integer generated by the database in response to executing a command, typically the
	// AUTO_INCREMENT command when inserting a new row.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Return the ID of the snippet (converted from int64 to int) along with no errors.
	return int(id), nil
}

// Define a function that will read and return a specified snippet based on its unique ID.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	// Generate an SQL statement for selecting a snippet from the database according to a given ID.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// Query a single row by calling QueryRow() on our connection pool.
	row := m.DB.QueryRow(stmt, id)

	// Initialize a pointer to a zeroed Snippet struct.
	s := &Snippet{}

	// Use row.Scan() to copy in columns from the queried row to the corresponding fields in the Snippet struct s.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		// Check if the query returns no rows using the errors.Is() function.
		// We will return our own ErrNoRecord error (see internal/models/errors.go).
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	// Return the address of the Snippet struct with no errors.
	return s, nil
}

// Define a function that will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	// Generate an SQL statement for selecting the 10 most recently created snippets.
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// Query multiple rows by calling Query() on our connection pool.
	// Query() returns an sql.Rows resultset containing the result of our query.
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// Defer a call to rows.Close() to ensure that the sql.Rows resultset is closed before
	// the Latest() function returns.
	defer rows.Close()

	// Initialize an empty slice to hold pointers to Snippet structs.
	snippets := []*Snippet{}

	// Iterate over each of the rows in the resultset.
	for rows.Next() {
		// Initialize a pointer to a zeroed Snippet struct.
		s := &Snippet{}

		// Use row.Scan() to copy in columns from the queried row to the corresponding fields in the Snippet struct s.
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		// Apend the snippet to the slice of snippets.
		snippets = append(snippets, s)
	}

	// Retrieve any error encountered during the iteration above.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Return the queried rows as a slice of Snippet struct pointers with no errors.
	return snippets, nil
}

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}
