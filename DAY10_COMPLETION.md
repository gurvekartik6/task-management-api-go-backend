# Day 10 Completion Record

## Topic
HTTP middleware: request IDs, request logging, panic recovery, integration, and tests.

## Completed implementation

- Request ID middleware creates a cryptographically random request ID, stores it in request context, and returns it in `X-Request-ID`.
- Logging middleware records method, path, response status, response byte count, elapsed time, and request ID.
- Response recorder tracks whether headers have been written, preserving the first status code and correctly recording implicit HTTP 200 responses.
- Response recorder implements `Unwrap` so `http.ResponseController` can reach the original writer.
- Recovery middleware catches panics in the current request goroutine, records a stack trace, and returns a generic HTTP 500 only when the response has not started.
- Middleware is integrated around the router in `cmd/api/main.go` in the order RequestID -> Logging -> Recovery -> Router.
- Task validation errors use `ErrInvalidTask`, allowing handlers to distinguish invalid input (400), missing tasks (404), and unexpected failures (500).
- The in-memory repository checks request context cancellation and protects shared map access with `sync.RWMutex`.
- Handler depends on a service interface, allowing implementations to be substituted in tests.
- README and `.gitignore` added/updated.

## Tests added

- Middleware request ID propagation and uniqueness tests.
- Logging output and response recorder tests.
- Recovery tests before and after response headers are committed.
- Middleware-chain integration test.
- Handler tests for health, create, list, get, update, patch, delete, malformed JSON, unknown fields, oversized bodies, invalid IDs, and missing records.
- Router tests for CRUD routing, malformed input, unknown paths, and unsupported methods.
- Response helper tests for JSON and error responses.
- Repository tests for CRUD, sorting, context cancellation, missing records, and concurrent creates.
- Service tests for default status, title trimming, validation, patch behavior, update, and not-found errors.

## Verification

Run these commands from the project root:

```powershell
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
go test -cover ./...
go build -o task-api.exe ./cmd/api
```

These checks were run in the preparation environment using Go 1.23.2. Formatting, `go vet`, `go test`, `go test -race`, and `go build` completed successfully. Package coverage reported:

- handler: 77.5%
- middleware: 92.5%
- response: 100%
- router: 100%
- task: 82.9%

The `cmd/api` entry point has no direct unit tests; it is exercised through the configured components and can receive lifecycle tests in a later lesson.

## Known scope limitations

- Task records remain in memory and are lost when the process stops.
- Authentication, authorization, PostgreSQL, migrations, and graceful shutdown are not part of this completed Day 9/10 scope; they remain future roadmap work.
- The response recorder supports the current JSON API use cases and exposes `Unwrap`; it does not directly implement streaming/WebSocket optional interfaces such as `http.Flusher` or `http.Hijacker`.

## Next lesson

Day 11: production-quality testing, table-driven tests, dependency injection, and fake repository implementations. We have already added a baseline test suite; Day 11 will deepen the testing approach rather than repeat these tests.
