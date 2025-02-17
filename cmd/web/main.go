package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"snippetbox.tomcat.net/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// application struct represents the core application instance containing all dependencies and configurations needed to run the web server.
// It follows the dependency injection pattern, making all components easy to test and replace.
//
// Field descriptions:
//   - logger: Structured logger
//     Type: *slog.Logger
//     Purpose: Provides structured logging at different severity levels (Info, Debug, Error, etc.) to improve observability
//   - snippets: Snippet database model
//     Type: *models.SnippetModel
//     Purpose: Handles all CRUD operations for snippets, abstracts database interactions, and provides a clean API for business logic
//   - templateCache: In-memory cache for HTML templates
//     Type: map[string]*template.Template
//     Purpose: Improves performance by caching parsed templates, reducing disk I/O for subsequent requests. Uses base.html for template inheritance
//   - formDecoder: HTML form decoder
//     Type: *form.Decoder
//     Purpose: Handles form parsing and validation, supports URL-encoded and multipart form data, allows custom validation rules
//   - sessionManager: User session manager
//     Type: *scs.SessionManager
//     Purpose: Manages user session data including authentication state, uses MySQL for session storage to enable persistence and scalability
type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Define flag for HTTP network address
	// Default: ":4000" (listen on all interfaces, port 4000)
	// Usage: -addr=":8080" to change listening port
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define flag for MySQL Data Source Name (DSN)
	// Format: "username:password@protocol(address)/dbname?param=value"
	// Default: "web:pass@/snippetbox?parseTime=true"
	// Usage: -dsn="user:password@tcp(localhost:3306)/dbname"
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Parse command line flags
	// Reads actual values from command line arguments
	flag.Parse()

	// Create a new structured logger that writes to standard output
	// Uses text format for human readability
	// Default log level is Info
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Attempt to open MySQL database connection using provided DSN
	// openDB function handles connection and ping verification
	db, err := openDB(*dsn)
	if err != nil {
		// Log error at Error level and exit program
		// Exit code 1 indicates general error
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Ensure database connection is closed when main function exits
	// Will execute even if subsequent errors occur
	defer db.Close()

	// Create new template cache from all template files in "ui/html" directory
	// Improves performance by parsing templates once at startup
	templateCache, err := newTemplateCache()
	if err != nil {
		// Log template cache initialization error and exit
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize new form decoder for handling HTML form data
	// Decoder handles URL-encoded and multipart form data
	formDecoder := form.NewDecoder()

	// Initialize new session manager using MySQL storage
	// Session data is stored in database with 12-hour lifetime
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true // Only send cookies over HTTPS

	// Initialize application instance with all required dependencies
	// Creates core application context that persists throughout the program
	app := &application{
		logger:         logger,                       // Structured logger
		snippets:       &models.SnippetModel{DB: db}, // Database model
		templateCache:  templateCache,                // Template cache
		formDecoder:    formDecoder,                  // Form decoder
		sessionManager: sessionManager,               // Session manager
	}

	// Configure TLS settings for secure communication
	// Prefers modern, secure elliptic curves for key exchange
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize HTTP server with configuration
	// Includes address, request handler, error logging, and TLS settings
	srv := &http.Server{
		Addr:      *addr,                                                // Network address to listen on
		Handler:   app.routes(),                                         // Router/mux for request handling
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError), // Error logger
		TLSConfig: tlsConfig,                                            // TLS configuration for HTTPS
	}

	// Log a message indicating that the server is starting
	// Includes the address the server will listen on
	logger.Info("starting server", "addr", srv.Addr)

	// Start the HTTP server on the specified address
	// The routes() method returns the configured router/mux
	// Start HTTPS server with TLS certificates
	// Uses self-signed certificates for local development
	err = srv.ListenAndServeTLS("./tls/localhost.pem", "./tls/localhost-key.pem")

	// If an error occurs while starting the server, log the error and exit
	// This typically indicates a port conflict or permission issue
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB establishes and verifies a connection to the MySQL database using the provided DSN.
// It implements proper resource cleanup in case of connection failures.
//
// Parameters:
//   - dsn: Data Source Name containing connection details
//     Format: "username:password@protocol(address)/dbname?param=value"
//     Common parameters:
//   - parseTime=true: Parse time.Time values from MySQL
//   - timeout=30s: Connection timeout duration
//   - readTimeout=30s: Read operation timeout
//   - writeTimeout=30s: Write operation timeout
//
// Returns:
//   - *sql.DB: Database connection handle, ready for query execution
//   - error: Any error that occurred during connection or verification
//
// The function performs these steps:
// 1. Opens a new database connection using the MySQL driver
// 2. Verifies the connection is alive by pinging the database
// 3. Returns the verified connection handle
//
// Error Handling:
//   - If connection fails, returns the original error immediately
//   - If ping fails, ensures proper resource cleanup by closing the connection
//     before returning the ping error, preventing connection leaks
//   - The returned *sql.DB handle is safe for concurrent use and manages
//     a pool of underlying connections automatically
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
