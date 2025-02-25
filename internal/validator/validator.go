package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// EmailRX is a compiled regular expression used for validating email addresses according to RFC 5322.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator struct encapsulates validation errors for a form, providing a structured way to manage and report errors.
//
// Fields:
//   - NonFieldErrors: A slice of strings to hold general form errors not tied to specific fields.
//   - FieldErrors: A map where keys are field names (string) and values are corresponding error messages (string), used for field-specific validation errors.
type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// Valid method determines if the Validator instance contains any errors.
//
// Returns true if both FieldErrors and NonFieldErrors are empty, indicating no validation errors exist. Otherwise, returns false.
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddNonFieldError appends a non-field error message to the Validator's NonFieldErrors slice.
//
// This method is used for adding validation errors that do not pertain to any specific form field.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// AddFieldError adds an error message to the FieldErrors map for a specified field.
//
// This method initializes the FieldErrors map if it's nil. It adds the error message only if the field does not already have an error, avoiding duplicate entries.
//
// Parameters:
//   - key: The name of the form field (e.g., "email", "password").
//   - message: The error message to associate with the field.
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// CheckField method validates a field and adds an error if validation fails.
//
// If the validation check (ok) fails, it invokes AddFieldError to record the error.
// Parameters:
//   - ok: Boolean indicating whether the field passed validation.
//   - key: The name of the field being validated.
//   - message: The error message to add if validation fails.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank function verifies if a string contains non-whitespace characters.
//
// Returns true if the string is not blank (contains at least one non-whitespace character), false otherwise.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars function checks if a string's character count does not exceed a specified limit.
//
// Utilizes utf8.RuneCountInString for accurate counting of multi-byte characters.
// Returns true if the string's character count is less than or equal to n, false otherwise.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue generic function verifies if a value is included in a list of permitted values.
//
// Employs slices.Contains to perform the check.
// Returns true if the value matches any in the permittedValues list, false otherwise.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// MinChars function checks if a string's character count meets or exceeds a specified minimum.
//
// Uses utf8.RuneCountInString to accurately count multi-byte characters.
// Returns true if the string's character count is greater than or equal to n, false otherwise.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Matches function tests if a string conforms to a given regular expression.
//
// Returns true if the string matches the provided regular expression, false otherwise.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
