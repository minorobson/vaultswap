package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/example/vaultswap/internal/health"
	"github.com/spf13/cobra"
)

var (
	healthAddresses []string
	healthNamespace string
	healthJSON      bool
	healthTimeout   time.Duration
)

func init() {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Check the health of one or more Vault endpoints",
		RunE:  runHealth,
	}

	healthCmd.Flags().StringSliceVarP(&healthAddresses, "address", "a", nil, "Vault address(es) to check (comma-separated or repeated)")
	healthCmd.Flags().StringVarP(&healthNamespace, "namespace", "n", "", "Vault namespace to include in requests")
	healthCmd.Flags().BoolVar(&healthJSON, "json", false, "Output results as JSON")
	healthCmd.Flags().DurationVar(&healthTimeout, "timeout", 5*time.Second, "HTTP timeout per check")
	_ = healthCmd.MarkFlagRequired("address")

	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), healthTimeout*time.Duration(len(healthAddresses)))
	defer cancel()

	checker := health.New(healthTimeout)

	targets := make([]health.Target, len(healthAddresses))
	for i, addr := range healthAddresses {
		targets[i] = health.Target{Address: strings.TrimSpace(addr), Namespace: healthNamespace}
	}

	statuses := checker.CheckMany(ctx, targets)

	if healthJSON {
		return json.NewEncoder(os.Stdout).Encode(statuses)
	}

	allHealthy := true
	for _, s := range statuses {
		switch {
		case s.Error != "":
			fmt.Fprintf(os.Stdout, "[ERROR]   %s — %s\n", s.Address, s.Error)
			allHealthy = false
		case s.Sealed:
			fmt.Fprintf(os.Stdout, "[SEALED]  %s (v%s)\n", s.Address, s.Version)
			allHealthy = false
		case s.Healthy:
			fmt.Fprintf(os.Stdout, "[OK]      %s (v%s)\n", s.Address, s.Version)
		default:
			fmt.Fprintf(os.Stdout, "[UNKNOWN] %s\n", s.Address)
			allHealthy = false
		}
	}

	if !allHealthy {
		return fmt.Errorf("one or more Vault endpoints are unhealthy")
	}
	return nil
}
