package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"high-throughput-api/internal/repository/db"
)

const maxContentLength = 10_000

type messageService interface {
	Create(ctx context.Context, userID, messageID, content string) (db.Message, error)
	GetByID(ctx context.Context, callerUserID, messageID string) (db.Message, error)
}

type MessageHandler struct {
	svc messageService
}

func NewMessageHandler(svc messageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

type createMessageRequest struct {
	UserID    string `json:"userId"`
	MessageID string `json:"messageId"`
	Content   string `json:"content"`
}

func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createMessageRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.UserID == "" || req.MessageID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "userId, messageId and content are required")
		return
	}
	if len(req.Content) > maxContentLength {
		writeError(w, http.StatusBadRequest, "content exceeds maximum length")
		return
	}

	msg, err := h.svc.Create(r.Context(), req.UserID, req.MessageID, req.Content)
	if err != nil {
		writeServiceErr(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, createMessageResponse{Status: "accepted", MessageID: msg.MessageID})
}

type createMessageResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"messageId"`
}

func (h *MessageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	callerID := r.Header.Get("X-User-ID")
	if callerID == "" {
		writeError(w, http.StatusUnauthorized, "X-User-ID header required")
		return
	}

	messageID := r.PathValue("messageId")
	msg, err := h.svc.GetByID(r.Context(), callerID, messageID)
	if err != nil {
		writeServiceErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toStatusResponse(msg))
}

type statusResponse struct {
	MessageID   string  `json:"messageId"`
	UserID      string  `json:"userId"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	ProcessedAt *string `json:"processedAt"`
}

func toStatusResponse(msg db.Message) statusResponse {
	resp := statusResponse{
		MessageID: msg.MessageID,
		UserID:    msg.UserID,
		Status:    msg.Status,
	}
	if msg.CreatedAt.Valid {
		resp.CreatedAt = msg.CreatedAt.Time.Format(time.RFC3339)
	}
	if msg.ProcessedAt.Valid {
		processedAt := msg.ProcessedAt.Time.Format(time.RFC3339)
		resp.ProcessedAt = &processedAt
	}
	return resp
}
