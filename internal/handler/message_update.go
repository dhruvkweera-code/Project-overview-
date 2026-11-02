package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type updateMessageRequest struct {
	Content string `json:"content"`
}

type updateMessageResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
	UpdatedAt string `json:"updatedAt"`
}

// Update handles PUT /messages/{messageId}: it replaces the content of the caller's message.
func (h *MessageHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "X-User-ID header required")
		return
	}
	messageID := r.PathValue("messageId")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request updateMessageRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if request.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	if len(request.Content) > maxContentLength {
		writeError(w, http.StatusBadRequest, "content exceeds maximum length")
		return
	}

	updatedMessage, err := h.svc.Update(r.Context(), userID, messageID, request.Content)
	if err != nil {
		writeServiceErr(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, updateMessageResponse{
		Status:    "accepted",
		MessageID: updatedMessage.MessageID,
		UpdatedAt: updatedMessage.UpdatedAt.Time.Format(time.RFC3339),
	})
}
