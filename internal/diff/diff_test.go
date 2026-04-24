package diff

import (
	"strings"
	"testing"
)

func TestCompare_Added(t *testing.T) {
	src := map[string]interface{}{"foo": "bar", "new": "value"}
	dst := map[string]interface{}{"foo": "bar"}

	r := Compare(src, dst)
	if !r.HasChanges() {
		t.Fatal("expected changes")
	}
	found := false
	for _, c := range r.Changes {
		if c.Key == "new" && c.Type == Added {
			found = true
		}
	}
	if !found {
		t.Error("expected 'new' key to be Added")
	}
}

func TestCompare_Removed(t *testing.T) {
	src := map[string]interface{}{"foo": "bar"}
	dst := map[string]interface{}{"foo": "bar", "old": "gone"}

	r := Compare(src, dst)
	for _, c := range r.Changes {
		if c.Key == "old" && c.Type != Removed {
			t.Errorf("expected 'old' to be Removed, got %s", c.Type)
		}
	}
}

func TestCompare_Modified(t *testing.T) {
	src := map[string]interface{}{"key": "new-val"}
	dst := map[string]interface{}{"key": "old-val"}

	r := Compare(src, dst)
	if len(r.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(r.Changes))
	}
	c := r.Changes[0]
	if c.Type != Modified {
		t.Errorf("expected Modified, got %s", c.Type)
	}
	if c.OldValue != "old-val" || c.NewValue != "new-val" {
		t.Errorf("unexpected values: old=%s new=%s", c.OldValue, c.NewValue)
	}
}

func TestCompare_Unchanged(t *testing.T) {
	src := map[string]interface{}{"a": "1"}
	dst := map[string]interface{}{"a": "1"}

	r := Compare(src, dst)
	if r.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestFormat_MaskValues(t *testing.T) {
	src := map[string]interface{}{"secret": "s3cr3t"}
	dst := map[string]interface{}{}

	r := Compare(src, dst)
	out := Format(r, true)
	if strings.Contains(out, "s3cr3t") {
		t.Error("masked output should not contain actual secret value")
	}
	if !strings.Contains(out, "***") {
		t.Error("masked output should contain '***'")
	}
}

func TestFormat_ShowValues(t *testing.T) {
	src := map[string]interface{}{"key": "visible"}
	dst := map[string]interface{}{"key": "old"}

	r := Compare(src, dst)
	out := Format(r, false)
	if !strings.Contains(out, "visible") {
		t.Error("unmasked output should contain actual value")
	}
}
