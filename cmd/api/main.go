package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"task-management-api/internal/middleware"
	"task-management-api/internal/router"
	"task-management-api/internal/task"
)

func main() {
	// Initialize repository
	repository := task.NewMemoryRepository()

	// Initialize service
	taskService := task.NewService(repository)

	// Initialize router
	apiHandler := router.New(taskService)

	// Apply middleware
	handler := middleware.RequestID(
		middleware.Logging(
			middleware.Recovery(apiHandler),
		),
	)

	// Configure HTTP server. PORT can be set per environment; 8080 is the default.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server starting on http://localhost:%s", port)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal("Server failed: ", err)
	}
}
