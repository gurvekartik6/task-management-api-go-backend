package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

// Compile-time check that the PostgreSQL repository
// implements the existing Repository interface.
var _ Repository = (*PostgresRepository)(nil)

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func scanTask(scanner interface {
	Scan(dest ...any) error
}) (Task, error) {
	var task Task

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
	)

	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	task Task,
) (Task, error) {
	query := `
		INSERT INTO tasks (title, description, status)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, status
	`

	createdTask, err := scanTask(
		r.db.QueryRowContext(
			ctx,
			query,
			task.Title,
			task.Description,
			task.Status,
		),
	)

	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return createdTask, nil
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id int,
) (Task, error) {
	query := `
		SELECT id, title, description, status
		FROM tasks
		WHERE id = $1
	`

	task, err := scanTask(
		r.db.QueryRowContext(ctx, query, id),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}

	if err != nil {
		return Task{}, fmt.Errorf("get task by ID: %w", err)
	}

	return task, nil
}

func (r *PostgresRepository) GetAll(
	ctx context.Context,
) ([]Task, error) {
	query := `
		SELECT id, title, description, status
		FROM tasks
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	task Task,
) (Task, error) {
	query := `
		UPDATE tasks
		SET title = $1,
		    description = $2,
		    status = $3
		WHERE id = $4
		RETURNING id, title, description, status
	`

	updatedTask, err := scanTask(
		r.db.QueryRowContext(
			ctx,
			query,
			task.Title,
			task.Description,
			task.Status,
			task.ID,
		),
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}

	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return updatedTask, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id int,
) error {
	query := `
		DELETE FROM tasks
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted row count: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}
