# Golang-web-application-snippetbox

## Project Structure

This project is a Golang web application named Snippetbox. It is structured as follows:

- **cmd**: Contains the main applications. In this project, it contains the `web` application.
    - **cmd/web**: Contains the source code for the web application.
        - `context.go`: Defines context keys used throughout the application.
        - `handlers.go`: Defines HTTP handler functions for different routes of the application, such as creating snippets, user signup, and login.
        - `handlers_test.go`: Contains tests for the HTTP handlers to ensure they function correctly.
        - `helpers.go`: Includes helper functions used by the handlers and other parts of the web application for common tasks like rendering templates and handling server errors.
        - `main.go`: The entry point of the web application. It sets up and starts the HTTP server, initializes dependencies like the database connection and template cache.
        - `middleware.go`: Defines middleware functions for request processing, such as setting common headers.
        - `routes.go`: Defines the application's routes and maps them to the corresponding handler functions using a ServeMux.
        - `templates.go`: Handles template management, including caching and rendering HTML templates. It also defines the `templateData` struct for passing data to templates.
        - `templates_test.go`: Contains tests for template-related functionalities.
        - `testutils_test.go`: Provides utility functions and types to support testing of HTTP handlers and middleware.

- **internal**: Contains internal packages that are not intended for use by external code.
    - **internal/assert**: Provides custom assertion functions to simplify and standardize testing.
    - **internal/models**: Contains the data models and database interaction logic.
        - **internal/models/mocks**: Includes mock implementations of the model interfaces, primarily used for testing purposes to isolate components and avoid database dependencies in tests.
        - `internal/models/snippets.go`: Defines the `Snippet` model and the `SnippetModel` struct with methods for interacting with the snippets table in the database. Implements `SnippetModelInterface`.
        - `internal/models/testutils_test.go`: Provides utilities for setting up and tearing down test databases for model testing.
        - `internal/models/users.go`: Defines the `User` model and the `UserModel` struct for handling user-related database operations. Implements `UserModelInterface`.
    - **internal/validator**: Implements a custom validator for handling and checking form data.

- **ui**: Contains user interface assets.
    - **ui/static**: Holds static files such as CSS stylesheets, JavaScript files, and images.
    - **ui/html**: Contains HTML templates used to render the web pages.

This structure helps to organize the codebase by separating concerns into different directories and packages, making the project more maintainable and understandable.
