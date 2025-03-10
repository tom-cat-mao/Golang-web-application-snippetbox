package models

import (
	"database/sql"
	"os"
	"testing"
)

// newTestDB initializes a test database connection pool and executes setup/teardown SQL scripts.
// It:
//   - Opens connection to MySQL database with test credentials and parameters
//   - Executes setup.sql to create tables and test data
//   - Registers cleanup function to:
//   - Execute teardown.sql to drop tables and reset state
//   - Close database connection
//   - Returns the database connection pool for use in tests
//   - Fails the test immediately if any setup operations fail
func newTestDB(t *testing.T) *sql.DB {
	// Establish a sql.DB connection pool for our test database. Because our
	// setup and teardown scripts use multi-line statements, we need to
	// use the "multiStatements=true" parameter in our DSN. This instructs
	// our MySQL database driver to support executing multiple SQL statements
	// in one db.Exec() call.
	db, err := sql.Open("mysql", "test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true")
	if err != nil {
		t.Fatal(err)
	}

	// Read the setup SQL script from the file and execute the statements, closing
	// the connection and calling t.Fatal() in the event of an error.
	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	// Use t.Cleanup() to register a function which will automatically be called
	// by Go when the current test (or sub-test) which calls newTestDB()
	// has finished. In this function, we read and execute the teardown script,
	// and close the database connection pool.
	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	})

	// Return the database connection pool
	return db
}
