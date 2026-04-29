package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultswap/internal/vault"
	"github.com/your-org/vaultswap/internal/watch"
)

var (
	watchInterval time.Duration
	watchMask     bool
)

func init() {
	watchCmd := &cobra.Command{
		Use:   "watch <path> [path...]",
		Short: "Poll Vault paths and print diffs when secrets change",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runWatch,
	}
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 30*time.Second, "poll interval (e.g. 10s, 1m)")
	watchCmd.Flags().BoolVar(&watchMask, "mask", true, "mask secret values in diff output")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	client, err := vault.NewClient(vault.Config{
		Address:   envOrDefault("VAULT_ADDR", ""),
		Token:     envOrDefault("VAULT_TOKEN", ""),
		Namespace: envOrDefault("VAULT_NAMESPACE", ""),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	paths := args
	w := watch.New(client, paths, watchInterval, watchMask)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(os.Stdout, "Watching %s (interval: %s) — press Ctrl+C to stop\n",
		strings.Join(paths, ", "), watchInterval)

	for result := range w.Watch(ctx) {
		if result.Err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", result.Path, result.Err)
			continue
		}
		if result.Changed {
			fmt.Fprintf(os.Stdout, "\n[CHANGED] %s at %s\n%s\n",
				result.Path, time.Now().Format(time.RFC3339), result.Diff)
		}
	}
	fmt.Fprintln(os.Stdout, "Watch stopped.")
	return nil
}
