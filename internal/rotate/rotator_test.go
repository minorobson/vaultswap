package rotate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/vaultswap/internal/rotate"
)

type mockClient struct {
	data     map[string]interface{}
	written  map[string]interface{}
	readErr  error
	writeErr error
}

func (m *mockClient) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	copy := make(map[string]interface{}, len(m.data))
	for k, v := range m.data {
		copy[k] = v
	}
	return copy, nil
}

func (m *mockClient) WriteSecret(_ context.Context, _ string, data map[string]interface{}) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	m.written = data
	return nil
}

func TestRotate_KeysUpdated(t *testing.T) {
	client := &mockClient{data: map[string]interface{}{"password": "old", "user": "alice"}}
	r := rotate.New(client, rotate.Options{KeysToRotate: []string{"password"}})

	res, err := r.Rotate(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DryRun {
		t.Error("expected DryRun=false")
	}
	if client.written["user"] != "alice" {
		t.Error("non-rotated key should be preserved")
	}
	if client.written["password"] == "old" {
		t.Error("rotated key should have a new value")
	}
}

func TestRotate_DryRun_DoesNotWrite(t *testing.T) {
	client := &mockClient{data: map[string]interface{}{"password": "old"}}
	r := rotate.New(client, rotate.Options{KeysToRotate: []string{"password"}, DryRun: true})

	res, err := r.Rotate(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.DryRun {
		t.Error("expected DryRun=true")
	}
	if client.written != nil {
		t.Error("dry-run should not write")
	}
}

func TestRotate_ReadError(t *testing.T) {
	client := &mockClient{readErr: errors.New("vault unavailable")}
	r := rotate.New(client, rotate.Options{KeysToRotate: []string{"password"}})

	_, err := r.Rotate(context.Background(), "secret/myapp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRotate_WriteError(t *testing.T) {
	client := &mockClient{
		data:     map[string]interface{}{"password": "old"},
		writeErr: errors.New("permission denied"),
	}
	r := rotate.New(client, rotate.Options{KeysToRotate: []string{"password"}})

	_, err := r.Rotate(context.Background(), "secret/myapp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
