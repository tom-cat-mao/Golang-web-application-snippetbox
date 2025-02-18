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

// application struct represents the core application instance.
// It holds all the dependencies and configurations required to run the web server.
// This struct promotes dependency injection, making components easily testable and replaceable.
type application struct {
	logger         *slog.Logger                  // Structured logger for consistent logging.
	snippets       *models.SnippetModel          // Snippet database model for managing snippets.
	templateCache  map[string]*template.Template // In-memory cache for parsed HTML templates.
	formDecoder    *form.Decoder                 // HTML form decoder for processing form data.
	sessionManager *scs.SessionManager           // User session manager for handling user sessions.
	users          *models.UserModel             // User database model for managing users.
}

func main() {
	// Define a flag for the HTTP network address.
	// Default: ":4000" (listen on all interfaces, port 4000).
	// Usage: -addr=":8080" to change the listening port.
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Define a flag for the MySQL Data Source Name (DSN).
	// Format: "username:password@protocol(address)/dbname?param=value".
	// Default: "web:pass@/snippetbox?parseTime=true".
	// Usage: -dsn="user:password@tcp(localhost:3306)/dbname".
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Parse command-line flags.
	// This reads the actual values provided when the program is executed.
	flag.Parse()

	// Create a new structured logger that writes to standard output.
	// It uses a text format for human readability.
	// The default log level is Info.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Open a MySQL database connection using the provided DSN.
	// The openDB function handles the connection and ping verification.
	db, err := openDB(*dsn)
	if err != nil {
		// Log the error at the Error level and exit the program.
		// An exit code of 1 indicates a general error.
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Ensure the database connection is closed when the main function exits.
	// This will execute even if subsequent errors occur.
	defer db.Close()

	// Create a new template cache from all template files in the "ui/html" directory.
	// This improves performance by parsing templates once at startup.
	templateCache, err := newTemplateCache()
	if err != nil {
		// Log the template cache initialization error and exit.
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize a new form decoder for handling HTML form data.
	// The decoder handles URL-encoded and multipart form data.
	formDecoder := form.NewDecoder()

	// Initialize a new session manager using MySQL storage.
	// Session data is stored in the database with a 12-hour lifetime.
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true // Only send cookies over HTTPS.

	// Initialize the application instance with all required dependencies.
	// This creates the core application context that persists throughout the program.
	app := &application{
		logger:         logger,                       // Structured logger.
		snippets:       &models.SnippetModel{DB: db}, // Snippet database model.
		templateCache:  templateCache,                // Template cache.
		formDecoder:    formDecoder,                  // Form decoder.
		sessionManager: sessionManager,               // Session manager.
		users:          &models.UserModel{DB: db},    // User database model.
	}

	// Configure TLS settings for secure communication.
	// This prefers modern, secure elliptic curves for key exchange.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize the HTTP server with configuration.
	// This includes the address, request handler, error logging, and TLS settings.
	srv := &http.Server{
		Addr:         *addr,                                                // Network address to listen on.
		Handler:      app.routes(),                                         // Router/mux for request handling.
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError), // Error logger.
		TLSConfig:    tlsConfig,                                            // TLS configuration for HTTPS.
		IdleTimeout:  time.Minute,                                          // Maximum time to wait for the next request when keep-alives are enabled.
		ReadTimeout:  5 * time.Second,                                      // Maximum duration for reading the entire request, including the body.
		WriteTimeout: 10 * time.Second,                                     // Maximum duration before timing out writes of the response.
	}

	// Log a message indicating that the server is starting.
	// This includes the address the server will listen on.
	logger.Info("starting server", "addr", srv.Addr)

	// Start the HTTP server on the specified address.
	// The routes() method returns the configured router/mux.
	// Start HTTPS server with TLS certificates.
	// Uses self-signed certificates for local development.
	err = srv.ListenAndServeTLS("./tls/localhost.pem", "./tls/localhost-key.pem")

	// If an error occurs while starting the server, log the error and exit.
	// This typically indicates a port conflict or permission issue.
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB establishes and verifies a connection to the MySQL database using the provided DSN.
// It ensures proper resource cleanup in case of connection failures.
//
// Parameters:
//   - dsn: Data Source Name containing connection details.
//     Format: "username:password@protocol(address)/dbname?param=value".
//     Common parameters:
//   - parseTime=true: Parse time.Time values from MySQL.
//   - timeout=30s: Connection timeout duration.
//   - readTimeout=30s: Read operation timeout.
//   - writeTimeout=30s: Write operation timeout.
//
// Returns:
//   - *sql.DB: Database connection handle, ready for query execution.
//   - error: Any error that occurred during connection or verification.
//
// The function performs these steps:
// 1. Opens a new database connection using the MySQL driver.
// 2. Verifies the connection is alive by pinging the database.
// 3. Returns the verified connection handle.
//
// Error Handling:
//   - If the connection fails, returns the original error immediately.
//   - If the ping fails, ensures proper resource cleanup by closing the connection
//     before returning the ping error, preventing connection leaks.
//   - The returned *sql.DB handle is safe for concurrent use and manages
//     a pool of underlying connections automatically.
func openDB(dsn string) (*sql.DB, error) {
	// Open a new database connection using the MySQL driver.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Verify the connection is alive by pinging the database.
	err = db.Ping()
	if err != nil {
		// If the ping fails, close the connection to free resources.
		db.Close()
		return nil, err
	}

	// Return the verified database connection.
	return db, nil
}
