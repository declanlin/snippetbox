package models

import "errors"

// Custom error for when an sql row query returns no matching records.
var ErrNoRecord = errors.New("models: no matching record found")
