// test helper assert function
package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper() // indicate to go this is a test helper

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
