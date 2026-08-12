#!/usr/bin/env bash
set -euo pipefail

readonly required_commit='795b2ca'
readonly container_name='kiro-go-codex'
readonly image_name='kiro-go:main_vino'

if [[ "$#" -ne 1 ]]; then
  printf 'Usage: %s <Kiro-Go-source-directory>\n' "$0" >&2
  exit 2
fi

repo="$1"
if [[ ! -d "$repo/.git" ]]; then
  printf 'ERROR: not a Kiro-Go Git repository: %s\n' "$repo" >&2
  exit 1
fi
repo="$(cd "$repo" && pwd -P)"

command -v docker >/dev/null 2>&1 || { printf 'ERROR: Docker is not installed.\n' >&2; exit 1; }
docker info >/dev/null 2>&1 || { printf 'ERROR: Docker Desktop is not running.\n' >&2; exit 1; }

if [[ "$(git -C "$repo" branch --show-current)" != 'main_vino' ]] || ! git -C "$repo" merge-base --is-ancestor "$required_commit" HEAD; then
  printf 'ERROR: source must be main_vino and contain %s.\n' "$required_commit" >&2
  exit 1
fi
if [[ -n "$(lsof -nP -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)" ]]; then
  printf 'ERROR: port 8080 is already occupied. Resolve it before installation.\n' >&2
  lsof -nP -iTCP:8080 -sTCP:LISTEN >&2 || true
  exit 1
fi
if docker container inspect "$container_name" >/dev/null 2>&1; then
  printf 'ERROR: container %s already exists. Refusing to replace it.\n' "$container_name" >&2
  exit 1
fi
if launchctl print "gui/$(id -u)/com.vino.kiro-go" >/dev/null 2>&1; then
  printf 'ERROR: native Kiro-Go LaunchAgent is loaded. Use only one deployment mode.\n' >&2
  exit 1
fi

mkdir -p "$repo/data"
config="$repo/data/config.json"
if [[ -e "$config" ]]; then
  /usr/bin/plutil -lint "$config" >/dev/null || { printf 'ERROR: existing config is not valid JSON.\n' >&2; exit 1; }
  config_host="$(/usr/bin/plutil -extract host raw -o - "$config" 2>/dev/null || true)"
  if [[ "$config_host" != '0.0.0.0' ]]; then
    printf 'ERROR: existing config host is not suitable for the container. Review before changing it.\n' >&2
    exit 1
  fi
fi

docker build -t "$image_name" "$repo"
if [[ ! -e "$config" ]]; then
  umask 077
  printf '%s\n' '{' \
    '  "password": "changeme",' \
    '  "port": 8080,' \
    '  "host": "0.0.0.0",' \
    '  "requireApiKey": false,' \
    '  "accounts": []' \
    '}' > "$config"
  chmod 600 "$config"
fi

docker run -d \
  --name "$container_name" \
  -p 127.0.0.1:8080:8080 \
  -e CONFIG_PATH=/app/data/config.json \
  -v "$repo/data:/app/data" \
  --restart unless-stopped \
  "$image_name" >/dev/null

healthy=0
for _ in {1..30}; do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null 2>&1; then
    healthy=1
    break
  fi
  sleep 0.5
done

if [[ "$healthy" -ne 1 ]]; then
  printf 'ERROR: container did not become healthy. Removing the newly created container.\n' >&2
  docker logs --tail 50 "$container_name" >&2 || true
  docker stop "$container_name" >/dev/null 2>&1 || true
  docker rm "$container_name" >/dev/null 2>&1 || true
  exit 1
fi

printf 'Docker Kiro-Go installed successfully.\n'
printf 'Admin: http://127.0.0.1:8080/admin\n'
printf 'Next: change the default admin password before adding an account.\n'
