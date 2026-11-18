package task

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	created, err := repo.Create(ctx, Task{Title: "First", Status: StatusPending})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("expected ID 1, got %d", created.ID)
	}

	found, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.Title != "First" {
		t.Fatalf("expected title First, got %q", found.Title)
	}

	created.Title = "Updated"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Updated" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected one task, got %d", len(all))
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

func TestMemoryRepositoryReturnsTasksInIDOrder(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()
	for _, title := range []string{"one", "two", "three"} {
		if _, err := repo.Create(ctx, Task{Title: title, Status: StatusPending}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	got, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	for i, item := range got {
		if item.ID != i+1 {
			t.Fatalf("expected ID %d at index %d, got %d", i+1, i, item.ID)
		}
	}
}

func TestMemoryRepositoryMissingTask(t *testing.T) {
	repo := NewMemoryRepository()
	if _, err := repo.GetByID(context.Background(), 99); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
	if err := repo.Delete(context.Background(), 99); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestMemoryRepositoryHonorsCanceledContext(t *testing.T) {
	repo := NewMemoryRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repo.Create(ctx, Task{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if _, err := repo.GetAll(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMemoryRepositoryConcurrentCreates(t *testing.T) {
	repo := NewMemoryRepository()
	const count = 100

	ids := make(chan int, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			created, err := repo.Create(context.Background(), Task{Title: "concurrent", Status: StatusPending})
			if err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			ids <- created.ID
		}(i)
	}
	wg.Wait()
	close(ids)

	seen := make(map[int]bool)
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		seen[id] = true
	}
	if len(seen) != count {
		t.Fatalf("expected %d unique IDs, got %d", count, len(seen))
	}
}
