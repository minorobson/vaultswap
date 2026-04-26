package promote_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/vaultswap/internal/promote"
)

func TestFprintResults_DryRunLabel(t *testing.T) {
	results := []promote.Result{
		{Path: "app/secret", Src: "staging", Dst: "prod", DryRun: true, Skipped: true},
	}
	var buf bytes.Buffer
	promote.FprintResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected 'dry-run' in output, got: %s", out)
	}
	if !strings.Contains(out, "app/secret") {
		t.Errorf("expected path in output, got: %s", out)
	}
}

func TestFprintResults_ErrorLabel(t *testing.T) {
	results := []promote.Result{
		{Path: "app/secret", Src: "staging", Dst: "prod", Err: fmt.Errorf("permission denied")},
	}
	var buf bytes.Buffer
	promote.FprintResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in output, got: %s", out)
	}
}

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	promote.FprintResults(&buf, nil)
	if !strings.Contains(buf.String(), "no paths") {
		t.Errorf("expected 'no paths' message, got: %s", buf.String())
	}
}

func TestFprintResults_PromotedLabel(t *testing.T) {
	results := []promote.Result{
		{Path: "app/db", Src: "dev", Dst: "staging"},
	}
	var buf bytes.Buffer
	promote.FprintResults(&buf, results)
	if !strings.Contains(buf.String(), "promoted") {
		t.Errorf("expected 'promoted' label, got: %s", buf.String())
	}
}
