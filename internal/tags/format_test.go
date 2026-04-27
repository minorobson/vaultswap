package tags

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{})
	if !strings.Contains(buf.String(), "no paths processed") {
		t.Errorf("expected 'no paths processed', got: %s", buf.String())
	}
}

func TestFprintResults_TaggedLabel(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{
		{
			Path:   "secret/app/db",
			Tags:   map[string]string{"env": "prod"},
			DryRun: false,
		},
	})
	out := buf.String()
	if !strings.Contains(out, "[tagged]") {
		t.Errorf("expected '[tagged]' label, got: %s", out)
	}
	if !strings.Contains(out, "secret/app/db") {
		t.Errorf("expected path in output, got: %s", out)
	}
	if !strings.Contains(out, "env=prod") {
		t.Errorf("expected tag pair in output, got: %s", out)
	}
}

func TestFprintResults_DryRunLabel(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{
		{
			Path:   "secret/app/api",
			Tags:   map[string]string{"team": "platform"},
			DryRun: true,
		},
	})
	out := buf.String()
	if !strings.Contains(out, "[dry-run]") {
		t.Errorf("expected '[dry-run]' label, got: %s", out)
	}
}

func TestFprintResults_ErrorLabel(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{
		{
			Path: "secret/app/broken",
			Err:  errors.New("permission denied"),
		},
	})
	out := buf.String()
	if !strings.Contains(out, "[error]") {
		t.Errorf("expected '[error]' label, got: %s", out)
	}
	if !strings.Contains(out, "permission denied") {
		t.Errorf("expected error message in output, got: %s", out)
	}
}

func TestFprintResults_SkippedLabel(t *testing.T) {
	var buf bytes.Buffer
	FprintResults(&buf, []Result{
		{
			Path:    "secret/app/unchanged",
			Skipped: true,
		},
	})
	out := buf.String()
	if !strings.Contains(out, "[skipped]") {
		t.Errorf("expected '[skipped]' label, got: %s", out)
	}
}
