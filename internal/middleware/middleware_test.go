package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := RequestID(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			id := GetRequestID(r.Context())

			if id == "" {
				t.Error("request ID missing from context")
			}

			w.WriteHeader(http.StatusOK)
		},
	))

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	requestID := recorder.Header().Get("X-Request-ID")

	if requestID == "" {
		t.Fatal("request ID missing from response header")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		},
	)

	handler := Recovery(panicHandler)

	request := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"internal server error",
	) {
		t.Error("expected generic internal server error response")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	next := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		},
	)

	handler := Logging(next)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/tasks",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status 201, got %d",
			recorder.Code,
		)
	}
}
