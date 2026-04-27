package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/vaultswap/internal/audit"
	"github.com/user/vaultswap/internal/export"
	"github.com/user/vaultswap/internal/vault"
)

var (
	exportOutput string
	exportFormat string
	exportDryRun bool
)

func init() {
	exportCmd := &cobra.Command{
		Use:   "export [paths...]",
		Short: "Export Vault secrets to a local file (JSON or YAML)",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runExport,
	}
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "secrets.json", "destination file path")
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "output format: json or yaml")
	exportCmd.Flags().BoolVar(&exportDryRun, "dry-run", false, "preview without writing")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	addr := os.Getenv("VAULT_ADDR")
	token := os.Getenv("VAULT_TOKEN")
	ns := os.Getenv("VAULT_NAMESPACE")

	if addr == "" {
		return fmt.Errorf("VAULT_ADDR is required")
	}
	if token == "" {
		return fmt.Errorf("VAULT_TOKEN is required")
	}

	c, err := vault.NewClient(vault.Config{
		Address:   addr,
		Token:     token,
		Namespace: ns,
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	paths := splitExportPaths(args)
	ex := export.New(c)
	results := ex.ExportPaths(paths)

	export.PrintResults(results, exportOutput, exportFormat, exportDryRun)

	logger, _ := audit.New("")
	logger.Log(audit.Entry{
		Operation: "export",
		Paths:     paths,
		DryRun:    exportDryRun,
		Details:   fmt.Sprintf("format=%s dest=%s", exportFormat, exportOutput),
	})

	if exportDryRun {
		return nil
	}

	hasErr := false
	for _, r := range results {
		if r.Error != nil {
			hasErr = true
		}
	}

	if !hasErr {
		if err := export.WriteFile(results, exportOutput, exportFormat); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}

func splitExportPaths(args []string) []string {
	var out []string
	for _, a := range args {
		for _, p := range strings.Split(a, ",") {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
	}
	return out
}
