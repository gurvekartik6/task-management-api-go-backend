package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"task-management-api/internal/database"
	"task-management-api/internal/middleware"
	"task-management-api/internal/router"
	"task-management-api/internal/task"
)

func main() {

	startupCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")

	db, err := database.Open(startupCtx, databaseURL)
	if err != nil {
		log.Fatalf("database startup failed: %v", err)
	}
	defer db.Close()

	taskRepository := task.NewPostgresRepository(db)
	taskService := task.NewService(taskRepository)

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
