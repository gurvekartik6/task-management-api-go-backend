package task

import "context"

// fakeRepository is a test double for the Repository interface.
// Each function can be configured independently by a test.
type fakeRepository struct {
	createFn  func(context.Context, Task) (Task, error)
	getByIDFn func(context.Context, int) (Task, error)
	getAllFn  func(context.Context) ([]Task, error)
	updateFn  func(context.Context, Task) (Task, error)
	deleteFn  func(context.Context, int) error
}

// Compile-time check: fakeRepository must implement Repository.
var _ Repository = (*fakeRepository)(nil)

func (f *fakeRepository) Create(
	ctx context.Context,
	task Task,
) (Task, error) {
	if f.createFn != nil {
		return f.createFn(ctx, task)
	}

	return task, nil
}

func (f *fakeRepository) GetByID(
	ctx context.Context,
	id int,
) (Task, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}

	return Task{}, ErrTaskNotFound
}

func (f *fakeRepository) GetAll(
	ctx context.Context,
) ([]Task, error) {
	if f.getAllFn != nil {
		return f.getAllFn(ctx)
	}

	return nil, nil
}

func (f *fakeRepository) Update(
	ctx context.Context,
	task Task,
) (Task, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, task)
	}

	return task, nil
}

func (f *fakeRepository) Delete(
	ctx context.Context,
	id int,
) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}

	return nil
}
