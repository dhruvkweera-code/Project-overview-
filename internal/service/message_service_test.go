package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"high-throughput-api/internal/repository/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func uniqueID(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
}

type fakePublisher struct {
	calls int
}

func (f *fakePublisher) Publish(ctx context.Context, key, value []byte) error {
	f.calls++
	return nil
}

func newTestService(t *testing.T) (*MessageService, *fakePublisher) {
	t.Helper()
	pub := &fakePublisher{}
	return NewMessageService(db.New(testPool(t)), pub), pub
}

func TestCreate_Valid(t *testing.T) {
	svc, _ := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)

	msg, err := svc.Create(context.Background(), userID, msgID, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if msg.Status != "PENDING" {
		t.Errorf("Status = %q, want PENDING", msg.Status)
	}
	if msg.UserID != userID || msg.MessageID != msgID {
		t.Errorf("got UserID=%q MessageID=%q", msg.UserID, msg.MessageID)
	}
}

func TestCreate_Duplicate(t *testing.T) {
	svc, _ := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)

	first, err := svc.Create(context.Background(), userID, msgID, "hello")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	second, err := svc.Create(context.Background(), userID, msgID, "different content")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("duplicate submit created a new row: first.ID=%d second.ID=%d", first.ID, second.ID)
	}
	if second.Content != "hello" {
		t.Errorf("duplicate submit overwrote content: got %q, want original %q", second.Content, "hello")
	}
}

func TestCreate_PublishesOnNewMessage(t *testing.T) {
	svc, pub := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)

	if _, err := svc.Create(context.Background(), userID, msgID, "hello"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if pub.calls != 1 {
		t.Errorf("Publish calls = %d, want 1", pub.calls)
	}
}

func TestCreate_DoesNotPublishOnDuplicate(t *testing.T) {
	svc, pub := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)

	if _, err := svc.Create(context.Background(), userID, msgID, "hello"); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if _, err := svc.Create(context.Background(), userID, msgID, "hello again"); err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if pub.calls != 1 {
		t.Errorf("Publish calls = %d, want 1 (no publish on duplicate)", pub.calls)
	}
}

func TestGetByID_Owner(t *testing.T) {
	svc, _ := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)
	created, err := svc.Create(context.Background(), userID, msgID, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.GetByID(context.Background(), userID, msgID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("got.ID = %d, want %d", got.ID, created.ID)
	}
}

func TestGetByID_WrongUser_ReturnsNotFound(t *testing.T) {
	svc, _ := newTestService(t)
	userID, msgID := uniqueID(t), uniqueID(t)
	if _, err := svc.Create(context.Background(), userID, msgID, "hello"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := svc.GetByID(context.Background(), uniqueID(t), msgID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newTestService(t)

	_, err := svc.GetByID(context.Background(), uniqueID(t), uniqueID(t))
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
