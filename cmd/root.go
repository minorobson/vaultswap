// Package cmd provides the CLI commands for vaultswap.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the vaultswap CLI.
var rootCmd = &cobra.Command{
	Use:   "vaultswap",
	Short: "Rotate and sync secrets across HashiCorp Vault namespaces",
	Long: `vaultswap is a CLI tool for rotating and syncing secrets across
HashiCorp Vault namespaces. It supports dry-run previews, diff output,
audit logging, snapshots, rollback, and more.

Common flags:
  --address   Vault server address (or VAULT_ADDR env var)
  --token     Vault token (or VAULT_TOKEN env var)
  --namespace Vault namespace (or VAULT_NAMESPACE env var)
  --dry-run   Preview changes without writing to Vault`,
	SilenceUsage: true,
}

// globalFlags holds values for persistent flags shared across all subcommands.
var globalFlags struct {
	Address   string
	Token     string
	Namespace string
	DryRun    bool
	Output    string
}

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Persistent flags are inherited by all subcommands.
	rootCmd.PersistentFlags().StringVar(
		&globalFlags.Address,
		"address",
		envOrDefault("VAULT_ADDR", "http://127.0.0.1:8200"),
		"Vault server address",
	)
	rootCmd.PersistentFlags().StringVar(
		&globalFlags.Token,
		"token",
		envOrDefault("VAULT_TOKEN", ""),
		"Vault token for authentication",
	)
	rootCmd.PersistentFlags().StringVar(
		&globalFlags.Namespace,
		"namespace",
		envOrDefault("VAULT_NAMESPACE", ""),
		"Vault namespace (Enterprise only)",
	)
	rootCmd.PersistentFlags().BoolVar(
		&globalFlags.DryRun,
		"dry-run",
		false,
		"Preview changes without writing to Vault",
	)
	rootCmd.PersistentFlags().StringVar(
		&globalFlags.Output,
		"output",
		"text",
		"Output format: text or json",
	)
}

// envOrDefault returns the value of the named environment variable,
// or the provided default if the variable is unset or empty.
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
