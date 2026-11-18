package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task-management-api/internal/task"
)

func newTestRouter() http.Handler {
	return New(task.NewService(task.NewMemoryRepository()))
}

func TestTaskCRUDRoutes(t *testing.T) {
	handler := newTestRouter()

	createBody := []byte(`{"title":"Learn Go","description":"REST API","status":"pending"}`)
	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", createRecorder.Code, createRecorder.Body.String())
	}

	var created task.Task
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected created task to have an ID")
	}

	getRecorder := httptest.NewRecorder()
	handler.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil))
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", getRecorder.Code)
	}

	patchBody := []byte(`{"status":"completed"}`)
	patchRecorder := httptest.NewRecorder()
	patchRequest := httptest.NewRequest(http.MethodPatch, "/api/tasks/1", bytes.NewReader(patchBody))
	patchRequest.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(patchRecorder, patchRequest)
	if patchRecorder.Code != http.StatusOK {
		t.Fatalf("patch: expected 200, got %d: %s", patchRecorder.Code, patchRecorder.Body.String())
	}

	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodDelete, "/api/tasks/1", nil))
	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", deleteRecorder.Code)
	}

	missingRecorder := httptest.NewRecorder()
	handler.ServeHTTP(missingRecorder, httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil))
	if missingRecorder.Code != http.StatusNotFound {
		t.Fatalf("missing task: expected 404, got %d", missingRecorder.Code)
	}
}

func TestRoutesRejectUnsupportedMethod(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/tasks", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}
}

func TestRoutesReturnNotFoundForUnknownPath(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestRouter().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/unknown", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestCreateRejectsMalformedJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"title":`))
	request.Header.Set("Content-Type", "application/json")
	newTestRouter().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
