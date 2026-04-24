package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourorg/vaultswap/internal/diff"
	"github.com/yourorg/vaultswap/internal/sync"
	"github.com/yourorg/vaultswap/internal/vault"
)

var (
	syncSrcAddr      string
	syncSrcToken     string
	syncSrcNamespace string
	syncDstAddr      string
	syncDstToken     string
	syncDstNamespace string
	syncPath         string
	syncDryRun       bool
	syncMaskValues   bool
)

// syncCmd represents the sync subcommand which copies secrets from a source
// Vault path to a destination Vault path, optionally showing a diff preview.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets from one Vault namespace to another",
	Long: `sync reads secrets from a source Vault instance and writes them to a
destination Vault instance. A diff preview is printed before any writes are
performed. Use --dry-run to preview changes without applying them.`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Source Vault flags
	syncCmd.Flags().StringVar(&syncSrcAddr, "src-addr", "", "Source Vault address (env: VAULT_SRC_ADDR)")
	syncCmd.Flags().StringVar(&syncSrcToken, "src-token", "", "Source Vault token (env: VAULT_SRC_TOKEN)")
	syncCmd.Flags().StringVar(&syncSrcNamespace, "src-namespace", "", "Source Vault namespace (optional)")

	// Destination Vault flags
	syncCmd.Flags().StringVar(&syncDstAddr, "dst-addr", "", "Destination Vault address (env: VAULT_DST_ADDR)")
	syncCmd.Flags().StringVar(&syncDstToken, "dst-token", "", "Destination Vault token (env: VAULT_DST_TOKEN)")
	syncCmd.Flags().StringVar(&syncDstNamespace, "dst-namespace", "", "Destination Vault namespace (optional)")

	// Shared flags
	syncCmd.Flags().StringVar(&syncPath, "path", "", "KV v2 secret path to sync (required)")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without writing to destination")
	syncCmd.Flags().BoolVar(&syncMaskValues, "mask-values", true, "Mask secret values in diff output")

	_ = syncCmd.MarkFlagRequired("path")
}

func runSync(cmd *cobra.Command, args []string) error {
	// Fall back to environment variables when flags are not provided.
	if syncSrcAddr == "" {
		syncSrcAddr = os.Getenv("VAULT_SRC_ADDR")
	}
	if syncSrcToken == "" {
		syncSrcToken = os.Getenv("VAULT_SRC_TOKEN")
	}
	if syncDstAddr == "" {
		syncDstAddr = os.Getenv("VAULT_DST_ADDR")
	}
	if syncDstToken == "" {
		syncDstToken = os.Getenv("VAULT_DST_TOKEN")
	}

	srcClient, err := vault.NewClient(vault.Config{
		Address:   syncSrcAddr,
		Token:     syncSrcToken,
		Namespace: syncSrcNamespace,
	})
	if err != nil {
		return fmt.Errorf("initialising source Vault client: %w", err)
	}

	dstClient, err := vault.NewClient(vault.Config{
		Address:   syncDstAddr,
		Token:     syncDstToken,
		Namespace: syncDstNamespace,
	})
	if err != nil {
		return fmt.Errorf("initialising destination Vault client: %w", err)
	}

	syncer := sync.New(srcClient, dstClient)

	result, err := syncer.Sync(cmd.Context(), syncPath, syncDryRun)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Print a human-readable diff of the changes.
	formatted := diff.Format(result.Changes, syncMaskValues)
	if formatted != "" {
		fmt.Println(formatted)
	}

	if syncDryRun {
		fmt.Println("[dry-run] no changes written to destination")
	} else if len(result.Changes) == 0 {
		fmt.Println("destination is already up to date — no changes applied")
	} else {
		fmt.Printf("synced %d key(s) to %s\n", len(result.Changes), syncPath)
	}

	return nil
}
