package rotate

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// SecretReader reads a secret from a path.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// SecretWriter writes a secret to a path.
type SecretWriter interface {
	WriteSecret(ctx context.Context, path string, data map[string]interface{}) error
}

// SecretReadWriter combines read and write operations.
type SecretReadWriter interface {
	SecretReader
	SecretWriter
}

// Options configures the rotation behaviour.
type Options struct {
	// KeysToRotate lists the keys within the secret to regenerate.
	KeysToRotate []string
	// ByteLength is the length of the generated random value in bytes (default 32).
	ByteLength int
	// DryRun previews changes without writing.
	DryRun bool
}

// Result holds the outcome of a rotation.
type Result struct {
	Path      string
	RotatedAt time.Time
	Keys      []string
	DryRun    bool
}

// Rotator rotates secrets at a given path.
type Rotator struct {
	client SecretReadWriter
	opts   Options
}

// New creates a new Rotator.
func New(client SecretReadWriter, opts Options) *Rotator {
	if opts.ByteLength <= 0 {
		opts.ByteLength = 32
	}
	return &Rotator{client: client, opts: opts}
}

// Rotate reads the secret at path, regenerates the configured keys and
// writes the updated secret back (unless DryRun is set).
func (r *Rotator) Rotate(ctx context.Context, path string) (*Result, error) {
	existing, err := r.client.ReadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("rotate: read %q: %w", path, err)
	}

	updated := make(map[string]interface{}, len(existing))
	for k, v := range existing {
		updated[k] = v
	}

	for _, key := range r.opts.KeysToRotate {
		newVal, err := generateSecret(r.opts.ByteLength)
		if err != nil {
			return nil, fmt.Errorf("rotate: generate value for key %q: %w", key, err)
		}
		updated[key] = newVal
	}

	if !r.opts.DryRun {
		if err := r.client.WriteSecret(ctx, path, updated); err != nil {
			return nil, fmt.Errorf("rotate: write %q: %w", path, err)
		}
	}

	return &Result{
		Path:      path,
		RotatedAt: time.Now().UTC(),
		Keys:      r.opts.KeysToRotate,
		DryRun:    r.opts.DryRun,
	}, nil
}

func generateSecret(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
