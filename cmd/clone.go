package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultswap/internal/clone"
	"github.com/vaultswap/internal/vault"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone secrets from one Vault namespace to another",
	RunE:  runClone,
}

func init() {
	cloneCmd.Flags().String("src-addr", "", "Source Vault address (required)")
	cloneCmd.Flags().String("src-token", "", "Source Vault token (required)")
	cloneCmd.Flags().String("src-namespace", "", "Source Vault namespace")
	cloneCmd.Flags().String("dst-addr", "", "Destination Vault address (required)")
	cloneCmd.Flags().String("dst-token", "", "Destination Vault token (required)")
	cloneCmd.Flags().String("dst-namespace", "", "Destination Vault namespace")
	cloneCmd.Flags().StringSlice("paths", nil, "Comma-separated secret paths to clone (required)")
	cloneCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, _ []string) error {
	srcAddr, _ := cmd.Flags().GetString("src-addr")
	srcToken, _ := cmd.Flags().GetString("src-token")
	srcNS, _ := cmd.Flags().GetString("src-namespace")
	dstAddr, _ := cmd.Flags().GetString("dst-addr")
	dstToken, _ := cmd.Flags().GetString("dst-token")
	dstNS, _ := cmd.Flags().GetString("dst-namespace")
	paths, _ := cmd.Flags().GetStringSlice("paths")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if srcAddr == "" || srcToken == "" || dstAddr == "" || dstToken == "" {
		return fmt.Errorf("--src-addr, --src-token, --dst-addr and --dst-token are required")
	}
	if len(paths) == 0 {
		return fmt.Errorf("--paths is required")
	}

	src, err := vault.NewClient(srcAddr, srcToken, srcNS)
	if err != nil {
		return fmt.Errorf("src client: %w", err)
	}
	dst, err := vault.NewClient(dstAddr, dstToken, dstNS)
	if err != nil {
		return fmt.Errorf("dst client: %w", err)
	}

	if dryRun {
		log.Println("[dry-run] no secrets will be written")
	}

	c := clone.New(src, dst, dryRun)
	results := c.ClonePaths(context.Background(), cleanPaths(paths))
	clone.PrintResults(results)

	for _, r := range results {
		if r.Err != nil {
			return fmt.Errorf("clone failed for %s: %w", r.Path, r.Err)
		}
	}
	return nil
}

func cleanPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
