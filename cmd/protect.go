package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/protect"
	"github.com/yourusername/vaultswap/internal/vault"
)

var (
	protectMount  string
	protectPaths  []string
	protectDryRun bool
)

func init() {
	protectCmd := &cobra.Command{
		Use:   "protect",
		Short: "Mark secrets as protected to prevent destructive operations",
		RunE:  runProtect,
	}

	protectCmd.Flags().StringVar(&protectMount, "mount", "secret", "KV v2 mount path")
	protectCmd.Flags().StringArrayVar(&protectPaths, "path", nil, "secret path(s) to protect (repeatable)")
	protectCmd.Flags().BoolVar(&protectDryRun, "dry-run", false, "preview changes without writing")
	_ = protectCmd.MarkFlagRequired("path")

	rootCmd.AddCommand(protectCmd)
}

func runProtect(cmd *cobra.Command, _ []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   envOrDefault("VAULT_ADDR", ""),
		Token:     envOrDefault("VAULT_TOKEN", ""),
		Namespace: envOrDefault("VAULT_NAMESPACE", ""),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	p := protect.New(client, protectDryRun)
	results := p.ProtectPaths(context.Background(), protectMount, protectPaths)
	protect.PrintResults(results)

	var errs []string
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, r.Error.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("protect completed with errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
