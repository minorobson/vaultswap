package watch

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/vaultswap/internal/diff"
	"github.com/your-org/vaultswap/internal/vault"
)

// Result holds the outcome of a single watch poll cycle.
type Result struct {
	Path    string
	Changed bool
	Diff    string
	Err     error
}

// Watcher polls one or more KV paths and reports when their contents change.
type Watcher struct {
	client   *vault.Client
	paths    []string
	interval time.Duration
	mask     bool
}

// New creates a Watcher for the given paths and poll interval.
func New(client *vault.Client, paths []string, interval time.Duration, mask bool) *Watcher {
	return &Watcher{
		client:   client,
		paths:    paths,
		interval: interval,
		mask:     mask,
	}
}

// Watch starts polling until ctx is cancelled. Each detected change is sent on
// the returned channel. The channel is closed when the context ends.
func (w *Watcher) Watch(ctx context.Context) <-chan Result {
	ch := make(chan Result, len(w.paths))
	go func() {
		defer close(ch)
		prev := make(map[string]map[string]string)
		// seed initial state
		for _, p := range w.paths {
			secret, err := w.client.ReadSecret(ctx, p)
			if err == nil {
				prev[p] = toStringMap(secret)
			}
		}
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, p := range w.paths {
					secret, err := w.client.ReadSecret(ctx, p)
					if err != nil {
						ch <- Result{Path: p, Err: fmt.Errorf("read: %w", err)}
						continue
					}
					curr := toStringMap(secret)
					changes := diff.Compare(prev[p], curr)
					hasChange := false
					for _, c := range changes {
						if c.Status != diff.StatusUnchanged {
							hasChange = true
							break
						}
					}
					if hasChange {
						ch <- Result{
							Path:    p,
							Changed: true,
							Diff:    diff.Format(changes, w.mask),
						}
						prev[p] = curr
					}
				}
			}
		}
	}()
	return ch
}

func toStringMap(m map[string]interface{}) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}
