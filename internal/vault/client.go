package vault

import (
	"errors"
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the HashiCorp Vault API client with namespace support.
type Client struct {
	underlying *vaultapi.Client
	namespace  string
}

// Config holds the configuration needed to connect to a Vault instance.
type Config struct {
	Address   string
	Token     string
	Namespace string
}

// NewClient creates and configures a new Vault client for the given config.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("vault address must not be empty")
	}
	if cfg.Token == "" {
		return nil, errors.New("vault token must not be empty")
	}

	apiCfg := vaultapi.DefaultConfig()
	apiCfg.Address = cfg.Address

	underlying, err := vaultapi.NewClient(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	underlying.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		underlying.SetNamespace(cfg.Namespace)
	}

	return &Client{
		underlying: underlying,
		namespace:  cfg.Namespace,
	}, nil
}

// ReadSecret reads a KV v2 secret at the given path and returns its data map.
func (c *Client) ReadSecret(path string) (map[string]interface{}, error) {
	secret, err := c.underlying.Logical().Read(kvV2DataPath(path))
	if err != nil {
		return nil, fmt.Errorf("reading secret %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format at path %q", path)
	}

	return data, nil
}

// WriteSecret writes key/value pairs to a KV v2 secret at the given path.
func (c *Client) WriteSecret(path string, data map[string]interface{}) error {
	payload := map[string]interface{}{"data": data}
	_, err := c.underlying.Logical().Write(kvV2DataPath(path), payload)
	if err != nil {
		return fmt.Errorf("writing secret %q: %w", path, err)
	}
	return nil
}

// Namespace returns the namespace this client is scoped to.
func (c *Client) Namespace() string {
	return c.namespace
}

// kvV2DataPath converts a logical path to the KV v2 data path.
func kvV2DataPath(path string) string {
	return "secret/data/" + path
}
