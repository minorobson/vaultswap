package policy

import (
	"context"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Applier writes policies to a Vault instance.
type Applier struct {
	client  *vaultapi.Client
	dryRun  bool
}

// NewApplier creates a new Applier.
func NewApplier(client *vaultapi.Client, dryRun bool) *Applier {
	return &Applier{client: client, dryRun: dryRun}
}

// ApplyResult holds the outcome of a single policy apply operation.
type ApplyResult struct {
	Name   string
	DryRun bool
	Skipped bool
}

// Apply writes the given policy to Vault. In dry-run mode it validates only.
func (a *Applier) Apply(ctx context.Context, p *Policy) (*ApplyResult, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("policy %q is invalid: %w", p.Name, err)
	}

	result := &ApplyResult{Name: p.Name, DryRun: a.dryRun}

	if a.dryRun {
		result.Skipped = true
		return result, nil
	}

	hcl := p.HCL()
	err := a.client.Sys().PutPolicyWithContext(ctx, p.Name, hcl)
	if err != nil {
		return nil, fmt.Errorf("failed to apply policy %q: %w", p.Name, err)
	}

	return result, nil
}
