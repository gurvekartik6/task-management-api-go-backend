package handler

import (
	"net/http"

	"task-management-api/internal/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(
	w http.ResponseWriter,
	r *http.Request,
) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
