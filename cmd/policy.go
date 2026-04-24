package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	vaultapi "github.com/hashicorp/vault/api"

	"github.com/yourorg/vaultswap/internal/policy"
)

var (
	policyName         string
	policyPath         string
	policyCapabilities []string
	policyDryRun       bool
)

func init() {
	policyCmd := &cobra.Command{
		Use:   "policy",
		Short: "Apply a policy to a Vault namespace",
		RunE:  runPolicy,
	}
	policyCmd.Flags().StringVar(&policyName, "name", "", "Policy name (required)")
	policyCmd.Flags().StringVar(&policyPath, "path", "", "Secret path for the policy rule (required)")
	policyCmd.Flags().StringSliceVar(&policyCapabilities, "capabilities", []string{"read"}, "Capabilities to grant")
	policyCmd.Flags().BoolVar(&policyDryRun, "dry-run", false, "Validate and preview without writing")
	_ = policyCmd.MarkFlagRequired("name")
	_ = policyCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(policyCmd)
}

func runPolicy(cmd *cobra.Command, args []string) error {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = os.Getenv("VAULT_ADDR")
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create vault client: %w", err)
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))

	p := &policy.Policy{
		Name: policyName,
		Rules: []policy.Rule{
			{Path: policyPath, Capabilities: policyCapabilities},
		},
	}

	applier := policy.NewApplier(client, policyDryRun)
	result, err := applier.Apply(context.Background(), p)
	if err != nil {
		return err
	}

	if result.DryRun {
		fmt.Printf("[dry-run] policy %q validated (not written)\n", result.Name)
		fmt.Println(p.HCL())
	} else {
		fmt.Printf("policy %q applied successfully\n", result.Name)
	}
	return nil
}
