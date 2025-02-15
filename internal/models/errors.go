// Package models defines the data structures and error types used by the application's
// data access layer. It provides common error definitions that are used throughout
// the application's database operations.
package models

import "errors"

// ErrNoRecord is returned when a database query returns no matching records.
// This error is typically used when attempting to fetch a specific record that
// doesn't exist in the database.
//
// Example usage:
//
//	if errors.Is(err, models.ErrNoRecord) {
//	    // Handle case where record was not found
//	}
var ErrNoRecord = errors.New("models: no matching record found")
