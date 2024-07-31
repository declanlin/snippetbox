package models

import "errors"

// Custom error for when an sql row query returns no matching records.
var ErrNoRecord = errors.New("models: no matching record found")

// Custom error for when a user attempts to login with an invalid email or invalid password.
var ErrInvalidCredentials = errors.New("models: invalid credentials")

// Custom error for when a user attempts to sign up with an email address that is already being used.
var ErrDuplicateEmail = errors.New("models: duplicate email")
