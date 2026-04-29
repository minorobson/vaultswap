package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/seatgeek/vaultswap/internal/search"
	"github.com/seatgeek/vaultswap/internal/vault"
	"github.com/spf13/cobra"
)

var (
	searchQuery    string
	searchPaths    []string
	searchKeysOnly bool
	searchMask     bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search secrets across Vault paths by key or value",
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "search query string (required)")
	searchCmd.Flags().StringSliceVar(&searchPaths, "paths", nil, "comma-separated list of secret paths to search")
	searchCmd.Flags().BoolVar(&searchKeysOnly, "keys-only", false, "match against keys only, not values")
	searchCmd.Flags().BoolVar(&searchMask, "mask", false, "mask secret values in output")
	_ = searchCmd.MarkFlagRequired("query")
	_ = searchCmd.MarkFlagRequired("paths")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, _ []string) error {
	if strings.TrimSpace(searchQuery) == "" {
		return fmt.Errorf("--query must not be empty")
	}

	address := envOrDefault("VAULT_ADDR", "")
	token := envOrDefault("VAULT_TOKEN", "")
	namespace := envOrDefault("VAULT_NAMESPACE", "")

	client, err := vault.NewClient(address, token, namespace)
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	s := search.New(client)
	results := s.SearchPaths(context.Background(), searchPaths, searchQuery, searchKeysOnly)

	search.PrintResults(results, searchMask)

	for _, r := range results {
		if r.Err != nil {
			os.Exit(1)
		}
	}
	return nil
}
