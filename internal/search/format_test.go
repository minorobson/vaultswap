package search_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/seatgeek/vaultswap/internal/search"
)

func TestFprintResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	search.FprintResults(&buf, nil, false)
	if !strings.Contains(buf.String(), "no matches found") {
		t.Errorf("expected 'no matches found', got %q", buf.String())
	}
}

func TestFprintResults_MatchLabel(t *testing.T) {
	var buf bytes.Buffer
	results := []search.Result{
		{Path: "secret/app", Key: "db_pass", Value: "hunter2"},
	}
	search.FprintResults(&buf, results, false)
	out := buf.String()
	if !strings.Contains(out, "[match]") {
		t.Errorf("expected '[match]' label, got %q", out)
	}
	if !strings.Contains(out, "db_pass=hunter2") {
		t.Errorf("expected key=value in output, got %q", out)
	}
}

func TestFprintResults_MaskValues(t *testing.T) {
	var buf bytes.Buffer
	results := []search.Result{
		{Path: "secret/app", Key: "token", Value: "supersecret"},
	}
	search.FprintResults(&buf, results, true)
	out := buf.String()
	if strings.Contains(out, "supersecret") {
		t.Errorf("expected value to be masked, got %q", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("expected '***' mask, got %q", out)
	}
}

func TestFprintResults_ErrorLabel(t *testing.T) {
	var buf bytes.Buffer
	results := []search.Result{
		{Path: "secret/missing", Err: errors.New("not found")},
	}
	search.FprintResults(&buf, results, false)
	out := buf.String()
	if !strings.Contains(out, "[error]") {
		t.Errorf("expected '[error]' label, got %q", out)
	}
}
