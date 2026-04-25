package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot represents a point-in-time capture of secrets at a given Vault path.
type Snapshot struct {
	Path      string            `json:"path"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	CapturedAt time.Time        `json:"captured_at"`
}

// Taker captures and persists secret snapshots.
type Taker struct {
	reader SecretReader
}

// SecretReader is satisfied by vault.Client.
type SecretReader interface {
	ReadSecret(namespace, path string) (map[string]string, error)
}

// New returns a new Taker backed by the provided SecretReader.
func New(r SecretReader) *Taker {
	return &Taker{reader: r}
}

// Take reads the secret at the given namespace/path and returns a Snapshot.
func (t *Taker) Take(namespace, path string) (*Snapshot, error) {
	data, err := t.reader.ReadSecret(namespace, path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read secret: %w", err)
	}
	return &Snapshot{
		Path:       path,
		Namespace:  namespace,
		Data:       data,
		CapturedAt: time.Now().UTC(),
	}, nil
}

// Save writes the snapshot as JSON to the given file path.
func (s *Snapshot) Save(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("snapshot: encode json: %w", err)
	}
	return nil
}

// Load reads a snapshot from a JSON file.
func Load(filePath string) (*Snapshot, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open file: %w", err)
	}
	defer f.Close()

	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("snapshot: decode json: %w", err)
	}
	return &s, nil
}

// KeyNames returns a sorted list of secret key names present in the snapshot.
// This is useful for comparing the structure of two snapshots without exposing values.
func (s *Snapshot) KeyNames() []string {
	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
