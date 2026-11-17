package task

import "context"

type Repository interface {
	Create(ctx context.Context, task Task) (Task, error)
	GetByID(ctx context.Context, id int) (Task, error)
	GetAll(ctx context.Context) ([]Task, error)
	Update(ctx context.Context, task Task) (Task, error)
	Delete(ctx context.Context, id int) error
}
