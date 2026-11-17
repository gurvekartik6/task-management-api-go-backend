package router

import (
	"net/http"

	"task-management-api/internal/handler"
	"task-management-api/internal/task"
)

func New(service *task.Service) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler()
	taskHandler := handler.NewTaskHandler(service)

	// Health endpoint
	mux.HandleFunc("GET /health", healthHandler.Check)

	// Task collection endpoints
	mux.HandleFunc("GET /api/tasks", taskHandler.GetAll)
	mux.HandleFunc("POST /api/tasks", taskHandler.Create)

	// Individual task endpoints
	mux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetByID)
	mux.HandleFunc("PUT /api/tasks/{id}", taskHandler.Update)
	mux.HandleFunc("PATCH /api/tasks/{id}", taskHandler.Patch)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.Delete)

	return mux
}
