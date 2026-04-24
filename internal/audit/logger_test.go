package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLog_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)

	if err := l.Log("rotate", "ns1", "secret/db", false, "ok", "test message"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var e Entry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if e.Operation != "rotate" {
		t.Errorf("expected operation=rotate, got %q", e.Operation)
	}
	if e.Namespace != "ns1" {
		t.Errorf("expected namespace=ns1, got %q", e.Namespace)
	}
	if e.Path != "secret/db" {
		t.Errorf("expected path=secret/db, got %q", e.Path)
	}
	if e.Status != "ok" {
		t.Errorf("expected status=ok, got %q", e.Status)
	}
	if e.Message != "test message" {
		t.Errorf("expected message='test message', got %q", e.Message)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLog_DryRun_Flag(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)

	_ = l.Log("sync", "ns2", "secret/api", true, "skipped", "")

	var e Entry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &e); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !e.DryRun {
		t.Error("expected dry_run=true")
	}
}

func TestLogSync_MessageContainsNamespaces(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)

	_ = l.LogSync("src-ns", "dst-ns", "secret/creds", false, "ok")

	var e Entry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &e); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !strings.Contains(e.Message, "src=src-ns") {
		t.Errorf("expected message to contain src namespace, got %q", e.Message)
	}
	if !strings.Contains(e.Message, "dst=dst-ns") {
		t.Errorf("expected message to contain dst namespace, got %q", e.Message)
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	l := New(nil)
	if l.out == nil {
		t.Error("expected non-nil writer when nil passed to New")
	}
}
