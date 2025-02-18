package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// EmailRX is a regular expression for validating email addresses.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator struct holds validation errors for a form.
//
// It contains two fields:
//   - FieldErrors:  A map to store validation errors for specific form fields.
//     The key is the field name (string), and the value is the error message (string).
//   - NonFieldErrors: A slice to store general form errors that are not associated with specific fields.
type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// Valid method checks if the Validator has any errors.
//
// It returns true if both FieldErrors and NonFieldErrors are empty,
// indicating no validation errors. Otherwise, it returns false.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddNonFieldError adds a non-field error message to the Validator.
//
// Use this for validation errors that are not associated with a specific form field.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// AddFieldError adds an error message to the FieldErrors map for a given field.
//
//   - If the FieldErrors map is nil, it initializes it.
//   - It only adds the error message if no error already exists for the field,
//     preventing duplicate error messages for the same field.
//
// Parameters:
//   - key:     The name of the form field (e.g., "email", "password").
//   - message: The validation error message to add for the field.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// CheckField method checks if a field passes validation.
// If ok is false, it calls AddFieldError to add the error message.
// key is the field name, and message is the error message.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank function checks if a string is empty or contains only whitespace characters.
// Returns true if the string is not blank, false if it is.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars function checks if the number of characters in a string is less than or equal to n.
// Uses utf8.RuneCountInString to correctly count multi-byte characters.
// Returns true if the number of characters is <= n, false if it is greater than n.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue generic function checks if a value is in the list of permitted values.
// Uses slices.Contains for the check.
// Returns true if the value is in the permitted list, false otherwise.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// MinChars function checks if the number of characters in a string is greater than or equal to n.
// Uses utf8.RuneCountInString to correctly count multi-byte characters.
// Returns true if the number of characters is >= n, false if it is less than n.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Matches function checks if a string matches a regular expression.
// Returns true if the string matches the regular expression, false otherwise.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
