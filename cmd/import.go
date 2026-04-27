package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	importpkg "github.com/your-org/vaultswap/internal/import"
	"github.com/your-org/vaultswap/internal/vault"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import secrets from a JSON file into Vault",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().String("address", "", "Vault address (or VAULT_ADDR)")
	importCmd.Flags().String("token", "", "Vault token (or VAULT_TOKEN)")
	importCmd.Flags().String("namespace", "", "Vault namespace")
	importCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	addr, _ := cmd.Flags().GetString("address")
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}
	namespace, _ := cmd.Flags().GetString("namespace")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	client, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	im := importpkg.New(client)
	results, err := im.ImportFile(context.Background(), filePath, dryRun)
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(cmd.ErrOrStderr(), "ERROR  %s: %v\n", r.Path, r.Err)
		case r.DryRun:
			fmt.Fprintf(cmd.OutOrStdout(), "DRY-RUN  %s\n", r.Path)
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "WRITTEN  %s\n", r.Path)
		}
	}
	return nil
}
