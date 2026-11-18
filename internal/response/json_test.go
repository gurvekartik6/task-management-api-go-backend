package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONWritesStatusHeaderAndBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	JSON(recorder, http.StatusCreated, map[string]string{"message": "created"})

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type: %q", recorder.Header().Get("Content-Type"))
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
	if body["message"] != "created" {
		t.Fatalf("unexpected response body: %#v", body)
	}
}

func TestErrorWritesErrorMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	Error(recorder, http.StatusBadRequest, "invalid input")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["error"] != "invalid input" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}
