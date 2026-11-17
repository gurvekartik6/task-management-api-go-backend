package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseRecorder) WriteHeader(statusCode int) {
	if rw.statusCode != http.StatusOK || rw.bytes > 0 {
		return
	}

	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseRecorder) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytes += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		start := time.Now()

		recorder := newResponseRecorder(w)

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		requestID := GetRequestID(r.Context())

		log.Printf(
			"method=%s path=%s status=%d bytes=%d duration=%s request_id=%s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			recorder.bytes,
			duration,
			requestID,
		)
	})
}
