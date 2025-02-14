package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"snippetbox.tomcat.net/internal/models"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	// Define a flag for the HTTP network address to listen on
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a flag for the MySQL data source name (DSN)
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Parse the command line flags
	flag.Parse()

	// Create a new logger that outputs to standard output in text format
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Attempt to open a connection to the MySQL database using the provided DSN
	db, err := openDB(*dsn)
	if err != nil {
		// Log the error and exit the program with status code 1
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the database connection is closed when the main function exits
	defer db.Close()

	// Create a new template cache from all template files in the "ui/html" directory
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize an application instance with the configured logger, snippet model, and template cache
	app := &application{
		logger:        logger,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	// Log a message indicating that the server is starting, including the address it will listen on
	logger.Info("starting server", "addr", *addr)

	// Start the HTTP server on the specified address and serve the application routes
	err = http.ListenAndServe(*addr, app.routes())

	// If an error occurs while starting the server, log the error and exit the program with status code 1
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB opens a connection to the MySQL database using the provided DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
