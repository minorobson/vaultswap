package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/fhemberger/vaultswap/internal/mirror"
	"github.com/fhemberger/vaultswap/internal/vault"
	"github.com/spf13/cobra"
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror secrets from a source namespace to a destination namespace",
	RunE:  runMirror,
}

func init() {
	mirrorCmd.Flags().String("src-namespace", "", "source Vault namespace")
	mirrorCmd.Flags().String("dst-namespace", "", "destination Vault namespace")
	mirrorCmd.Flags().StringSlice("paths", nil, "comma-separated src:dst path pairs (e.g. secret/a:secret/b)")
	mirrorCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	mirrorCmd.MarkFlagRequired("paths") //nolint:errcheck
	RootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, _ []string) error {
	srcNS, _ := cmd.Flags().GetString("src-namespace")
	dstNS, _ := cmd.Flags().GetString("dst-namespace")
	rawPaths, _ := cmd.Flags().GetStringSlice("paths")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	srcClient, err := vault.NewClient(vault.Config{
		Address:   envOrDefault("VAULT_ADDR", ""),
		Token:     envOrDefault("VAULT_TOKEN", ""),
		Namespace: srcNS,
	})
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}

	dstClient, err := vault.NewClient(vault.Config{
		Address:   envOrDefault("VAULT_ADDR", ""),
		Token:     envOrDefault("VAULT_TOKEN", ""),
		Namespace: dstNS,
	})
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	pairs, err := parseMirrorPairs(rawPaths)
	if err != nil {
		return err
	}

	m := mirror.New(srcClient, dstClient, dryRun)
	results := m.MirrorPaths(context.Background(), pairs)
	mirror.PrintResults(results)
	return nil
}

func parseMirrorPairs(raw []string) ([][2]string, error) {
	pairs := make([][2]string, 0, len(raw))
	for _, p := range raw {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid path pair %q: expected src:dst", p)
		}
		pairs = append(pairs, [2]string{parts[0], parts[1]})
	}
	return pairs, nil
}
