package mocks

import "snippetbox.tomcat.net/internal/models"

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
