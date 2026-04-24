package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/vaultswap/internal/rotate"
	"github.com/example/vaultswap/internal/vault"
	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <path>",
	Short: "Rotate one or more secret keys at a Vault path",
	Args:  cobra.ExactArgs(1),
	RunE:  runRotate,
}

var (
	rotateKeys    []string
	rotateBytes   int
	rotateDryRun  bool
)

func init() {
	rotateCmd.Flags().StringSliceVarP(&rotateKeys, "keys", "k", nil, "Comma-separated list of keys to rotate (required)")
	rotateCmd.Flags().IntVar(&rotateBytes, "bytes", 32, "Number of random bytes for each new secret value")
	rotateCmd.Flags().BoolVar(&rotateDryRun, "dry-run", false, "Preview rotation without writing to Vault")
	_ = rotateCmd.MarkFlagRequired("keys")

	rootCmd.AddCommand(rotateCmd)
}

func runRotate(cmd *cobra.Command, args []string) error {
	path := args[0]

	if len(rotateKeys) == 0 {
		return fmt.Errorf("at least one key must be specified via --keys")
	}

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	r := rotate.New(client, rotate.Options{
		KeysToRotate: rotateKeys,
		ByteLength:   rotateBytes,
		DryRun:       rotateDryRun,
	})

	result, err := r.Rotate(cmd.Context(), path)
	if err != nil {
		return err
	}

	rotate.PrintResult(cmd.OutOrStdout(), result)

	if rotateDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "\n[dry-run] keys that would be rotated: %s\n", strings.Join(rotateKeys, ", "))
	}
	return nil
}
