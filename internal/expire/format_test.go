package expire

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFprintResults_NoTTL(t *testing.T) {
	var buf bytes.Buffer
	results := []Result{
		{Path: "secret/foo", NoTTL: true},
	}
	FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "[NO TTL]") {
		t.Errorf("expected [NO TTL] label, got: %s", buf.String())
	}
}

func TestFprintResults_Expired(t *testing.T) {
	var buf bytes.Buffer
	expiredAt := time.Now().Add(-1 * time.Hour)
	results := []Result{
		{
			Path:      "secret/bar",
			TTL:       24 * time.Hour,
			ExpiresAt: expiredAt,
			Expired:   true,
		},
	}
	FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "[EXPIRED]") {
		t.Errorf("expected [EXPIRED] label, got: %s", buf.String())
	}
}

func TestFprintResults_OK(t *testing.T) {
	var buf bytes.Buffer
	results := []Result{
		{
			Path:      "secret/baz",
			TTL:       48 * time.Hour,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Expired:   false,
		},
	}
	FprintResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "[OK]") {
		t.Errorf("expected [OK] label, got: %s", out)
	}
	if !strings.Contains(out, "secret/baz") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestFprintResults_Error(t *testing.T) {
	var buf bytes.Buffer
	results := []Result{
		{Path: "secret/missing", Err: errors.New("not found")},
	}
	FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Errorf("expected [ERROR] label, got: %s", buf.String())
	}
}

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{})
	if !strings.Contains(buf.String(), "no paths checked") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
