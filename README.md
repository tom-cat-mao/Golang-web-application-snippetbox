# Snippetbox - Go Web Application

A production-ready snippet sharing web application built with Go. Users can create, view, and manage code snippets with secure authentication.

## Features

- User authentication (Signup/Login)
- CRUD operations for code snippets
- Session management with secure cookies
- Template caching for fast rendering
- Secure headers middleware
- Database connection pooling
- HTTPS support with modern TLS configuration
- Structured logging
- Mock implementations for testing

## Project Structure

```text
snippetbox/
├── cmd/
│   └── web/                  # Main application entry point
│       ├── context.go        # Context key definitions
│       ├── handlers.go       # HTTP handlers (controller logic)
│       ├── helpers.go        # Template rendering & error helpers
│       ├── main.go           # Server configuration & startup
│       ├── middleware.go     # Authentication/CSRF middleware
│       ├── routes.go         # Route definitions with alice middleware
│       ├── templates.go      # Template cache management
│       └── testutils_test.go # Handler test utilities
├── internal/
│   ├── assert/               # Custom test assertions
│   ├── models/               # Database models and operations
│   │   ├── mocks/           # Mock implementations for testing
│   │   ├── snippets.go      # Snippet model (CRUD operations)
│   │   ├── users.go         # User model (auth/management)
│   │   └── testutils_test.go# Model test database utilities
│   └── validator/           # Custom form validation
├── ui/
│   ├── html/                # HTML templates
│   │   └── pages/          # Page-specific templates
│   └── static/             # Static assets (CSS/JS/images)
└── go.mod                  # Go module dependencies
```

### Key Components

- **HTTP Layer** (`cmd/web`):
  - Route-handler mapping with middleware chaining
  - Session management with SCS
  - Secure headers and CSRF protection
  - Template caching and rendering pipeline

- **Data Layer** (`internal/models`):
  - Database operations for snippets and users
  - Bcrypt password hashing
  - Mock implementations for isolated testing
  - Test database management utilities

- **Validation** (`internal/validator`):
  - Form validation helpers
  - Error message customization
  - Field-specific validation rules

- **UI Assets** (`ui/`):
  - HTML template inheritance system
  - Static file embedding for production
  - Responsive CSS layout
