package rotate_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/vaultswap/internal/rotate"
)

func TestPrintResult_ContainsExpectedFields(t *testing.T) {
	r := &rotate.Result{
		Path:      "secret/myapp",
		RotatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Keys:      []string{"password", "api_key"},
		DryRun:    false,
	}

	var buf bytes.Buffer
	rotate.PrintResult(&buf, r)
	out := buf.String()

	checks := []string{
		"secret/myapp",
		"applied",
		"2024-01-15T10:00:00Z",
		"password",
		"api_key",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestPrintResult_DryRunLabel(t *testing.T) {
	r := &rotate.Result{
		Path:      "secret/myapp",
		RotatedAt: time.Now(),
		Keys:      []string{"token"},
		DryRun:    true,
	}

	var buf bytes.Buffer
	rotate.PrintResult(&buf, r)
	out := buf.String()

	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected 'dry-run' in output, got:\n%s", out)
	}
}
