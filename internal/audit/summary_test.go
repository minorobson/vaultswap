package audit

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func baseTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-06-01T10:00:00Z")
	return t
}

func TestFprintSummary_ContainsOperation(t *testing.T) {
	var buf bytes.Buffer
	s := OperationSummary{
		Operation: "rotate",
		Total:     5,
		Succeeded: 5,
		StartedAt: baseTime(),
		Duration:  120 * time.Millisecond,
	}
	FprintSummary(&buf, s)
	if !strings.Contains(buf.String(), "rotate") {
		t.Errorf("expected output to contain operation name, got: %s", buf.String())
	}
}

func TestFprintSummary_DryRunLabel(t *testing.T) {
	var buf bytes.Buffer
	s := OperationSummary{
		Operation: "sync",
		DryRun:    true,
		StartedAt: baseTime(),
	}
	FprintSummary(&buf, s)
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run label in output, got: %s", buf.String())
	}
}

func TestFprintSummary_NamespaceOmittedWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	s := OperationSummary{
		Operation: "clone",
		Namespace: "",
		StartedAt: baseTime(),
	}
	FprintSummary(&buf, s)
	if strings.Contains(buf.String(), "Namespace:") {
		t.Errorf("expected Namespace line to be omitted when empty, got: %s", buf.String())
	}
}

func TestFprintSummary_NamespacePresentWhenSet(t *testing.T) {
	var buf bytes.Buffer
	s := OperationSummary{
		Operation: "promote",
		Namespace: "team-alpha",
		StartedAt: baseTime(),
	}
	FprintSummary(&buf, s)
	if !strings.Contains(buf.String(), "team-alpha") {
		t.Errorf("expected namespace in output, got: %s", buf.String())
	}
}

func TestFprintSummary_CountsPresent(t *testing.T) {
	var buf bytes.Buffer
	s := OperationSummary{
		Operation: "import",
		Total:     10,
		Succeeded: 7,
		Failed:    2,
		Skipped:   1,
		StartedAt: baseTime(),
		Duration:  50 * time.Millisecond,
	}
	FprintSummary(&buf, s)
	out := buf.String()
	for _, want := range []string{"10", "7", "2", "1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %s", want, out)
		}
	}
}
