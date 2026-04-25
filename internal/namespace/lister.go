package namespace

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Lister lists and filters Vault namespaces.
type Lister struct {
	client *vaultapi.Client
}

// New creates a new Lister using the provided Vault client.
func New(client *vaultapi.Client) *Lister {
	return &Lister{client: client}
}

// List returns all child namespaces under the given parent namespace path.
// Pass an empty string to list top-level namespaces.
func (l *Lister) List(ctx context.Context, parent string) ([]string, error) {
	path := "sys/namespaces"
	if parent != "" {
		parent = strings.Trim(parent, "/")
		path = fmt.Sprintf("%s/sys/namespaces", parent)
	}

	secret, err := l.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("listing namespaces at %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	raw, ok := keys.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for namespace keys: %T", keys)
	}

	result := make([]string, 0, len(raw))
	for _, v := range raw {
		name, ok := v.(string)
		if !ok {
			continue
		}
		result = append(result, strings.TrimSuffix(name, "/"))
	}
	return result, nil
}

// Filter returns only namespaces whose names contain the given substring.
func Filter(namespaces []string, substr string) []string {
	if substr == "" {
		return namespaces
	}
	out := make([]string, 0)
	for _, ns := range namespaces {
		if strings.Contains(ns, substr) {
			out = append(out, ns)
		}
	}
	return out
}
