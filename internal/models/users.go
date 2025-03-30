package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// UserModelInterface defines the interface for user-related database operations.
// It specifies the methods required for user management, including:
// - User creation and authentication
// - User existence verification
// - User data retrieval
// - Password updates
type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (User, error)
	PasswordUpdate(id int, current_password, new_password string) error
}

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
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

// Get retrieves a user by their ID from the database.
//
// # Parameters
// - id: The ID of the user to retrieve
//
// # Returns
// - User: The retrieved user if found
// - error: nil on success, or:
//   - ErrNoRecord if no user with the given ID exists
//   - Other errors for database failures
func (m *UserModel) Get(id int) (User, error) {
	var user User
	stmt := "SELECT id, name, email, created FROM users WHERE id = ?"

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNoRecord
		} else {
			return User{}, err
		}
	}

	return user, nil
}

// PasswordUpdate updates the password for a user with the given ID.
// It first verifies the current password and then updates it to the new password.
//
// Parameters:
// - id: The ID of the user whose password is to be updated.
// - current_password: The current password of the user.
// - new_password: The new password to set for the user.
//
// Returns:
// - error: nil on success, or an error if the update fails.
func (m *UserModel) PasswordUpdate(id int, current_password, new_password string) error {
	var currentHash []byte

	// Prepare SQL statement to retrieve the current hashed password for the user.
	stmt := "SELECT hashed_password FROM users WHERE id = ?"

	// Execute the query and scan the result into currentHash.
	err := m.DB.QueryRow(stmt, id).Scan(&currentHash)
	if err != nil {
		// If no rows are returned, the user ID is invalid.
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		} else {
			// Return any other error encountered during the query.
			return err
		}
	}

	// Compare the provided current password with the stored hashed password.
	err = bcrypt.CompareHashAndPassword(currentHash, []byte(current_password))
	if err != nil {
		// If the passwords do not match, return an invalid credentials error.
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			// Return any other error encountered during password comparison.
			return err
		}
	}

	// Generate a new hashed password from the new password provided.
	newHash, err := bcrypt.GenerateFromPassword([]byte(new_password), 12)
	if err != nil {
		// Return an error if password hashing fails.
		return err
	}

	// Prepare SQL statement to update the user's password in the database.
	stmt = "UPDATE users SET hashed_password = ? WHERE id = ?"
	// Execute the update statement with the new hashed password.
	_, err = m.DB.Exec(stmt, newHash, id)
	// Return any error encountered during the update.
	return err
}
