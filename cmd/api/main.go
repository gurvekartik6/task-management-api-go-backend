package main

import (
	"log"
	"net/http"
	"time"

	"task-management-api/internal/router"
	"task-management-api/internal/task"
)

func main() {
	// Initialize repository
	repository := task.NewMemoryRepository()

	// Initialize service
	taskService := task.NewService(repository)

	// Initialize routes
	handler := router.New(taskService)

	// Configure HTTP server
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Server starting on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal("Server failed: ", err)
	}
}
