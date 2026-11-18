package task

import (
	"context"
	"errors"
	"testing"
)

func newTestService() *Service {
	return NewService(NewMemoryRepository())
}

func TestServiceCreateDefaultsStatusAndTrimsTitle(t *testing.T) {
	service := newTestService()
	created, err := service.Create(context.Background(), CreateTaskInput{Title: "  Learn Go  "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Title != "Learn Go" {
		t.Fatalf("expected trimmed title, got %q", created.Title)
	}
	if created.Status != StatusPending {
		t.Fatalf("expected default pending status, got %q", created.Status)
	}
}

func TestServiceCreateRejectsInvalidInput(t *testing.T) {
	service := newTestService()
	tests := []struct {
		name  string
		input CreateTaskInput
	}{
		{name: "empty title", input: CreateTaskInput{Title: "   "}},
		{name: "invalid status", input: CreateTaskInput{Title: "Task", Status: "unknown"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := service.Create(context.Background(), tt.input); !errors.Is(err, ErrInvalidTask) {
				t.Fatalf("expected ErrInvalidTask, got %v", err)
			}
		})
	}
}

func TestServicePatchChangesOnlyProvidedFields(t *testing.T) {
	service := newTestService()
	original, err := service.Create(context.Background(), CreateTaskInput{
		Title: "Original", Description: "Keep this", Status: StatusPending,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newTitle := "Updated"
	updated, err := service.Patch(context.Background(), original.ID, TaskPatch{Title: &newTitle})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if updated.Title != "Updated" || updated.Description != "Keep this" || updated.Status != StatusPending {
		t.Fatalf("unexpected patched task: %+v", updated)
	}
}

func TestServicePatchRejectsInvalidValues(t *testing.T) {
	service := newTestService()
	original, err := service.Create(context.Background(), CreateTaskInput{Title: "Original"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	emptyTitle := "  "
	if _, err := service.Patch(context.Background(), original.ID, TaskPatch{Title: &emptyTitle}); !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("expected ErrInvalidTask for empty title, got %v", err)
	}
	invalidStatus := Status("unknown")
	if _, err := service.Patch(context.Background(), original.ID, TaskPatch{Status: &invalidStatus}); !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("expected ErrInvalidTask for invalid status, got %v", err)
	}
}

func TestServiceUpdateAndDeleteMissingTask(t *testing.T) {
	service := newTestService()
	_, err := service.Update(context.Background(), 100, UpdateTaskInput{Title: "Updated", Status: StatusPending})
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
	if err := service.Delete(context.Background(), 100); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
