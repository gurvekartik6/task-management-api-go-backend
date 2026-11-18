package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	bytes       int
	wroteHeader bool
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader records and forwards only the first status code, matching the
// behavior of net/http response writers.
func (rw *responseRecorder) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}

	// Forward first so an invalid status panic does not make the recorder
	// incorrectly believe that the response has already been committed.
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
	rw.wroteHeader = true
}

// Write records the number of bytes actually accepted by the underlying writer.
// An implicit WriteHeader(200) is recorded when the handler writes a body first.
func (rw *responseRecorder) Write(data []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	n, err := rw.ResponseWriter.Write(data)
	rw.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the original ResponseWriter.
func (rw *responseRecorder) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Logging records request metadata after the downstream handler completes.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := newResponseRecorder(w)

		next.ServeHTTP(recorder, r)

		log.Printf(
			"method=%s path=%s status=%d bytes=%d duration=%s request_id=%s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			recorder.bytes,
			time.Since(start),
			GetRequestID(r.Context()),
		)
	})
}
