package models

import (
	"database/sql"
	"errors"
	"time"
)

// Snippet represents a single snippet in the database.
// It contains the snippet's ID, title, content, creation time, and expiration time.
type Snippet struct {
	ID      int       // Unique identifier for the snippet
	Title   string    // Title of the snippet
	Content string    // Content of the snippet
	Created time.Time // Time when the snippet was created
	Expires time.Time // Time when the snippet will expire
}

// SnippetModel wraps a sql.DB connection pool and provides methods
// for interacting with the snippets table in the database.
type SnippetModel struct {
	DB *sql.DB // Database connection pool
}

// Insert creates a new snippet record in the database.
// It takes the snippet's title, content, and expiration period (in days) as parameters.
// Returns the ID of the newly created snippet or an error if the operation fails.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, nil
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get retrieves a specific snippet from the database by its ID.
// It returns the snippet if found, or ErrNoRecord if no matching record exists.
// Returns an error if the database operation fails.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`
	row := m.DB.QueryRow(stmt, id)

	var s Snippet

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// Latest retrieves the 10 most recently created snippets from the database.
// It returns a slice of Snippet objects or an error if the database operation fails.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)

		if err = rows.Err(); err != nil {
			return nil, err
		}

	}

	return snippets, nil
}
