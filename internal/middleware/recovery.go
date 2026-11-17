package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"task-management-api/internal/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		recorder := newResponseRecorder(w)

		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf(
					"panic recovered: request_id=%s error=%v\n%s",
					GetRequestID(r.Context()),
					recovered,
					debug.Stack(),
				)

				if recorder.bytes == 0 &&
					recorder.statusCode == http.StatusOK {
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
