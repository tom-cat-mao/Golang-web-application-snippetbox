// test helper assert function
package assert

import (
	"strings"
	"testing"
)

// Equal is a generic test helper function that compares two values of the same comparable type.
// It takes a testing.T instance and two values to compare.
// If the actual value does not equal the expected value, it fails the test with an error message.
// Parameters:
//   - t: The testing.T instance for reporting test failures
//   - actual: The value being tested
//   - expected: The value to compare against
func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper() // indicate to go this is a test helper

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

// StringContains is a test helper that checks if a string contains an expected substring.
// If the actual string does not contain the expected substring, it fails the test with an error message.
func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubstring)
	}
}

func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got: %v: expected: nil", actual)
	}
}
