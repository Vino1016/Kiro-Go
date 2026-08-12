# Native macOS deployment

Use this mode for the recommended lightweight installation.

## Preconditions

- macOS 12 or later.
- Homebrew and Git available.
- Port 8080 unused.
- No Docker Kiro-Go container running.
- Source prepared on `main_vino` with commit `795b2ca`.

## Installer behavior

Run:

```bash
"$SKILL_DIR/scripts/install-native.sh" "/absolute/path/to/Kiro-Go"
```

The installer:

1. installs Go with Homebrew when absent;
2. runs `go test ./...`;
3. builds the `kiro-go` binary;
4. creates a loopback-only initial config only when no config exists;
5. installs `com.vino.kiro-go` from the repository launchd template;
6. starts the LaunchAgent and verifies `/health`.

It refuses to replace an existing LaunchAgent, modify an existing non-loopback config, or take over an occupied port. Inspect and ask before resolving those cases.

## Maintenance

After an approved source update:

```bash
go test ./...
go build -o kiro-go .
launchctl kickstart -k "gui/$(id -u)/com.vino.kiro-go"
curl -fsS http://127.0.0.1:8080/health
```

Logs are `logs/kiro-go.log` and `logs/kiro-go.error.log` under the source directory.
