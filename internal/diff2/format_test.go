package diff2_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vaultswap/vaultswap/internal/diff2"
)

func TestFprintResults_NoDiff(t *testing.T) {
	var buf bytes.Buffer
	results := []diff2.Result{{Path: "secret/a"}}
	diff2.FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "[no diff]") {
		t.Errorf("expected [no diff] label, got: %s", buf.String())
	}
}

func TestFprintResults_WithAdded(t *testing.T) {
	var buf bytes.Buffer
	results := []diff2.Result{
		{
			Path:  "secret/b",
			Added: map[string]string{"foo": "bar"},
		},
	}
	diff2.FprintResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "[diff]") {
		t.Errorf("expected [diff] label")
	}
	if !strings.Contains(out, "foo") {
		t.Errorf("expected key foo in output")
	}
}

func TestFprintResults_WithError(t *testing.T) {
	var buf bytes.Buffer
	results := []diff2.Result{
		{Path: "secret/c", Error: fmt.Errorf("read failed")},
	}
	diff2.FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "[error]") {
		t.Errorf("expected [error] label")
	}
}

func TestFprintResults_ChangedAndRemoved(t *testing.T) {
	var buf bytes.Buffer
	results := []diff2.Result{
		{
			Path:    "secret/d",
			Removed: map[string]string{"gone": "val"},
			Changed: map[string][2]string{"k": {"old", "new"}},
		},
	}
	diff2.FprintResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "gone") {
		t.Errorf("expected removed key in output")
	}
	if !strings.Contains(out, "old") || !strings.Contains(out, "new") {
		t.Errorf("expected changed values in output")
	}
}
