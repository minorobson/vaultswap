package health

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Status represents the health status of a Vault endpoint.
type Status struct {
	Address   string `json:"address"`
	Namespace string `json:"namespace,omitempty"`
	Healthy   bool   `json:"healthy"`
	Sealed    bool   `json:"sealed"`
	Version   string `json:"version"`
	Error     string `json:"error,omitempty"`
}

// Checker checks the health of one or more Vault addresses.
type Checker struct {
	client  *http.Client
	timeout time.Duration
}

// New creates a new Checker with the given timeout.
func New(timeout time.Duration) *Checker {
	return &Checker{
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

// Check performs a health check against the given Vault address and optional namespace.
func (c *Checker) Check(ctx context.Context, address, namespace string) Status {
	status := Status{
		Address:   address,
		Namespace: namespace,
	}

	url := fmt.Sprintf("%s/v1/sys/health", address)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		status.Error = fmt.Sprintf("build request: %v", err)
		return status
	}

	if namespace != "" {
		req.Header.Set("X-Vault-Namespace", namespace)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		status.Error = fmt.Sprintf("request failed: %v", err)
		return status
	}
	defer resp.Body.Close()

	// Vault returns 200 (active), 429 (standby), 472/473 (DR/perf), 501/503 (sealed/uninit)
	switch resp.StatusCode {
	case http.StatusOK, 429, 472, 473:
		status.Healthy = true
		status.Version = resp.Header.Get("X-Vault-Version")
	case 501, 503:
		status.Sealed = true
		status.Version = resp.Header.Get("X-Vault-Version")
	default:
		status.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
	}

	return status
}

// CheckMany checks multiple addresses concurrently and returns all statuses.
func (c *Checker) CheckMany(ctx context.Context, targets []Target) []Status {
	results := make([]Status, len(targets))
	ch := make(chan indexedStatus, len(targets))

	for i, t := range targets {
		go func(idx int, target Target) {
			ch <- indexedStatus{idx: idx, status: c.Check(ctx, target.Address, target.Namespace)}
		}(i, t)
	}

	for range targets {
		is := <-ch
		results[is.idx] = is.status
	}
	return results
}

// Target holds an address and optional namespace for a health check.
type Target struct {
	Address   string
	Namespace string
}

type indexedStatus struct {
	idx    int
	status Status
}
