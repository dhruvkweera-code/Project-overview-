package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"high-throughput-api/internal/repository/db"
	"high-throughput-api/internal/service"

	"github.com/jackc/pgx/v5/pgtype"
)

// sendUpdate builds a PUT /messages/msg-1 request, runs the handler and returns the recorded response.
func sendUpdate(t *testing.T, fake *fakeMessageService, userID, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/messages/msg-1", strings.NewReader(body))
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	req.SetPathValue("messageId", "msg-1")
	rec := httptest.NewRecorder()
	NewMessageHandler(fake).Update(rec, req)
	return rec
}

// serviceReturning makes a fake service whose Update always returns the given error.
func serviceReturning(err error) *fakeMessageService {
	return &fakeMessageService{
		updateFn: func(ctx context.Context, userID, messageID, newContent string) (db.Message, error) {
			return db.Message{}, err
		},
	}
}

func TestUpdate_Valid(t *testing.T) {
	updateTime := time.Date(2026, 10, 29, 15, 50, 12, 0, time.UTC)
	fake := &fakeMessageService{
		updateFn: func(ctx context.Context, userID, messageID, newContent string) (db.Message, error) {
			if userID != "user-A" || messageID != "msg-1" || newContent != "new text" {
				t.Errorf("service got (%q, %q, %q)", userID, messageID, newContent)
			}
			return db.Message{
				UserID:    userID,
				MessageID: messageID,
				UpdatedAt: pgtype.Timestamptz{Time: updateTime, Valid: true},
			}, nil
		},
	}

	rec := sendUpdate(t, fake, "user-A", `{"content":"new text"}`)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
	want := `{"status":"accepted","messageId":"msg-1","updatedAt":"2026-10-29T15:50:12Z"}`
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
}

func TestUpdate_MissingUserHeader(t *testing.T) {
	rec := sendUpdate(t, serviceReturning(nil), "", `{"content":"new text"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdate_BadBody(t *testing.T) {
	badBodies := map[string]string{
		"not json":      `hello`,
		"empty content": `{"content":""}`,
		"unknown field": `{"content":"x","userId":"user-B"}`,
		"too long":      `{"content":"` + strings.Repeat("a", maxContentLength+1) + `"}`,
	}
	for name, body := range badBodies {
		t.Run(name, func(t *testing.T) {
			rec := sendUpdate(t, serviceReturning(nil), "user-A", body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestUpdate_ServiceErrors(t *testing.T) {
	cases := map[string]struct {
		err        error
		wantStatus int
	}{
		"not found":  {service.ErrNotFound, http.StatusNotFound},
		"processing": {service.ErrConflict, http.StatusConflict},
		"kafka down": {service.ErrQueueUnavailable, http.StatusServiceUnavailable},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := sendUpdate(t, serviceReturning(tc.err), "user-A", `{"content":"new text"}`)
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestGetByID_ShowsUpdatedAt(t *testing.T) {
	fake := &fakeMessageService{
		getFn: func(ctx context.Context, callerUserID, messageID string) (db.Message, error) {
			return db.Message{
				UserID:    callerUserID,
				MessageID: messageID,
				UpdatedAt: pgtype.Timestamptz{Time: time.Date(2026, 10, 29, 15, 50, 12, 0, time.UTC), Valid: true},
			}, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/messages/msg-1", nil)
	req.Header.Set("X-User-ID", "user-A")
	req.SetPathValue("messageId", "msg-1")
	rec := httptest.NewRecorder()

	NewMessageHandler(fake).GetByID(rec, req)

	if !strings.Contains(rec.Body.String(), `"updatedAt":"2026-10-29T15:50:12Z"`) {
		t.Errorf("body missing updatedAt: %s", rec.Body.String())
	}
}
