package task

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"task-management-api/internal/database"
)

func TestPostgresRepositoryCRUD(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not configured; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	defer db.Close()

	// Apply the project migration to the dedicated test database.
	migration, err := os.ReadFile(
		"../../migrations/001_create_tasks.sql",
	)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	// Clear test data before the test.
	if _, err := db.ExecContext(
		ctx,
		"TRUNCATE TABLE tasks RESTART IDENTITY",
	); err != nil {
		t.Fatalf("clear test table: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		_, _ = db.ExecContext(
			cleanupCtx,
			"TRUNCATE TABLE tasks RESTART IDENTITY",
		)
	})

	repo := NewPostgresRepository(db)

	// 1. Create
	created, err := repo.Create(ctx, Task{
		Title:       "Learn PostgreSQL",
		Description: "Integrate PostgreSQL with Go",
		Status:      StatusPending,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("expected generated ID, got %d", created.ID)
	}

	if created.Title != "Learn PostgreSQL" {
		t.Errorf("unexpected title: %q", created.Title)
	}

	// 2. GetByID
	found, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if found.Description != "Integrate PostgreSQL with Go" {
		t.Errorf("unexpected description: %q", found.Description)
	}

	// 3. GetAll
	tasks, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("get all tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	// 4. Update
	found.Title = "Learn PostgreSQL properly"
	found.Status = StatusInProgress

	updated, err := repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	if updated.Title != "Learn PostgreSQL properly" {
		t.Errorf("unexpected updated title: %q", updated.Title)
	}

	if updated.Status != StatusInProgress {
		t.Errorf("unexpected updated status: %q", updated.Status)
	}

	// 5. Delete
	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	// 6. Verify deleted task cannot be retrieved.
	_, err = repo.GetByID(ctx, created.ID)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}

	// 7. Verify deleting a missing task returns not found.
	err = repo.Delete(ctx, created.ID)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound on delete, got %v", err)
	}
}
