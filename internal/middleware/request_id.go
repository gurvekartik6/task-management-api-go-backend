package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey struct{}

var requestIDKey contextKey

// generateRequestID returns a cryptographically random 128-bit request ID.
func generateRequestID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return hex.EncodeToString(value), nil
}

// RequestID assigns a new ID to each incoming request, stores it in the
// request context, and exposes it to the client through X-Request-ID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, err := generateRequestID()
		if err != nil {
			http.Error(w, "could not generate request ID", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r)
	})
}

// GetRequestID retrieves the request ID from a request context.
func GetRequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey).(string)
	return requestID
}
