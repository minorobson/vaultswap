package diff2_test

import (
	"testing"

	"github.com/vaultswap/vaultswap/internal/diff2"
)

func TestCompare_Added(t *testing.T) {
	c := diff2.New(false)
	src := map[string]any{}
	dst := map[string]any{"newkey": "newval"}
	res := c.Compare("secret/path", src, dst)
	if !res.HasChanges() {
		t.Fatal("expected changes")
	}
	if _, ok := res.Added["newkey"]; !ok {
		t.Errorf("expected newkey in Added")
	}
}

func TestCompare_Removed(t *testing.T) {
	c := diff2.New(false)
	src := map[string]any{"oldkey": "oldval"}
	dst := map[string]any{}
	res := c.Compare("secret/path", src, dst)
	if _, ok := res.Removed["oldkey"]; !ok {
		t.Errorf("expected oldkey in Removed")
	}
}

func TestCompare_Changed(t *testing.T) {
	c := diff2.New(false)
	src := map[string]any{"key": "old"}
	dst := map[string]any{"key": "new"}
	res := c.Compare("secret/path", src, dst)
	pair, ok := res.Changed["key"]
	if !ok {
		t.Fatal("expected key in Changed")
	}
	if pair[0] != "old" || pair[1] != "new" {
		t.Errorf("unexpected pair: %v", pair)
	}
}

func TestCompare_Unchanged(t *testing.T) {
	c := diff2.New(false)
	src := map[string]any{"key": "val"}
	dst := map[string]any{"key": "val"}
	res := c.Compare("secret/path", src, dst)
	if res.HasChanges() {
		t.Error("expected no changes")
	}
}

func TestCompare_MaskValues(t *testing.T) {
	c := diff2.New(true)
	src := map[string]any{"key": "secret"}
	dst := map[string]any{"key": "other"}
	res := c.Compare("secret/path", src, dst)
	pair, ok := res.Changed["key"]
	if !ok {
		t.Fatal("expected key in Changed")
	}
	for _, v := range pair {
		for _, ch := range v {
			if ch != '*' {
				t.Errorf("expected masked value, got %q", v)
				break
			}
		}
	}
}

func TestHasChanges_False(t *testing.T) {
	r := diff2.Result{Path: "p"}
	if r.HasChanges() {
		t.Error("expected no changes on empty result")
	}
}
