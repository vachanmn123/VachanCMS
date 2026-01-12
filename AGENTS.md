# VachanCMS - Agent Handbook

## Overview
VachanCMS is a headless Content Management System (CMS) built in Go using the Gin web framework. It leverages GitHub repositories as the primary backend storage mechanism for content, allowing users to manage structured data, media files, and configurations directly within GitHub repos. This approach provides version control, collaboration features, and scalability inherent to Git repositories.

The CMS is designed for headless use, with a planned VueJS admin frontend for content creation and editing. The actual CMS content is consumed by clients via static site generation, hosted on GitHub Pages within the same repository.

## Key Features
- **Authentication**: Integrates with GitHub OAuth 2.0 for user login. Users authenticate via GitHub, and their access tokens are encrypted and stored in JWTs.
- **Repository Management**: Users can list their GitHub repositories and initialize them for CMS use by creating a configuration file and directory structure.
- **Content Types**: Define custom content types with fields (e.g., text, number, boolean, select). Each content type has a unique slug and is stored in the repo's `config/config.json`.
- **Content Values**: Create, read, update, and list content instances based on defined types. Data is paginated and stored as JSON files in `data/<content-type-slug>/` with indexes for efficient retrieval.
- **Media Management**: Upload and serve media files (e.g., images, documents) stored in the repo's `media/` directory, with similar pagination and indexing.
- **API-Driven**: All operations are exposed via RESTful API endpoints, making it suitable for headless CMS use where a separate frontend consumes the API.
- **Security**: Uses JWT for session management, encrypts GitHub access tokens, and relies on GitHub's permissions for access control.

## Project Structure
The codebase follows a clean architecture with separation of concerns:

- **`config/`**: Handles environment variable loading and configuration (e.g., GitHub OAuth credentials, JWT secrets).
- **`handlers/`**: Contains HTTP request handlers for authentication, repository operations, content types, content values, and media.
- **`middleware/`**: Includes authentication middleware to protect routes and validate JWTs.
- **`models/`**: Defines data structures for configurations, content types, content values, and media files.
- **`routes/`**: Sets up Gin router with protected and unprotected endpoints.
- **`services/`**: Provides business logic, including JWT handling, GitHub API interactions, and CMS-specific operations.
- **`main.go`**: Entry point that initializes the server.
- **`go.mod` and `go.sum`**: Go module dependencies (e.g., Gin, GitHub API client, JWT library).
- **`.gitignore`**: Standard Go ignores, including `.env` files.
- **`tests.js`**: JavaScript snippets for testing API endpoints via fetch requests (e.g., creating content, uploading media).
- **`AGENTS.md`**: This file, providing an overview for future developers.

## How It Works
1. **Setup and Authentication**:
   - The server loads configuration from a `.env` file.
   - Users initiate login via `/auth/login`, which redirects to GitHub OAuth.
   - Upon callback, a JWT is generated containing the encrypted GitHub access token and stored in a cookie.

2. **Repository Initialization**:
   - Authenticated users can list their repos via `/repos`.
   - For a selected repo (`/:owner/:repo`), they fetch or initialize the config via `/config` and `/init`, creating `config/config.json`, `content/.gitkeep`, and `media/.gitkeep`.

3. **Content Management**:
   - **Content Types**: Create and list types via `/content-types`. Each type defines fields with types and validation rules.
   - **Content Values**: For a type (`/:ctSlug`), create/update values via POST/PUT, list with pagination via GET, and retrieve by ID.
   - Data is stored as JSON files (e.g., `data/posts/123.json` for a content value, with indexes like `data/posts/index-1.json`).

4. **Media Handling**:
   - Upload files via POST to `/media`, storing them in `media/<id>` with metadata in indexes.
   - List and retrieve media with pagination.

5. **Storage and Persistence**:
   - All data is committed to the GitHub repo as files, using the GitHub API for CRUD operations.
   - Pagination is handled via index files (e.g., `index-1.json`) and config files tracking total items/pages.

6. **Security and Middleware**:
   - Protected routes use `AuthMiddleware` to validate JWTs and extract user context.
   - GitHub access tokens are encrypted in JWTs to prevent exposure.

## Technologies and Dependencies
- **Go 1.25.4**: Core language.
- **Gin**: Web framework for routing and HTTP handling.
- **GitHub API**: Via `go-github` library for repo and file operations.
- **JWT**: For authentication tokens.
- **OAuth2**: For GitHub login flow.
- **AES Encryption**: To secure access tokens in JWTs.
- **godotenv**: For environment variable management.

## Important Notes
- **Headless Nature**: No built-in frontend; it's designed for API consumption (e.g., by a React app or static site generator).
- **GitHub as Storage**: Content is versioned and collaborative but may have rate limits or costs for large repos.
- **Validation**: Content values are validated against their type definitions (e.g., field types, required fields).
- **Pagination**: Implemented for content and media lists to handle large datasets efficiently.
- **Testing**: Includes a `tests.js` file with example API calls for manual testing.
- **Production Readiness**: Includes flags for production mode, but lacks features like rate limiting, caching, or advanced error handling.
- **Frontend**: Planned VueJS admin interface for content management.
- **Content Consumption**: Static sites generated from repo data and served via GitHub Pages.
- **CRUD Limitations**: Deletes are not implemented yet to avoid pagination issues; full CRUD planned.
- **Performance**: Caching is planned for frequently fetched files to improve speed and reduce API calls.

## Roadmap and Future Plans
- **Frontend Development**: Build a VueJS admin interface for creating/editing data.
- **Deployment**: Plan for hosting the Go server (e.g., Docker, cloud services).
- **Full CRUD**: Implement deletes with pagination handling (e.g., soft deletes or index rebuilding).
- **Caching**: Add in-memory or Redis caching for configs, indexes, and file contents to reduce GitHub API latency.
- **Static Site Generation**: Develop tools to generate GitHub Pages-compatible sites from CMS data.
- **Additional Features**: User roles, webhooks, integrations with SSGs like Hugo/Gatsby.

## Getting Started
1. Clone the repo and set up Go 1.25.4.
2. Create a `.env` file with required variables (e.g., GitHub OAuth credentials, JWT secret).
3. Run `go mod tidy` and `go run main.go`.
4. Use `tests.js` for API testing or build the Vue frontend.

For questions or contributions, refer to this document or the codebase comments. Happy coding!
