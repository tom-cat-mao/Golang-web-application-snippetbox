package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"snippetbox.tomcat.net/internal/models"

	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// application represents the main application instance containing all dependencies
// and configuration needed to run the web server.
//
// Fields:
//   - logger: Structured logger for application logging
//     Type: *slog.Logger
//     Purpose: Provides consistent logging format and levels
//   - snippets: Database model for snippet operations
//     Type: *models.SnippetModel
//     Purpose: Handles all database interactions for snippets
//   - templateCache: In-memory cache of parsed HTML templates
//     Type: map[string]*template.Template
//     Purpose: Improves performance by caching parsed templates
//   - formDecoder: Form decoder for processing HTML form data
//     Type: *form.Decoder
//     Purpose: Handles form parsing and validation
type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
}

func main() {
	// Define a flag for the HTTP network address to listen on
	// Default: ":4000" (listen on all interfaces, port 4000)
	// Usage: -addr=":8080" to change the listening port
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a flag for the MySQL data source name (DSN)
	// Format: "username:password@protocol(address)/dbname?param=value"
	// Default: "web:pass@/snippetbox?parseTime=true"
	// Usage: -dsn="user:password@tcp(localhost:3306)/dbname"
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Parse the command line flags
	// This reads the actual values from command line arguments
	flag.Parse()

	// Create a new structured logger that outputs to standard output
	// Uses text format for human-readable logs
	// Log level defaults to Info
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Attempt to open a connection to the MySQL database using the provided DSN
	// The openDB function handles both connection and ping verification
	db, err := openDB(*dsn)
	if err != nil {
		// Log the error with Error level and exit the program
		// Exit code 1 indicates a general error
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the database connection is closed when the main function exits
	// This is executed even if an error occurs later in the program
	defer db.Close()

	// Create a new template cache from all template files in the "ui/html" directory
	// The cache improves performance by parsing templates once at startup
	templateCache, err := newTemplateCache()
	if err != nil {
		// Log template cache initialization error and exit
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize a new form decoder for processing HTML form data
	// The decoder handles both URL-encoded and multipart form data
	formDecoder := form.NewDecoder()

	// Initialize the application instance with all required dependencies
	// This creates the central application context used throughout the program
	app := &application{
		logger:        logger,                       // Structured logger
		snippets:      &models.SnippetModel{DB: db}, // Database model
		templateCache: templateCache,                // Template cache
		formDecoder:   formDecoder,                  // Form decoder
	}

	// Log a message indicating that the server is starting
	// Includes the address the server will listen on
	logger.Info("starting server", "addr", *addr)

	// Start the HTTP server on the specified address
	// The routes() method returns the configured router/mux
	err = http.ListenAndServe(*addr, app.routes())

	// If an error occurs while starting the server, log the error and exit
	// This typically indicates a port conflict or permission issue
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB opens and verifies a connection to the MySQL database using the provided DSN.
//
// Parameters:
//   - dsn: Data Source Name containing connection details
//     Format: "username:password@protocol(address)/dbname?param=value"
//
// Returns:
//   - *sql.DB: Database connection handle
//   - error: Any error that occurred during connection or verification
//
// The function performs these steps:
// 1. Opens a new database connection using the MySQL driver
// 2. Verifies the connection is alive by pinging the database
// 3. Returns the verified connection handle
//
// Error Handling:
//   - If connection fails, returns the original error
//   - If ping fails, closes the connection and returns the ping error
func openDB(dsn string) (*sql.DB, error) {
	// Open a new database connection using the MySQL driver
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Verify the connection is alive by pinging the database
	err = db.Ping()
	if err != nil {
		// If ping fails, close the connection to free resources
		db.Close()
		return nil, err
	}

	// Return the verified database connection
	return db, nil
}
