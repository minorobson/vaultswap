package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultswap/internal/promote"
	"github.com/your-org/vaultswap/internal/vault"
)

var (
	promoteSrcAddr      string
	promoteSrcToken     string
	promoteSrcNamespace string
	promoteDstAddr      string
	promoteDstToken     string
	promoteDstNamespace string
	promotePaths        string
	promoteDryRun       bool
)

var promoteCmd = &cobra.Command{
	Use:   "promote",
	Short: "Promote secrets from one Vault namespace to another",
	RunE:  runPromote,
}

func init() {
	promoteCmd.Flags().StringVar(&promoteSrcAddr, "src-addr", "", "source Vault address")
	promoteCmd.Flags().StringVar(&promoteSrcToken, "src-token", "", "source Vault token")
	promoteCmd.Flags().StringVar(&promoteSrcNamespace, "src-namespace", "", "source Vault namespace")
	promoteCmd.Flags().StringVar(&promoteDstAddr, "dst-addr", "", "destination Vault address")
	promoteCmd.Flags().StringVar(&promoteDstToken, "dst-token", "", "destination Vault token")
	promoteCmd.Flags().StringVar(&promoteDstNamespace, "dst-namespace", "", "destination Vault namespace")
	promoteCmd.Flags().StringVar(&promotePaths, "paths", "", "comma-separated list of secret paths")
	promoteCmd.Flags().BoolVar(&promoteDryRun, "dry-run", false, "preview changes without writing")
	rootCmd.AddCommand(promoteCmd)
}

func runPromote(cmd *cobra.Command, args []string) error {
	src, err := vault.NewClient(vault.Config{
		Address:   promoteSrcAddr,
		Token:     promoteSrcToken,
		Namespace: promoteSrcNamespace,
	})
	if err != nil {
		return err
	}

	dst, err := vault.NewClient(vault.Config{
		Address:   promoteDstAddr,
		Token:     promoteDstToken,
		Namespace: promoteDstNamespace,
	})
	if err != nil {
		return err
	}

	paths := splitPaths(promotePaths)
	if len(paths) == 0 {
		log.Println("no paths specified")
		return nil
	}

	p := promote.New(src, dst, promoteDryRun)
	results, err := p.Promote(context.Background(), paths)
	if err != nil {
		return err
	}
	promote.PrintResults(results)
	return nil
}

func splitPaths(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
