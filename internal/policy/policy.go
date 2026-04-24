package policy

import (
	"fmt"
	"strings"
)

// Rule represents a single path-based access rule.
type Rule struct {
	Path         string
	Capabilities []string
}

// Policy represents a Vault policy with a name and a set of rules.
type Policy struct {
	Name  string
	Rules []Rule
}

// Validate checks that the policy has a name and at least one valid rule.
func (p *Policy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("policy name must not be empty")
	}
	for i, r := range p.Rules {
		if strings.TrimSpace(r.Path) == "" {
			return fmt.Errorf("rule[%d]: path must not be empty", i)
		}
		if len(r.Capabilities) == 0 {
			return fmt.Errorf("rule[%d]: capabilities must not be empty", i)
		}
	}
	return nil
}

// HCL renders the policy as an HCL string suitable for writing to Vault.
func (p *Policy) HCL() string {
	var sb strings.Builder
	for _, r := range p.Rules {
		caps := make([]string, len(r.Capabilities))
		for i, c := range r.Capabilities {
			caps[i] = fmt.Sprintf("%q", c)
		}
		sb.WriteString(fmt.Sprintf("path %q {\n", r.Path))
		sb.WriteString(fmt.Sprintf("  capabilities = [%s]\n", strings.Join(caps, ", ")))
		sb.WriteString("}\n")
	}
	return sb.String()
}
