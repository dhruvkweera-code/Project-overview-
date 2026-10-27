package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"high-throughput-api/internal/repository/db"
	"high-throughput-api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

type fakeMessageService struct {
	createFn func(ctx context.Context, userID, messageID, content string) (db.Message, error)
	getFn    func(ctx context.Context, callerUserID, messageID string) (db.Message, error)
}

func (f *fakeMessageService) Create(ctx context.Context, userID, messageID, content string) (db.Message, error) {
	return f.createFn(ctx, userID, messageID, content)
}

func (f *fakeMessageService) GetByID(ctx context.Context, callerUserID, messageID string) (db.Message, error) {
	return f.getFn(ctx, callerUserID, messageID)
}

func TestCreate_Valid(t *testing.T) {
	fake := &fakeMessageService{
		createFn: func(ctx context.Context, userID, messageID, content string) (db.Message, error) {
			return db.Message{UserID: userID, MessageID: messageID, Content: content, Status: "PENDING"}, nil
		},
	}
	h := NewMessageHandler(fake)

	body := strings.NewReader(`{"userId":"u1","messageId":"m1","content":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/messages", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	want := `{"status":"accepted","messageId":"m1"}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	h := NewMessageHandler(&fakeMessageService{})
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(`not json`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreate_MissingFields(t *testing.T) {
	h := NewMessageHandler(&fakeMessageService{})
	req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(`{"userId":"u1"}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreate_ContentTooLong(t *testing.T) {
	h := NewMessageHandler(&fakeMessageService{})
	longContent := strings.Repeat("a", maxContentLength+1)
	req := httptest.NewRequest(http.MethodPost, "/messages",
		strings.NewReader(`{"userId":"u1","messageId":"m1","content":"`+longContent+`"}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByID_Valid(t *testing.T) {
	fake := &fakeMessageService{
		getFn: func(ctx context.Context, callerUserID, messageID string) (db.Message, error) {
			return db.Message{
				UserID:    callerUserID,
				MessageID: messageID,
				Status:    "PENDING",
				CreatedAt: pgtype.Timestamptz{Valid: true},
			}, nil
		},
	}
	h := NewMessageHandler(fake)

	req := httptest.NewRequest(http.MethodGet, "/messages/m1", nil)
	req.Header.Set("X-User-ID", "u1")
	req.SetPathValue("messageId", "m1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetByID_MissingUserHeader(t *testing.T) {
	h := NewMessageHandler(&fakeMessageService{})
	req := httptest.NewRequest(http.MethodGet, "/messages/m1", nil)
	req.SetPathValue("messageId", "m1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	fake := &fakeMessageService{
		getFn: func(ctx context.Context, callerUserID, messageID string) (db.Message, error) {
			return db.Message{}, service.ErrNotFound
		},
	}
	h := NewMessageHandler(fake)

	req := httptest.NewRequest(http.MethodGet, "/messages/m1", nil)
	req.Header.Set("X-User-ID", "u1")
	req.SetPathValue("messageId", "m1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
