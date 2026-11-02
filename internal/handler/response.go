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
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "message not found")
		return
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "message is being processed, try again later")
		return
	case errors.Is(err, service.ErrQueueUnavailable):
		log.Printf("queue unavailable: %v", err)
		writeError(w, http.StatusServiceUnavailable, "queue unavailable, try again later")
		return
	}
	log.Printf("unhandled service error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}
