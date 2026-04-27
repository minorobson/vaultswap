package lock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultswap/internal/lock"
)

// mockClient is a minimal in-memory vault client for testing.
type mockClient struct {
	store  map[string]map[string]interface{}
	writeErr error
	deleteErr error
}

func newMockClient() *mockClient {
	return &mockClient{store: make(map[string]map[string]interface{})}
}

func (m *mockClient) ReadSecret(_ context.Context, path string) (map[string]interface{}, error) {
	v, ok := m.store[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (m *mockClient) WriteSecret(_ context.Context, path string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.store[path] = data
	return nil
}

func (m *mockClient) DeleteSecret(_ context.Context, path string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.store, path)
	return nil
}

func TestLock_Success(t *testing.T) {
	client := newMockClient()
	l := lock.New(client, "locks/deploy", false)
	res := l.Lock(context.Background(), "ci-bot")
	if res.Action != "locked" {
		t.Fatalf("expected locked, got %s", res.Action)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
}

func TestLock_AlreadyLocked_Skipped(t *testing.T) {
	client := newMockClient()
	client.store["locks/deploy"] = map[string]interface{}{"owner": "someone"}
	l := lock.New(client, "locks/deploy", false)
	res := l.Lock(context.Background(), "ci-bot")
	if res.Action != "skipped" {
		t.Fatalf("expected skipped, got %s", res.Action)
	}
}

func TestLock_DryRun_DoesNotWrite(t *testing.T) {
	client := newMockClient()
	l := lock.New(client, "locks/deploy", true)
	res := l.Lock(context.Background(), "ci-bot")
	if res.Action != "locked" || !res.DryRun {
		t.Fatalf("expected dry-run locked, got action=%s dryRun=%v", res.Action, res.DryRun)
	}
	if _, exists := client.store["locks/deploy"]; exists {
		t.Fatal("dry-run should not write to vault")
	}
}

func TestUnlock_Success(t *testing.T) {
	client := newMockClient()
	client.store["locks/deploy"] = map[string]interface{}{"owner": "ci-bot"}
	l := lock.New(client, "locks/deploy", false)
	res := l.Unlock(context.Background())
	if res.Action != "unlocked" {
		t.Fatalf("expected unlocked, got %s", res.Action)
	}
	if _, exists := client.store["locks/deploy"]; exists {
		t.Fatal("lock entry should have been deleted")
	}
}

func TestUnlock_NoLock_Skipped(t *testing.T) {
	client := newMockClient()
	l := lock.New(client, "locks/deploy", false)
	res := l.Unlock(context.Background())
	if res.Action != "skipped" {
		t.Fatalf("expected skipped, got %s", res.Action)
	}
}

func TestLock_WriteError_ReturnsError(t *testing.T) {
	client := newMockClient()
	client.writeErr = errors.New("permission denied")
	l := lock.New(client, "locks/deploy", false)
	res := l.Lock(context.Background(), "ci-bot")
	if res.Action != "error" {
		t.Fatalf("expected error action, got %s", res.Action)
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error")
	}
}
