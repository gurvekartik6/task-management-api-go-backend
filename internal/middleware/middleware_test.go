package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	var contextID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	headerID := recorder.Header().Get("X-Request-ID")
	if headerID == "" {
		t.Fatal("request ID missing from response header")
	}
	if contextID == "" || contextID != headerID {
		t.Fatalf("context request ID %q does not match response header %q", contextID, headerID)
	}
}

func TestRequestIDsAreUnique(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	seen := make(map[string]bool)
	for i := 0; i < 5; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
		id := recorder.Header().Get("X-Request-ID")
		if id == "" {
			t.Fatal("request ID is missing")
		}
		if seen[id] {
			t.Fatalf("duplicate request ID: %s", id)
		}
		seen[id] = true
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	})

	recorder := httptest.NewRecorder()
	Recovery(panicHandler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "internal server error") {
		t.Fatalf("expected generic error body, got %q", recorder.Body.String())
	}
}

func TestRecoveryDoesNotRewriteCommittedResponse(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("request accepted"))
		panic("panic after response")
	})

	recorder := httptest.NewRecorder()
	Recovery(panicHandler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", recorder.Code)
	}
	if recorder.Body.String() != "request accepted" {
		t.Fatalf("unexpected body: %q", recorder.Body.String())
	}
}

func TestLoggingMiddlewareRecordsRequest(t *testing.T) {
	var output bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previousOutput)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	recorder := httptest.NewRecorder()
	Logging(next).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/tasks", nil))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	for _, expected := range []string{"method=POST", "path=/api/tasks", "status=201", "bytes=7"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("expected log to contain %q, got %q", expected, output.String())
		}
	}
}

func TestMiddlewareChain(t *testing.T) {
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("integration test panic")
	})

	handler := RequestID(Logging(Recovery(panicHandler)))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("request ID missing from response")
	}
	if !strings.Contains(recorder.Body.String(), "internal server error") {
		t.Fatalf("unexpected response body: %q", recorder.Body.String())
	}
}
