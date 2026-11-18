package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTask  = errors.New("invalid task")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func isValidStatus(status Status) bool {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	default:
		return false
	}
}

func validateTask(title string, status Status) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidTask)
	}
	if !isValidStatus(status) {
		return fmt.Errorf("%w: invalid task status", ErrInvalidTask)
	}
	return nil
}

func (s *Service) Create(ctx context.Context, input CreateTaskInput) (Task, error) {
	if input.Status == "" {
		input.Status = StatusPending
	}
	if err := validateTask(input.Title, input.Status); err != nil {
		return Task{}, err
	}

	newTask := Task{
		Title:       strings.TrimSpace(input.Title),
		Description: input.Description,
		Status:      input.Status,
	}
	return s.repository.Create(ctx, newTask)
}

func (s *Service) GetAll(ctx context.Context) ([]Task, error) {
	return s.repository.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int) (Task, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, input UpdateTaskInput) (Task, error) {
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return Task{}, err
	}
	if err := validateTask(input.Title, input.Status); err != nil {
		return Task{}, err
	}

	existing.Title = strings.TrimSpace(input.Title)
	existing.Description = input.Description
	existing.Status = input.Status
	return s.repository.Update(ctx, existing)
}

func (s *Service) Patch(ctx context.Context, id int, patch TaskPatch) (Task, error) {
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return Task{}, err
	}

	if patch.Title != nil {
		existing.Title = strings.TrimSpace(*patch.Title)
	}
	if patch.Description != nil {
		existing.Description = *patch.Description
	}
	if patch.Status != nil {
		existing.Status = *patch.Status
	}

	if err := validateTask(existing.Title, existing.Status); err != nil {
		return Task{}, err
	}
	return s.repository.Update(ctx, existing)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}
