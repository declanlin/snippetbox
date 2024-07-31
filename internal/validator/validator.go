package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Define a new Validator type which contains a map of validation errors for form data.
type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// Valid() returns true if the FieldErrors map contains no entries.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddFieldError() adds a [form field name : error message] mapping to the FieldErrors map,
// so long as an entry does not already exist.
func (v *Validator) AddFieldError(key, message string) {
	// Initialize the FieldErrors map if it does not yet exist.
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	// Check to see if an entry exists for the current key (form field name).
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// AddNonFieldError adds an error message to the NonFieldErrors string array,
// which stores error messages for validation errors not related to the form field data.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// CheckField() adds an error message to the FieldError map only if the validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank() returns true if a value is a non-empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars() returns true if the length of the input string is not greater than the limit n.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// MinChars() returns true if the length of the input string is at least n.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// PermittedInt() returns true if a value is in a list of permitted integers.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// Regex expression to validate the format of an email string.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validates a string against a regex expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
