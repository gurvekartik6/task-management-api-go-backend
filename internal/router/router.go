package router

import (
	"net/http"

	"task-management-api/internal/handler"
)

func New(service handler.TaskService) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler()
	taskHandler := handler.NewTaskHandler(service)

	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/tasks", taskHandler.GetAll)
	mux.HandleFunc("POST /api/tasks", taskHandler.Create)
	mux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetByID)
	mux.HandleFunc("PUT /api/tasks/{id}", taskHandler.Update)
	mux.HandleFunc("PATCH /api/tasks/{id}", taskHandler.Patch)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.Delete)

	return mux
}
