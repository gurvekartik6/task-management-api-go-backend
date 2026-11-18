# Task Management REST API (Go)

A beginner-to-production-oriented Go REST API built with the standard library. The project currently uses an in-memory repository; PostgreSQL persistence is planned for a later roadmap day.

## Requirements

- Go 1.23 or later
- Git (optional, for version control)

The project uses Go 1.22+ method-aware `http.ServeMux` patterns.

## Run locally

Run commands from the directory containing `go.mod`:

```powershell
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go run ./cmd/api
```

The API listens on `http://localhost:8080` by default. Set the `PORT` environment variable to use another port.

Build an executable:

```powershell
go build -o task-api.exe ./cmd/api
```

## Project structure

```text
cmd/api/main.go                 Application entry point and HTTP server setup
internal/handler/               HTTP handlers and request decoding
internal/middleware/            Request ID, logging and panic recovery
internal/response/              Shared JSON response helpers
internal/router/                Method-aware route registration
internal/task/                  Task model, service and repository abstraction
```

## Architecture

The request flow is:

```text
HTTP request
  -> Request ID middleware
  -> Logging middleware
  -> Recovery middleware
  -> Router
  -> Handler
  -> Task service
  -> Repository interface
  -> In-memory repository
```

- **Handler:** translates HTTP input/output and status codes.
- **Service:** owns task validation and business rules.
- **Repository:** abstracts persistence behind an interface.
- **MemoryRepository:** thread-safe in-memory implementation using `sync.RWMutex`.
- **Middleware:** provides cross-cutting request processing.

## Middleware

1. Request ID generates a cryptographically random ID, stores it in request context and returns it in `X-Request-ID`.
2. Logging records method, path, status, response bytes, duration and request ID.
3. Recovery catches panics in the current request goroutine, logs a stack trace and returns a generic 500 response if the response has not already been committed.

## API endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/api/tasks` | List tasks |
| POST | `/api/tasks` | Create task |
| GET | `/api/tasks/{id}` | Get one task |
| PUT | `/api/tasks/{id}` | Replace editable task fields |
| PATCH | `/api/tasks/{id}` | Update supplied task fields |
| DELETE | `/api/tasks/{id}` | Delete task |

### Create task example

```powershell
$body = @{
    title = "Learn Go"
    description = "Build a REST API"
    status = "pending"
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/api/tasks -Method POST -ContentType "application/json" -Body $body
```

Allowed statuses: `pending`, `in_progress`, `completed`. When status is omitted during creation, it defaults to `pending`.

## Testing

```powershell
go test ./...
go test -race ./...
go test -cover ./...
```

Tests cover middleware behavior, response recording, repository CRUD and concurrency, service validation and business rules, and HTTP routes.

## Current limitations

- Data is stored in memory and is lost when the process stops.
- Authentication and authorization are not implemented yet.
- Pagination and database persistence are planned for later roadmap work.
- The response recorder supports the current JSON API use cases and exposes `Unwrap`; streaming/WebSocket-specific optional interfaces are not implemented.
