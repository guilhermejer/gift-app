package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gift-app/api/internal/logger"
)

const (
	logLevelError = "error"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		logger.Log(logLevelError, "writeJSON failed", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	if err == nil {
		err = errors.New(msg)
	}
	logger.Log(logLevelError, msg, err)
	writeJSON(w, status, map[string]string{"error": msg})
}
