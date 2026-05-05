package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nicholasgasior/vaultswap/internal/archive"
	"github.com/nicholasgasior/vaultswap/internal/audit"
	"github.com/nicholasgasior/vaultswap/internal/vault"
	"github.com/spf13/cobra"
)

var (
	archiveBase   string
	archiveDryRun bool
)

func init() {
	archiveCmd := &cobra.Command{
		Use:   "archive [paths...]",
		Short: "Archive current secret versions to a timestamped path in Vault",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runArchive,
	}

	archiveCmd.Flags().StringVar(&archiveBase, "base", "archive", "Base path for archived secrets")
	archiveCmd.Flags().BoolVar(&archiveDryRun, "dry-run", false, "Preview archive without writing")

	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   envOrDefault("VAULT_ADDR", ""),
		Token:     envOrDefault("VAULT_TOKEN", ""),
		Namespace: envOrDefault("VAULT_NAMESPACE", ""),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	logger, err := audit.New("")
	if err != nil {
		return fmt.Errorf("audit logger: %w", err)
	}

	a := archive.New(client, archiveDryRun)
	results := a.ArchivePaths(context.Background(), args, archiveBase)

	archive.PrintResults(results)

	for _, r := range results {
		if r.Err != nil {
			logger.Log(audit.Entry{
				Operation: "archive",
				Path:      r.Path,
				DryRun:    archiveDryRun,
				Error:     r.Err.Error(),
			})
		} else {
			logger.Log(audit.Entry{
				Operation: "archive",
				Path:      r.Path,
				DryRun:    archiveDryRun,
			})
		}
	}

	for _, r := range results {
		if r.Err != nil {
			os.Exit(1)
		}
	}
	return nil
}
