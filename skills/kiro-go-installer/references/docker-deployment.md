# Docker deployment

Use this mode when the user explicitly chooses container isolation.

## Preconditions

- Docker Desktop running.
- Port 8080 unused.
- Native launchd Kiro-Go not running.
- Source prepared on `main_vino` with commit `795b2ca`.

## Installer behavior

Run:

```bash
"$SKILL_DIR/scripts/install-docker.sh" "/absolute/path/to/Kiro-Go"
```

The installer:

1. builds `kiro-go:main_vino` from local source;
2. creates an initial container config only when none exists;
3. starts `kiro-go-codex` with `127.0.0.1:8080:8080`;
4. mounts the source `data/` directory at `/app/data`;
5. sets `--restart unless-stopped` and verifies `/health`.

Do not use the repository's unmodified `docker-compose.yml` for this local-only installation because its `8080:8080` mapping is not loopback-restricted. The installer refuses to replace an existing container or take over an occupied port.

## Maintenance

Keep `data/` when replacing a container. Never remove it unless the user explicitly requests credential/configuration deletion and understands recovery implications.
