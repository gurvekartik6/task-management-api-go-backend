package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"task-management-api/internal/response"
	"task-management-api/internal/task"
)

type TaskService interface {
	Create(ctx context.Context, input task.CreateTaskInput) (task.Task, error)
	GetAll(ctx context.Context) ([]task.Task, error)
	GetByID(ctx context.Context, id int) (task.Task, error)
	Update(ctx context.Context, id int, input task.UpdateTaskInput) (task.Task, error)
	Patch(ctx context.Context, id int, patch task.TaskPatch) (task.Task, error)
	Delete(ctx context.Context, id int) error
}

type TaskHandler struct {
	service TaskService
}

func NewTaskHandler(service TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

func writeTaskServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, task.ErrTaskNotFound):
		response.Error(w, http.StatusNotFound, "task not found")
	case errors.Is(err, task.ErrInvalidTask):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("request must contain only one JSON value")
		}

		return err
	}

	return nil
}

func parseTaskID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("invalid task ID")
	}

	return id, nil
}

func (h *TaskHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	tasks, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "could not retrieve tasks")
		return
	}

	response.JSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input task.CreateTaskInput

	if err := decodeJSON(w, r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	created, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeTaskServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, created)
}

func (h *TaskHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseTaskID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	found, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			response.Error(w, http.StatusNotFound, "task not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, "could not retrieve task")
		return
	}

	response.JSON(w, http.StatusOK, found)
}

func (h *TaskHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseTaskID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var input task.UpdateTaskInput

	if err := decodeJSON(w, r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	updated, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		writeTaskServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *TaskHandler) Patch(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseTaskID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var raw map[string]json.RawMessage

	if err := decodeJSON(w, r, &raw); err != nil || raw == nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request")
		return
	}

	var patch task.TaskPatch

	for key, value := range raw {
		switch key {
		case "title":
			if err := json.Unmarshal(value, &patch.Title); err != nil || patch.Title == nil {
				response.Error(w, http.StatusBadRequest, "title must be a string")
				return
			}

		case "description":
			if err := json.Unmarshal(value, &patch.Description); err != nil || patch.Description == nil {
				response.Error(w, http.StatusBadRequest, "description must be a string")
				return
			}

		case "status":
			if err := json.Unmarshal(value, &patch.Status); err != nil || patch.Status == nil {
				response.Error(w, http.StatusBadRequest, "status must be a string")
				return
			}

		default:
			response.Error(w, http.StatusBadRequest, "unknown field: "+key)
			return
		}
	}

	if len(raw) == 0 {
		response.Error(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	updated, err := h.service.Patch(r.Context(), id, patch)
	if err != nil {
		writeTaskServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, updated)
}

func (h *TaskHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := parseTaskID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, task.ErrTaskNotFound) {
			response.Error(w, http.StatusNotFound, "task not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, "could not delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
