package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultswap/internal/audit"
	"github.com/vaultswap/internal/rollback"
	"github.com/vaultswap/internal/vault"
)

var (
	rollbackSnapshotFile string
	rollbackDryRun       bool
	rollbackAuditLog     string
)

func init() {
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Restore secrets from a snapshot file",
		RunE:  runRollback,
	}

	rollbackCmd.Flags().StringVar(&rollbackSnapshotFile, "snapshot", "", "Path to snapshot file (required)")
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "Preview restore without writing")
	rollbackCmd.Flags().StringVar(&rollbackAuditLog, "audit-log", "", "Path to audit log file (default: stdout)")
	_ = rollbackCmd.MarkFlagRequired("snapshot")

	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, _ []string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	namespace := os.Getenv("VAULT_NAMESPACE")

	client, err := vault.NewClient(addr, token, namespace)
	if err != nil {
		log.Fatalf("vault client: %v", err)
	}

	logger, err := audit.New(rollbackAuditLog)
	if err != nil {
		return err
	}

	rb := rollback.New(client, logger, rollbackDryRun)

	results, err := rb.Rollback(cmd.Context(), rollbackSnapshotFile)
	if err != nil {
		return err
	}

	rollback.PrintResults(results, os.Stdout)
	return nil
}
