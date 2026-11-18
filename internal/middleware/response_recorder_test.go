package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseRecorderOnlyWritesFirstStatus(t *testing.T) {
	original := httptest.NewRecorder()
	recorder := newResponseRecorder(original)

	recorder.WriteHeader(http.StatusCreated)
	recorder.WriteHeader(http.StatusInternalServerError)

	if recorder.statusCode != http.StatusCreated {
		t.Fatalf("expected recorded status 201, got %d", recorder.statusCode)
	}
	if !recorder.wroteHeader {
		t.Fatal("expected wroteHeader to be true")
	}
	if original.Code != http.StatusCreated {
		t.Fatalf("expected underlying status 201, got %d", original.Code)
	}
}

func TestResponseRecorderImplicitStatusAndByteCount(t *testing.T) {
	original := httptest.NewRecorder()
	recorder := newResponseRecorder(original)

	n, err := recorder.Write([]byte("Hello"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != 5 || recorder.bytes != 5 {
		t.Fatalf("expected 5 bytes written, got n=%d recorded=%d", n, recorder.bytes)
	}
	if recorder.statusCode != http.StatusOK || !recorder.wroteHeader {
		t.Fatalf("expected implicit 200 and committed headers, got status=%d wroteHeader=%v", recorder.statusCode, recorder.wroteHeader)
	}
}

func TestResponseRecorderUnwrap(t *testing.T) {
	original := httptest.NewRecorder()
	recorder := newResponseRecorder(original)

	if recorder.Unwrap() != original {
		t.Fatal("Unwrap did not return the original ResponseWriter")
	}
}
