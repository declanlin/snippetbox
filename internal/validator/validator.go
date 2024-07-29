package validator

import (
	"strings"
	"unicode/utf8"
)

// Define a new Validator type which contains a map of validation errors for form data.
type Validator struct {
	FieldErrors map[string]string
}

// Valid() returns true if the FieldErrors map contains no entries.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
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

// CheckField() adds an error message to the FieldError map only if the validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank() returns true if a value is a non-empty string.
func (v *Validator) NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars() returns true if the length of the input string is not greater than the limit n.
func (v *Validator) MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedInt() returns true if a value is in a list of permitted integers.
func (v *Validator) PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}
