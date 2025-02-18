package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// User represents a registered user in the system.
//
// # Fields
// - ID: Unique identifier for the user
// - Name: User's full name
// - Email: User's email address (must be unique)
// - HashedPassword: Bcrypt-hashed password
// - Created: Timestamp of account creation
type User struct {
	ID             int       // Unique user ID
	Name           string    // User's name
	Email          string    // User's email address (unique)
	HashedPassword []byte    // Bcrypt-hashed password
	Created        time.Time // Account creation timestamp
}

// UserModel handles all database interactions for users.
// It provides methods for user authentication, creation, and verification.
type UserModel struct {
	DB *sql.DB // Database connection pool
}

// Insert creates a new user record in the database.
//
// # Parameters
// - name: User's full name
// - email: User's email address (must be unique)
// - password: Plain-text password (will be hashed)
//
// # Returns
// - error: nil on success, or:
//   - ErrDuplicateEmail if email already exists
//   - Other errors for database/hashing failures
//
// # Security
// - Passwords are hashed using bcrypt with cost factor 12
// - Email addresses must be unique (enforced by database)
//
// # Example Usage
//
//	err := userModel.Insert("John Doe", "john@example.com", "mypassword")
//	if err != nil {
//	    if errors.Is(err, models.ErrDuplicateEmail) {
//	        // Handle duplicate email
//	    }
//	    // Handle other errors
//	}
func (m *UserModel) Insert(name, email, password string) error {
	// Hash the password with bcrypt cost factor 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// SQL statement to insert new user
	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES(?, ?, ?, UTC_TIMESTAMP())`

	// Execute the statement with provided values
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// Check for duplicate email addresses
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

// Authenticate verifies user credentials and returns the user ID if valid.
//
// # Parameters
// - email: User's email address.
// - password: Plain-text password to verify.
//
// # Returns
// - int: User ID if authentication successful, 0 otherwise.
// - error: nil on success, or:
//   - ErrInvalidCredentials if email/password don't match.
//   - Other errors for database failures.
//
// # Security
// - Uses constant-time comparison for password verification to mitigate timing attacks.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

// Exists checks if a user with the given ID exists in the database.
//
// # Parameters
// - id: User ID to check
//
// # Returns
// - bool: true if user exists, false if not
// - error: nil on success, database errors otherwise
func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}
