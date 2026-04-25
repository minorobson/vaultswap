package rollback_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vaultswap/internal/rollback"
)

func TestPrintResults_LiveMode(t *testing.T) {
	results := []rollback.Result{
		{Path: "secret/app", DryRun: false, Restored: true},
		{Path: "secret/db", DryRun: false, Restored: true},
	}

	var buf bytes.Buffer
	rollback.PrintResults(results, &buf)
	out := buf.String()

	if !strings.Contains(out, "LIVE") {
		t.Errorf("expected LIVE label, got: %s", out)
	}
	if !strings.Contains(out, "restored") {
		t.Errorf("expected 'restored' status, got: %s", out)
	}
	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected path secret/app, got: %s", out)
	}
}

func TestPrintResults_DryRunLabel(t *testing.T) {
	results := []rollback.Result{
		{Path: "secret/app", DryRun: true, Restored: false},
	}

	var buf bytes.Buffer
	rollback.PrintResults(results, &buf)
	out := buf.String()

	if !strings.Contains(out, "DRY-RUN") {
		t.Errorf("expected DRY-RUN label, got: %s", out)
	}
	if !strings.Contains(out, "would restore") {
		t.Errorf("expected 'would restore' status, got: %s", out)
	}
}

func TestPrintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	rollback.PrintResults(nil, &buf)
	out := buf.String()

	if !strings.Contains(out, "No secrets to restore") {
		t.Errorf("expected empty message, got: %s", out)
	}
}
