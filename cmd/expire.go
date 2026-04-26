package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/vaultswap/internal/expire"
	"github.com/vaultswap/internal/vault"
)

var (
	expirePaths     []string
	expireNamespace string
	expireWarnTTL   time.Duration
)

func init() {
	expireCmd := &cobra.Command{
		Use:   "expire",
		Short: "Check secret expiry across one or more Vault paths",
		RunE:  runExpire,
	}
	expireCmd.Flags().StringSliceVar(&expirePaths, "paths", nil, "Secret paths to check (required)")
	expireCmd.Flags().StringVar(&expireNamespace, "namespace", "", "Vault namespace")
	expireCmd.Flags().DurationVar(&expireWarnTTL, "warn-ttl", 0, "Warn if remaining TTL is below this threshold (e.g. 72h)")
	_ = expireCmd.MarkFlagRequired("paths")
	rootCmd.AddCommand(expireCmd)
}

func runExpire(cmd *cobra.Command, _ []string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")

	client, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: expireNamespace,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	checker := expire.New(client)
	results := checker.CheckPaths(cmd.Context(), expirePaths)

	expire.PrintResults(results)

	for _, r := range results {
		if r.Err != nil || r.Expired {
			return fmt.Errorf("one or more secrets are expired or unreachable")
		}
		if expireWarnTTL > 0 && !r.NoTTL && time.Until(r.ExpiresAt) < expireWarnTTL {
			return fmt.Errorf("one or more secrets expire within the warning threshold (%s)", expireWarnTTL)
		}
	}
	return nil
}
