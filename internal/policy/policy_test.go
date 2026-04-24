package policy

import (
	"strings"
	"testing"
)

func TestValidate_MissingName(t *testing.T) {
	p := &Policy{Name: "", Rules: []Rule{{Path: "secret/*", Capabilities: []string{"read"}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidate_MissingPath(t *testing.T) {
	p := &Policy{Name: "test", Rules: []Rule{{Path: "", Capabilities: []string{"read"}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestValidate_MissingCapabilities(t *testing.T) {
	p := &Policy{Name: "test", Rules: []Rule{{Path: "secret/*", Capabilities: []string{}}}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing capabilities")
	}
}

func TestValidate_Valid(t *testing.T) {
	p := &Policy{
		Name: "test",
		Rules: []Rule{
			{Path: "secret/*", Capabilities: []string{"read", "list"}},
		},
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHCL_Output(t *testing.T) {
	p := &Policy{
		Name: "test",
		Rules: []Rule{
			{Path: "secret/data/*", Capabilities: []string{"read", "create"}},
		},
	}
	hcl := p.HCL()
	if !strings.Contains(hcl, `path "secret/data/*"`) {
		t.Errorf("expected path in HCL output, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `"read"`) {
		t.Errorf("expected read capability in HCL output, got:\n%s", hcl)
	}
	if !strings.Contains(hcl, `"create"`) {
		t.Errorf("expected create capability in HCL output, got:\n%s", hcl)
	}
}
