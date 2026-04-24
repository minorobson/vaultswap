# vaultswap

> CLI tool for rotating and syncing secrets across HashiCorp Vault namespaces with diff previews

---

## Installation

```bash
go install github.com/yourusername/vaultswap@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultswap.git
cd vaultswap
go build -o vaultswap .
```

---

## Usage

Sync secrets from one Vault namespace to another with a diff preview before applying:

```bash
# Preview changes between namespaces
vaultswap diff --src namespace/prod --dst namespace/staging

# Rotate and sync secrets
vaultswap sync --src namespace/prod --dst namespace/staging

# Rotate a specific secret and propagate changes
vaultswap rotate --path secret/myapp/db-password --namespaces staging,qa,prod
```

### Flags

| Flag | Description |
|------|-------------|
| `--src` | Source Vault namespace path |
| `--dst` | Destination Vault namespace path |
| `--dry-run` | Preview changes without applying |
| `--token` | Vault token (or set `VAULT_TOKEN`) |
| `--addr` | Vault address (or set `VAULT_ADDR`) |

---

## Configuration

`vaultswap` respects standard Vault environment variables:

```bash
export VAULT_ADDR=https://vault.example.com
export VAULT_TOKEN=s.yourtoken
```

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)