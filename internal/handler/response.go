package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"high-throughput-api/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeServiceErr(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	log.Printf("unhandled service error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
