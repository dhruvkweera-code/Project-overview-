package service

import (
	"context"
	"errors"
	"testing"

	"high-throughput-api/internal/repository/db"
)

type brokenPublisher struct{}

func (brokenPublisher) Publish(ctx context.Context, key, value []byte) error {
	return errors.New("kafka is down")
}

// createTestMessage saves a fresh message and returns it, so each test has its own row.
func createTestMessage(t *testing.T, svc *MessageService) db.Message {
	t.Helper()
	saved, err := svc.Create(context.Background(), "user-"+uniqueID(t), "msg-"+uniqueID(t), "old text")
	if err != nil {
		t.Fatalf("create test message: %v", err)
	}
	return saved
}

func TestUpdate_ChangesContentAndTime(t *testing.T) {
	svc, _ := newTestService(t)
	original := createTestMessage(t, svc)

	updated, err := svc.Update(context.Background(), original.UserID, original.MessageID, "new text")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Content != "new text" {
		t.Errorf("content = %q, want %q", updated.Content, "new text")
	}
	if updated.Status != "PENDING" {
		t.Errorf("status = %q, want PENDING", updated.Status)
	}
	if !updated.UpdatedAt.Time.After(original.UpdatedAt.Time) {
		t.Errorf("updated_at %v is not after %v", updated.UpdatedAt.Time, original.UpdatedAt.Time)
	}
	if !updated.CreatedAt.Time.Equal(original.CreatedAt.Time) {
		t.Errorf("created_at changed from %v to %v", original.CreatedAt.Time, updated.CreatedAt.Time)
	}
}

func TestUpdate_PublishesEvent(t *testing.T) {
	svc, pub := newTestService(t)
	original := createTestMessage(t, svc)
	callsBefore := pub.calls

	if _, err := svc.Update(context.Background(), original.UserID, original.MessageID, "new text"); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if pub.calls != callsBefore+1 {
		t.Errorf("publish calls = %d, want %d", pub.calls, callsBefore+1)
	}
}

func TestUpdate_MessageDoesNotExist(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.Update(context.Background(), "user-"+uniqueID(t), "missing", "new text")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdate_OtherUsersMessage(t *testing.T) {
	svc, _ := newTestService(t)
	original := createTestMessage(t, svc)

	_, err := svc.Update(context.Background(), "someone-else", original.MessageID, "hacked")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	stored, _ := svc.GetByID(context.Background(), original.UserID, original.MessageID)
	if stored.Content != "old text" {
		t.Errorf("content = %q, other user must not change it", stored.Content)
	}
}

func TestUpdate_WhileProcessing(t *testing.T) {
	svc, _ := newTestService(t)
	original := createTestMessage(t, svc)
	_, err := svc.q.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
		UserID: original.UserID, MessageID: original.MessageID, Status: "PROCESSING",
	})
	if err != nil {
		t.Fatalf("set PROCESSING: %v", err)
	}

	_, err = svc.Update(context.Background(), original.UserID, original.MessageID, "new text")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("err = %v, want ErrConflict", err)
	}
	stored, _ := svc.GetByID(context.Background(), original.UserID, original.MessageID)
	if stored.Content != "old text" {
		t.Errorf("content = %q, must not change while PROCESSING", stored.Content)
	}
}

func TestUpdate_KafkaDown(t *testing.T) {
	svc, _ := newTestService(t)
	original := createTestMessage(t, svc)
	brokenSvc := NewMessageService(svc.q, brokenPublisher{})

	_, err := brokenSvc.Update(context.Background(), original.UserID, original.MessageID, "new text")
	if !errors.Is(err, ErrQueueUnavailable) {
		t.Errorf("err = %v, want ErrQueueUnavailable", err)
	}
}
