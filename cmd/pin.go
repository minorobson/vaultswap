package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vaultswap/vaultswap/internal/pin"
	"github.com/vaultswap/vaultswap/internal/vault"
)

func init() {
	pinCmd := &cobra.Command{
		Use:   "pin [paths...]",
		Short: "Pin KV v2 secrets to a specific version",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runPin,
	}
	pinCmd.Flags().String("mount", "secret", "KV v2 mount path")
	pinCmd.Flags().Int("version", 0, "Version to pin (required)")
	pinCmd.Flags().Bool("dry-run", false, "Preview changes without writing")
	_ = pinCmd.MarkFlagRequired("version")
	rootCmd.AddCommand(pinCmd)
}

func runPin(cmd *cobra.Command, args []string) error {
	addr := envOrDefault("VAULT_ADDR", "")
	token := envOrDefault("VAULT_TOKEN", "")
	ns := envOrDefault("VAULT_NAMESPACE", "")

	client, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: ns,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	mount, _ := cmd.Flags().GetString("mount")
	version, _ := cmd.Flags().GetInt("version")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	paths := normalisePaths(args)

	p := pin.New(client, dryRun)
	results := p.PinPaths(context.Background(), mount, paths, version)
	pin.PrintResults(results)

	for _, r := range results {
		if r.Err != nil {
			os.Exit(1)
		}
	}
	return nil
}

func normalisePaths(args []string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		out = append(out, strings.TrimPrefix(a, "/"))
	}
	return out
}
