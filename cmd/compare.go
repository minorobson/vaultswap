package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vaultswap/internal/compare"
	"github.com/vaultswap/internal/vault"
)

var (
	cmpSrcAddr  string
	cmpSrcToken string
	cmpDstAddr  string
	cmpDstToken string
	cmpPaths    string
	cmpMask     bool
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Diff secrets between two Vault instances",
	RunE:  runCompare,
}

func init() {
	compareCmd.Flags().StringVar(&cmpSrcAddr, "src-addr", "", "Source Vault address (required)")
	compareCmd.Flags().StringVar(&cmpSrcToken, "src-token", "", "Source Vault token (required)")
	compareCmd.Flags().StringVar(&cmpDstAddr, "dst-addr", "", "Destination Vault address (required)")
	compareCmd.Flags().StringVar(&cmpDstToken, "dst-token", "", "Destination Vault token (required)")
	compareCmd.Flags().StringVar(&cmpPaths, "paths", "", "Comma-separated list of secret paths")
	compareCmd.Flags().BoolVar(&cmpMask, "mask", true, "Mask secret values in output")
	_ = compareCmd.MarkFlagRequired("src-addr")
	_ = compareCmd.MarkFlagRequired("src-token")
	_ = compareCmd.MarkFlagRequired("dst-addr")
	_ = compareCmd.MarkFlagRequired("dst-token")
	_ = compareCmd.MarkFlagRequired("paths")
	rootCmd.AddCommand(compareCmd)
}

func runCompare(cmd *cobra.Command, _ []string) error {
	src, err := vault.NewClient(cmpSrcAddr, cmpSrcToken, "")
	if err != nil {
		return fmt.Errorf("source client: %w", err)
	}
	dst, err := vault.NewClient(cmpDstAddr, cmpDstToken, "")
	if err != nil {
		return fmt.Errorf("destination client: %w", err)
	}

	paths := strings.Split(cmpPaths, ",")
	for i, p := range paths {
		paths[i] = strings.TrimSpace(p)
	}

	cmp := compare.New(src, dst, cmpMask)
	results, err := cmp.ComparePaths(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	compare.PrintResults(results, cmpMask)
	return nil
}
