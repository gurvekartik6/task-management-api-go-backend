package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"task-management-api/internal/task"
)

func newTestTaskHandler() *TaskHandler {
	return NewTaskHandler(task.NewService(task.NewMemoryRepository()))
}

func requestWithTaskID(method, path, body string, handler http.HandlerFunc) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.SetPathValue("id", strings.TrimPrefix(path, "/api/tasks/"))
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestCreateTaskHandler(t *testing.T) {
	h := newTestTaskHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"title":"Learn Go"}`))
	h.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var created task.Task
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if created.ID != 1 || created.Status != task.StatusPending {
		t.Fatalf("unexpected created task: %+v", created)
	}
}

func TestCreateTaskRejectsMalformedUnknownAndOversizedJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed", body: `{"title":`},
		{name: "unknown field", body: `{"title":"Task","unexpected":true}`},
		{name: "multiple JSON values", body: `{"title":"Task"} {"title":"Second"}`},
		{name: "empty body", body: ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestTaskHandler()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(tt.body))
			h.Create(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}

	h := newTestTaskHandler()
	recorder := httptest.NewRecorder()
	largeBody := `{"title":"` + strings.Repeat("x", (1<<20)+1) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(largeBody))
	h.Create(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for oversized body, got %d", recorder.Code)
	}
}

func TestCreateTaskRejectsInvalidBusinessInput(t *testing.T) {
	h := newTestTaskHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{"title":"  ","status":"pending"}`))
	h.Create(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestGetAllTasksHandler(t *testing.T) {
	h := newTestTaskHandler()
	if _, err := h.service.Create(context.Background(), task.CreateTaskInput{Title: "One"}); err != nil {
		t.Fatalf("seed task: %v", err)
	}
	recorder := httptest.NewRecorder()
	h.GetAll(recorder, httptest.NewRequest(http.MethodGet, "/api/tasks", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestGetTaskByIDHandler(t *testing.T) {
	h := newTestTaskHandler()
	created, err := h.service.Create(context.Background(), task.CreateTaskInput{Title: "One"})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}

	tests := []struct {
		name string
		id   string
		want int
	}{
		{name: "valid", id: "1", want: http.StatusOK},
		{name: "invalid ID", id: "abc", want: http.StatusBadRequest},
		{name: "missing task", id: "999", want: http.StatusNotFound},
	}
	_ = created
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := requestWithTaskID(http.MethodGet, "/api/tasks/"+tt.id, "", h.GetByID)
			if recorder.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, recorder.Code)
			}
		})
	}
}

func TestUpdateTaskHandler(t *testing.T) {
	h := newTestTaskHandler()
	_, err := h.service.Create(context.Background(), task.CreateTaskInput{Title: "Original"})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}

	recorder := requestWithTaskID(http.MethodPut, "/api/tasks/1", `{"title":"Updated","description":"New","status":"in_progress"}`, h.Update)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	invalid := requestWithTaskID(http.MethodPut, "/api/tasks/1", `{"title":"","status":"pending"}`, h.Update)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid update, got %d", invalid.Code)
	}

	missing := requestWithTaskID(http.MethodPut, "/api/tasks/999", `{"title":"Updated","status":"pending"}`, h.Update)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing task, got %d", missing.Code)
	}
}

func TestPatchTaskHandler(t *testing.T) {
	h := newTestTaskHandler()
	_, err := h.service.Create(context.Background(), task.CreateTaskInput{Title: "Original"})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}

	tests := []struct {
		name string
		body string
		want int
	}{
		{name: "valid partial update", body: `{"status":"completed"}`, want: http.StatusOK},
		{name: "empty object", body: `{}`, want: http.StatusBadRequest},
		{name: "unknown field", body: `{"unknown":"value"}`, want: http.StatusBadRequest},
		{name: "null field", body: `{"title":null}`, want: http.StatusBadRequest},
		{name: "invalid JSON", body: `not-json`, want: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := requestWithTaskID(http.MethodPatch, "/api/tasks/1", tt.body, h.Patch)
			if recorder.Code != tt.want {
				t.Fatalf("expected %d, got %d: %s", tt.want, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestDeleteTaskHandler(t *testing.T) {
	h := newTestTaskHandler()
	_, err := h.service.Create(context.Background(), task.CreateTaskInput{Title: "To delete"})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}

	deleted := requestWithTaskID(http.MethodDelete, "/api/tasks/1", "", h.Delete)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", deleted.Code)
	}

	missing := requestWithTaskID(http.MethodDelete, "/api/tasks/1", "", h.Delete)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", missing.Code)
	}
}
