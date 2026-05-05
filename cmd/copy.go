package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"vaultswap/internal/copy"
	"vaultswap/internal/vault"
)

func init() {
	copyCmd := &cobra.Command{
		Use:   "copy SRC:DST [SRC:DST ...]",
		Short: "Copy secrets from one path to another",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runCopy,
	}
	copyCmd.Flags().Bool("dry-run", false, "Preview copy without writing")
	root.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	addr := envOrDefault("VAULT_ADDR", "")
	token := envOrDefault("VAULT_TOKEN", "")
	ns := envOrDefault("VAULT_NAMESPACE", "")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	client, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: ns,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	pairs, err := parseCopyPairs(args)
	if err != nil {
		return err
	}

	cp := copy.New(client, dryRun)
	results := cp.CopyPaths(context.Background(), pairs)
	copy.PrintResults(results)
	return nil
}

func parseCopyPairs(args []string) ([][2]string, error) {
	pairs := make([][2]string, 0, len(args))
	for _, arg := range args {
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid pair %q: expected SRC:DST format", arg)
		}
		pairs = append(pairs, [2]string{parts[0], parts[1]})
	}
	return pairs, nil
}
