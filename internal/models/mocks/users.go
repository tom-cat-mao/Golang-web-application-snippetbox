package mocks

import (
	"time"

	"snippetbox.tomcat.net/internal/models"
)

type UserModel struct{}

// Mock the Insert method.
// It simulates a successful user insertion and a duplicate email scenario.
// If the provided email is "dupe@example.com", it returns an ErrDuplicateEmail error.
// Otherwise, it simulates a successful insertion and returns nil (no error).
func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

// Mock the Authenticate method.
// It simulates a successful authentication and invalid credentials scenario.
// If the provided email is "alice@example.com" and the password is "pa$$word", it returns a user ID of 1.
// Otherwise, it returns an ErrInvalidCredentials error.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return 1, nil
	}

	return 0, models.ErrInvalidCredentials
}

// Mock the Exists method.
// It simulates checking if a user exists by ID.
// If the ID is 1, it returns true (user exists).
// Otherwise, it returns false (user does not exist).
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

// Get mocks the retrieval of a user by ID.
// It simulates two scenarios:
// - If the ID is 1, returns a mock user with ID 1, email "alice@example.com", name "Alice", and current timestamp
// - For any other ID, returns an empty User and ErrNoRecord to simulate a non-existent user
func (m *UserModel) Get(id int) (models.User, error) {
	switch id {
	case 1:
		return models.User{
			ID:      1,
			Email:   "alice@example.com",
			Name:    "Alice",
			Created: time.Now(),
		}, nil
	default:
		return models.User{}, models.ErrNoRecord
	}
}

func (m *UserModel) PasswordUpdate(id int, current_password, new_password string) error {
	if id == 1 {
		if current_password != "pa$$word" {
			return models.ErrInvalidCredentials
		}

		return nil
	}

	return models.ErrNoRecord
}
