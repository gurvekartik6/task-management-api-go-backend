package task

import (
	"context"
	"sort"
	"sync"
)

type MemoryRepository struct {
	mu     sync.RWMutex
	tasks  map[int]Task
	nextID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks:  make(map[int]Task),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(
	ctx context.Context,
	task Task,
) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++

	r.tasks[task.ID] = task

	return task, nil
}

func (r *MemoryRepository) GetByID(
	ctx context.Context,
	id int,
) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return Task{}, ErrTaskNotFound
	}

	return task, nil
}

func (r *MemoryRepository) GetAll(
	ctx context.Context,
) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}

func (r *MemoryRepository) Update(
	ctx context.Context,
	task Task,
) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return Task{}, ErrTaskNotFound
	}

	r.tasks[task.ID] = task

	return task, nil
}

func (r *MemoryRepository) Delete(
	ctx context.Context,
	id int,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)

	return nil
}
