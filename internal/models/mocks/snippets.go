// Mocking is essential in unit testing because it allows us to isolate the unit of code we want to test (in this case, our handlers) from its dependencies (in this case, the database/model layer).

// By using a mock SnippetModel, we can:

// 1. Control the behavior of the SnippetModel: We can predefine the return values of the `Get` and `Insert` methods, simulating different scenarios (e.g., successful retrieval, not found error, database error) without needing a real database.
// 2. Avoid side effects: We don't need to interact with a real database, which could be slow, require setup/teardown, or modify data. This makes our tests faster, more reliable, and prevents unintended changes to the database.
// 3. Test specific cases: We can easily test edge cases or error conditions that might be difficult to reproduce with a real database.
// 4. Isolate failures: If a test fails, we know the problem is in the handler, not in the database interaction, because the database interaction is replaced by our simplified mock.

package mocks

import (
	"time"

	"snippetbox.tomcat.net/internal/models"
)

// Mock data that mimics a real database entry.
var mockSnippet = models.Snippet{
	ID:      1,
	Title:   "An old silent pond",
	Content: "An old silent pond...",
	Created: time.Now(),
	Expires: time.Now(),
}

type SnippetModel struct{}

// Mock the Insert method.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	return 2, nil
}

// Mock the Get method.
// It simulates fetching a snippet by ID.
// If the ID is 1, it returns a predefined mock snippet.
// Otherwise, it returns an ErrNoRecord error, indicating that no record was found.
func (m *SnippetModel) Get(id int) (models.Snippet, error) {
	switch id {
	case 1:
		return mockSnippet, nil
	default:
		return models.Snippet{}, models.ErrNoRecord
	}
}

// Mock the Latest method.
// It simulates fetching the 10 most recently created snippets.
// It returns a slice containing the mock snippet.
func (m *SnippetModel) Latest() ([]models.Snippet, error) {
	return []models.Snippet{mockSnippet}, nil
}
