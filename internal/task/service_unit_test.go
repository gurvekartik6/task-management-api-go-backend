package task

import (
	"context"
	"errors"
	"testing"
)

func TestServiceCreateTableDriven(t *testing.T) {
	repositoryErr := errors.New("simulated storage failure")

	tests := []struct {
		name          string
		input         CreateTaskInput
		repositoryErr error
		wantErr       error
		wantTitle     string
		wantStatus    Status
		wantRepoCall  bool
	}{
		{
			name: "valid task with default status",
			input: CreateTaskInput{
				Title:       "  Learn Go  ",
				Description: "Study testing",
			},
			wantTitle:    "Learn Go",
			wantStatus:   StatusPending,
			wantRepoCall: true,
		},
		{
			name: "valid task with explicit status",
			input: CreateTaskInput{
				Title:  "Build API",
				Status: StatusInProgress,
			},
			wantTitle:    "Build API",
			wantStatus:   StatusInProgress,
			wantRepoCall: true,
		},
		{
			name: "empty title",
			input: CreateTaskInput{
				Title: "",
			},
			wantErr:      ErrInvalidTask,
			wantRepoCall: false,
		},
		{
			name: "whitespace-only title",
			input: CreateTaskInput{
				Title: "   ",
			},
			wantErr:      ErrInvalidTask,
			wantRepoCall: false,
		},
		{
			name: "invalid status",
			input: CreateTaskInput{
				Title:  "Learn Go",
				Status: Status("unknown"),
			},
			wantErr:      ErrInvalidTask,
			wantRepoCall: false,
		},
		{
			name: "repository failure",
			input: CreateTaskInput{
				Title: "Save task",
			},
			repositoryErr: repositoryErr,
			wantErr:       repositoryErr,
			wantRepoCall:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositoryCalled := false
			var receivedTask Task

			repo := &fakeRepository{
				createFn: func(
					ctx context.Context,
					task Task,
				) (Task, error) {
					repositoryCalled = true
					receivedTask = task

					if tt.repositoryErr != nil {
						return Task{}, tt.repositoryErr
					}

					return task, nil
				},
			}

			service := NewService(repo)

			got, err := service.Create(
				context.Background(),
				tt.input,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.wantErr,
						err,
					)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if repositoryCalled != tt.wantRepoCall {
				t.Fatalf(
					"repository called = %v, want %v",
					repositoryCalled,
					tt.wantRepoCall,
				)
			}

			if !tt.wantRepoCall {
				return
			}

			if tt.repositoryErr != nil {
				return
			}

			if receivedTask.Title != tt.wantTitle {
				t.Errorf(
					"title = %q, want %q",
					receivedTask.Title,
					tt.wantTitle,
				)
			}

			if receivedTask.Status != tt.wantStatus {
				t.Errorf(
					"status = %q, want %q",
					receivedTask.Status,
					tt.wantStatus,
				)
			}

			if got.Title != tt.wantTitle {
				t.Errorf(
					"returned title = %q, want %q",
					got.Title,
					tt.wantTitle,
				)
			}
		})
	}
}

func TestServiceUpdateTableDriven(t *testing.T) {
	repositoryErr := errors.New("update storage failure")

	tests := []struct {
		name       string
		input      UpdateTaskInput
		getErr     error
		updateErr  error
		wantErr    error
		wantUpdate bool
		wantTitle  string
		wantStatus Status
	}{
		{
			name: "successful update",
			input: UpdateTaskInput{
				Title:       "  Updated task  ",
				Description: "Updated description",
				Status:      StatusCompleted,
			},
			wantUpdate: true,
			wantTitle:  "Updated task",
			wantStatus: StatusCompleted,
		},
		{
			name: "task does not exist",
			input: UpdateTaskInput{
				Title:  "Updated task",
				Status: StatusPending,
			},
			getErr:     ErrTaskNotFound,
			wantErr:    ErrTaskNotFound,
			wantUpdate: false,
		},
		{
			name: "invalid title",
			input: UpdateTaskInput{
				Title:  "   ",
				Status: StatusPending,
			},
			wantErr:    ErrInvalidTask,
			wantUpdate: false,
		},
		{
			name: "invalid status",
			input: UpdateTaskInput{
				Title:  "Updated task",
				Status: Status("invalid"),
			},
			wantErr:    ErrInvalidTask,
			wantUpdate: false,
		},
		{
			name: "repository update failure",
			input: UpdateTaskInput{
				Title:  "Updated task",
				Status: StatusPending,
			},
			updateErr:  repositoryErr,
			wantErr:    repositoryErr,
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			var updatedTask Task

			repo := &fakeRepository{
				getByIDFn: func(
					ctx context.Context,
					id int,
				) (Task, error) {
					if tt.getErr != nil {
						return Task{}, tt.getErr
					}

					return Task{
						ID:          id,
						Title:       "Original task",
						Description: "Original description",
						Status:      StatusPending,
					}, nil
				},
				updateFn: func(
					ctx context.Context,
					task Task,
				) (Task, error) {
					updateCalled = true
					updatedTask = task

					if tt.updateErr != nil {
						return Task{}, tt.updateErr
					}

					return task, nil
				},
			}

			service := NewService(repo)

			got, err := service.Update(
				context.Background(),
				10,
				tt.input,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.wantErr,
						err,
					)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if updateCalled != tt.wantUpdate {
				t.Fatalf(
					"update called = %v, want %v",
					updateCalled,
					tt.wantUpdate,
				)
			}

			if !tt.wantUpdate || tt.updateErr != nil {
				return
			}

			if updatedTask.ID != 10 {
				t.Errorf(
					"ID = %d, want 10",
					updatedTask.ID,
				)
			}

			if updatedTask.Title != tt.wantTitle {
				t.Errorf(
					"title = %q, want %q",
					updatedTask.Title,
					tt.wantTitle,
				)
			}

			if updatedTask.Status != tt.wantStatus {
				t.Errorf(
					"status = %q, want %q",
					updatedTask.Status,
					tt.wantStatus,
				)
			}

			if got.Title != tt.wantTitle {
				t.Errorf(
					"returned title = %q, want %q",
					got.Title,
					tt.wantTitle,
				)
			}
		})
	}
}

func TestServicePatchTableDriven(t *testing.T) {
	repositoryErr := errors.New("patch storage failure")

	tests := []struct {
		name       string
		patch      TaskPatch
		getErr     error
		updateErr  error
		wantErr    error
		wantUpdate bool
		wantTitle  string
		wantStatus Status
	}{
		{
			name: "update only title",
			patch: TaskPatch{
				Title: stringPointer("New title"),
			},
			wantUpdate: true,
			wantTitle:  "New title",
			wantStatus: StatusPending,
		},
		{
			name: "update only status",
			patch: TaskPatch{
				Status: statusPointer(StatusCompleted),
			},
			wantUpdate: true,
			wantTitle:  "Original task",
			wantStatus: StatusCompleted,
		},
		{
			name: "update title and description",
			patch: TaskPatch{
				Title:       stringPointer("Changed title"),
				Description: stringPointer("Changed description"),
			},
			wantUpdate: true,
			wantTitle:  "Changed title",
			wantStatus: StatusPending,
		},
		{
			name: "empty title is invalid",
			patch: TaskPatch{
				Title: stringPointer("   "),
			},
			wantErr:    ErrInvalidTask,
			wantUpdate: false,
		},
		{
			name: "invalid status",
			patch: TaskPatch{
				Status: statusPointer(Status("unknown")),
			},
			wantErr:    ErrInvalidTask,
			wantUpdate: false,
		},
		{
			name: "task not found",
			patch: TaskPatch{
				Title: stringPointer("Changed"),
			},
			getErr:     ErrTaskNotFound,
			wantErr:    ErrTaskNotFound,
			wantUpdate: false,
		},
		{
			name: "repository update failure",
			patch: TaskPatch{
				Title: stringPointer("Changed"),
			},
			updateErr:  repositoryErr,
			wantErr:    repositoryErr,
			wantUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			var updatedTask Task

			repo := &fakeRepository{
				getByIDFn: func(
					ctx context.Context,
					id int,
				) (Task, error) {
					if tt.getErr != nil {
						return Task{}, tt.getErr
					}

					return Task{
						ID:          id,
						Title:       "Original task",
						Description: "Original description",
						Status:      StatusPending,
					}, nil
				},
				updateFn: func(
					ctx context.Context,
					task Task,
				) (Task, error) {
					updateCalled = true
					updatedTask = task

					if tt.updateErr != nil {
						return Task{}, tt.updateErr
					}

					return task, nil
				},
			}

			service := NewService(repo)

			got, err := service.Patch(
				context.Background(),
				10,
				tt.patch,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.wantErr,
						err,
					)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if updateCalled != tt.wantUpdate {
				t.Fatalf(
					"update called = %v, want %v",
					updateCalled,
					tt.wantUpdate,
				)
			}

			if !tt.wantUpdate || tt.updateErr != nil {
				return
			}

			if updatedTask.Title != tt.wantTitle {
				t.Errorf(
					"title = %q, want %q",
					updatedTask.Title,
					tt.wantTitle,
				)
			}

			if updatedTask.Status != tt.wantStatus {
				t.Errorf(
					"status = %q, want %q",
					updatedTask.Status,
					tt.wantStatus,
				)
			}

			if got.ID != 10 {
				t.Errorf("ID = %d, want 10", got.ID)
			}
		})
	}
}

// Small helpers for pointer fields in TaskPatch.
func stringPointer(value string) *string {
	return &value
}

func statusPointer(value Status) *Status {
	return &value
}

func TestServiceRepositoryErrorPropagation(t *testing.T) {
	repositoryErr := errors.New("repository unavailable")

	t.Run("GetAll", func(t *testing.T) {
		repo := &fakeRepository{
			getAllFn: func(
				ctx context.Context,
			) ([]Task, error) {
				return nil, repositoryErr
			},
		}

		service := NewService(repo)

		_, err := service.GetAll(context.Background())

		if !errors.Is(err, repositoryErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		repo := &fakeRepository{
			getByIDFn: func(
				ctx context.Context,
				id int,
			) (Task, error) {
				return Task{}, repositoryErr
			},
		}

		service := NewService(repo)

		_, err := service.GetByID(context.Background(), 10)

		if !errors.Is(err, repositoryErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		repo := &fakeRepository{
			deleteFn: func(
				ctx context.Context,
				id int,
			) error {
				return repositoryErr
			},
		}

		service := NewService(repo)

		err := service.Delete(context.Background(), 10)

		if !errors.Is(err, repositoryErr) {
			t.Fatalf("expected repository error, got %v", err)
		}
	})
}
