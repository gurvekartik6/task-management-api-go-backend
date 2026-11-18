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
	return &MemoryRepository{tasks: make(map[int]Task), nextID: 1}
}

func (r *MemoryRepository) Create(ctx context.Context, newTask Task) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	newTask.ID = r.nextID
	r.nextID++
	r.tasks[newTask.ID] = newTask
	return newTask, nil
}

func (r *MemoryRepository) GetByID(ctx context.Context, id int) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	found, exists := r.tasks[id]
	if !exists {
		return Task{}, ErrTaskNotFound
	}
	return found, nil
}

func (r *MemoryRepository) GetAll(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Task, 0, len(r.tasks))
	for _, current := range r.tasks {
		result = append(result, current)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *MemoryRepository) Update(ctx context.Context, updated Task) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[updated.ID]; !exists {
		return Task{}, ErrTaskNotFound
	}
	r.tasks[updated.ID] = updated
	return updated, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}
