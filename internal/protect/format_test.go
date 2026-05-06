package protect_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/vaultswap/internal/protect"
)

func TestFprintResults_ProtectedLabel(t *testing.T) {
	var buf bytes.Buffer
	protect.FprintResults(&buf, []protect.Result{
		{Path: "secret/foo", DryRun: false, Skipped: false},
	})
	if !strings.Contains(buf.String(), "[protected]") {
		t.Errorf("expected [protected] label, got: %s", buf.String())
	}
}

func TestFprintResults_DryRunLabel(t *testing.T) {
	var buf bytes.Buffer
	protect.FprintResults(&buf, []protect.Result{
		{Path: "secret/foo", DryRun: true},
	})
	if !strings.Contains(buf.String(), "[dry-run]") {
		t.Errorf("expected [dry-run] label, got: %s", buf.String())
	}
}

func TestFprintResults_SkippedLabel(t *testing.T) {
	var buf bytes.Buffer
	protect.FprintResults(&buf, []protect.Result{
		{Path: "secret/foo", Skipped: true},
	})
	if !strings.Contains(buf.String(), "[skipped]") {
		t.Errorf("expected [skipped] label, got: %s", buf.String())
	}
}

func TestFprintResults_ErrorLabel(t *testing.T) {
	var buf bytes.Buffer
	protect.FprintResults(&buf, []protect.Result{
		{Path: "secret/foo", Error: fmt.Errorf("boom")},
	})
	if !strings.Contains(buf.String(), "[error]") {
		t.Errorf("expected [error] label, got: %s", buf.String())
	}
}

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	protect.FprintResults(&buf, []protect.Result{})
	if !strings.Contains(buf.String(), "no paths") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
