package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/api/types"
	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	title := "Internal Server Error"

	if errors.Is(err, core.ErrNotFound) {
		status = http.StatusNotFound
		title = "Resource Not Found"
	} else if errors.Is(err, core.ErrValidationFailed) || errors.Is(err, core.ErrCycleDetected) {
		status = http.StatusBadRequest
		title = "Bad Request"
	} else if errors.Is(err, core.ErrAlreadyExists) {
		status = http.StatusConflict
		title = "Conflict"
	} else if errors.Is(err, core.ErrInvalidStateTransition) {
		status = http.StatusConflict
		title = "Invalid State Transition"
	} else if errors.Is(err, core.ErrUnauthorized) {
		status = http.StatusUnauthorized
		title = "Unauthorized"
	} else if errors.Is(err, core.ErrForbidden) {
		status = http.StatusForbidden
		title = "Forbidden"
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	problem := types.ProblemDetails{
		Type:      "about:blank",
		Title:     title,
		Status:    status,
		Detail:    err.Error(),
		Instance:  r.URL.Path,
		Timestamp: time.Now().UTC(),
	}
	_ = json.NewEncoder(w).Encode(problem)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
