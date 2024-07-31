package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Define a User type to hold data for an individual User.
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
	Created        time.Time
}

// Define a UserModel type which wraps an sql.DB connection pool.
type UserModel struct {
	DB *sql.DB
}

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
}

// Define a function that will insert a new user into the MYSQL database.
func (m *UserModel) Insert(name, email, password string) error {
	// Hash the password that the user wants to sign up with a cost of 12.
	// The cost of 12 entails (2^12=4096) bcrypt iterations to generate the hash.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Generate an SQL statement to insert a new user into our users table.
	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES (?, ?, ?, UTC_TIMESTAMP())`

	// Execute the SQL statement to insert a new user into the users table.
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))

	// If an error occurs executing the SQL statement, check if the error has the type *mysql.MySQLError.
	// If it does, the error will be assigned to the mySQLError variable.
	// Check whether or not the error relates to the users_uc_email key by checking if the SQL error code equals
	// 1062 (ERR_DUP_ENTRY) and the contents of the error message string.
	// If it does, return an ErrDuplicateEmail error (see internal/models/errors.go).
	if err != nil {
		var mySQLError *mysql.MySQLError

		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		// Return all other types of errors as is.
		return err
	}

	// Return without errors once the user has been created successfully in the database.
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the ID and hashed password associated with the given email.
	var id int
	var hashedPassword []byte

	// Generate an SQL statement for selecting user information for a matching email record.
	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`

	// Execute the SQL statment.
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Check whether the hashed password and plaintext password match.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// If the user's email and password are authenticated successfully, return the user's ID with no errors.
	return id, nil
}

// Function to check if a user with a specific ID exists in our database.
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS(SELECT true FROM users WHERE id = ?)`

	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}
