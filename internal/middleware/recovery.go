package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"task-management-api/internal/response"
)

// Recovery catches panics raised by downstream handlers in the same goroutine.
// If no response has been committed, it sends a generic HTTP 500 response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newResponseRecorder(w)

		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"panic recovered: request_id=%s error=%v\n%s",
					GetRequestID(r.Context()),
					recovered,
					debug.Stack(),
				)

				if !recorder.wroteHeader {
					response.Error(
						recorder,
						http.StatusInternalServerError,
						"internal server error",
					)
				}
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}
