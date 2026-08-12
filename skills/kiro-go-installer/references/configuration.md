# Kiro-Go and CC Switch configuration

## Kiro-Go admin

Open `http://127.0.0.1:8080/admin`.

1. Sign in with the initial password and immediately set a new password.
2. Add the user's authorized Kiro account using the correct authentication type.
3. For enterprise SSO, let the user complete browser authorization and select the correct profile.
4. Confirm the account is normal, enabled, and passes its test.
5. Refresh models.
6. In API management, create an enabled gateway key named `codex-local` and enable API-key validation.

The gateway key is not the Kiro account's `ksk_...` credential. Never read or expose either value.

## CC Switch installation and preservation

Back up existing Codex files before changes:

```bash
BACKUP_DIR="$HOME/.codex/backups/$(date +%Y%m%d-%H%M%S)-pre-kiro-go"
mkdir -p "$BACKUP_DIR"
test ! -f "$HOME/.codex/config.toml" || cp "$HOME/.codex/config.toml" "$BACKUP_DIR/config.toml"
test ! -f "$HOME/.codex/auth.json" || cp "$HOME/.codex/auth.json" "$BACKUP_DIR/auth.json"
chmod 700 "$BACKUP_DIR"
chmod 600 "$BACKUP_DIR"/* 2>/dev/null || true
```

Install or update:

```bash
brew install --cask cc-switch
```

Preserve the imported `default` provider and official Codex authentication. Enable preservation of official authentication, optionally enable unified history, apply common configuration, and keep Local Routing off.

## Kiro-Go Local provider

| Field | Value |
| --- | --- |
| Name | `Kiro-Go Local` |
| Base URL | `http://127.0.0.1:8080/v1` |
| API key | User pastes the Kiro-Go `codex-local` gateway key |
| API format | `Responses (native)` / `Responses（原生）` |
| Default actual model | `gpt-5.6-sol` |
| Apply common configuration | On |
| Local Routing | Off |

## Model mappings

| Display name | Actual request model | Context window |
| --- | --- | ---: |
| `Claude Opus 5` | `claude-opus-5` | 1,000,000 |
| `Claude Sonnet 5` | `claude-sonnet-5` | 1,000,000 |
| `Claude Opus 4.8` | `claude-opus-4.8` | 1,000,000 |
| `gpt-5.6-sol-kiro` | `gpt-5.6-sol` | 200,000 |
| `gpt-5.6-terra-kiro` | `gpt-5.6-terra` | 200,000 |

Treat refreshed Kiro model availability and limits as authoritative. The `-kiro` suffix is only a display distinction; every model under this provider goes to Kiro-Go and consumes Kiro credits.

## Configuration checks

After saving, verify the user-level Codex configuration still contains expected MCP, plugin, and hook sections. Do not print secrets while comparing. Nested per-tool approval blocks may need manual restoration from the backup.

After a full Codex restart, use `/model`, perform a read-only tool call, and confirm the Kiro-Go dashboard counters increase. Switch back to `default` to verify official Codex remains recoverable.
