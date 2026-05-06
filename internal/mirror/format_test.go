package mirror_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fhemberger/vaultswap/internal/mirror"
)

func TestFprintResults_MirroredLabel(t *testing.T) {
	var buf bytes.Buffer
	mirror.FprintResults(&buf, []mirror.Result{
		{SourcePath: "src/a", DestPath: "dst/a"},
	})
	if !strings.Contains(buf.String(), "[mirrored]") {
		t.Errorf("expected [mirrored] label, got: %s", buf.String())
	}
}

func TestFprintResults_DryRunLabel(t *testing.T) {
	var buf bytes.Buffer
	mirror.FprintResults(&buf, []mirror.Result{
		{SourcePath: "src/a", DestPath: "dst/a", DryRun: true},
	})
	if !strings.Contains(buf.String(), "[dry-run]") {
		t.Errorf("expected [dry-run] label, got: %s", buf.String())
	}
}

func TestFprintResults_SkippedLabel(t *testing.T) {
	var buf bytes.Buffer
	mirror.FprintResults(&buf, []mirror.Result{
		{SourcePath: "src/a", DestPath: "dst/a", Skipped: true},
	})
	if !strings.Contains(buf.String(), "[skipped]") {
		t.Errorf("expected [skipped] label, got: %s", buf.String())
	}
}

func TestFprintResults_ErrorLabel(t *testing.T) {
	var buf bytes.Buffer
	mirror.FprintResults(&buf, []mirror.Result{
		{SourcePath: "src/a", DestPath: "dst/a", Err: errors.New("boom")},
	})
	if !strings.Contains(buf.String(), "[error]") {
		t.Errorf("expected [error] label, got: %s", buf.String())
	}
}

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	mirror.FprintResults(&buf, []mirror.Result{})
	if !strings.Contains(buf.String(), "no paths") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
