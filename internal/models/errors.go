// Package models defines the data structures and error types for the application.
// It provides common error definitions used throughout the application.
package models

import "errors"

// ErrNoRecord is returned when a database query does not find a matching record.
//
// Example:
//
//	if errors.Is(err, models.ErrNoRecord) {
//		// Handle case where record was not found
//	}
var (
	ErrNoRecord = errors.New("models: no matching record found")

	// ErrInvalidCredentials is returned when the provided credentials (e.g., password)
	// do not match the expected value.
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail is returned when attempting to create a user with an email
	// address that already exists in the database.
	ErrDuplicateEmail = errors.New("models: duplicate emails")
)
