#!/usr/bin/env bash
set -euo pipefail

readonly required_commit='795b2ca'

if [[ "$#" -lt 2 || "$#" -gt 3 ]]; then
  printf 'Usage: %s <native|docker> <Kiro-Go-source-directory> [--expect-auth]\n' "$0" >&2
  exit 2
fi

mode="$1"
repo="$2"
expect_auth="${3:-}"

if [[ "$mode" != 'native' && "$mode" != 'docker' ]]; then
  printf 'ERROR: mode must be native or docker.\n' >&2
  exit 2
fi
if [[ -n "$expect_auth" && "$expect_auth" != '--expect-auth' ]]; then
  printf 'ERROR: unknown option: %s\n' "$expect_auth" >&2
  exit 2
fi
if [[ ! -d "$repo/.git" ]]; then
  printf 'ERROR: not a Kiro-Go Git repository: %s\n' "$repo" >&2
  exit 1
fi
repo="$(cd "$repo" && pwd -P)"

printf 'Branch: %s\n' "$(git -C "$repo" branch --show-current)"
git -C "$repo" merge-base --is-ancestor "$required_commit" HEAD
printf 'Required commit %s: OK\n' "$required_commit"

health="$(curl -fsS http://127.0.0.1:8080/health)"
printf 'Health: %s\n' "$health"

if [[ "$mode" == 'native' ]]; then
  launchctl print "gui/$(id -u)/com.vino.kiro-go" >/dev/null
  printf 'LaunchAgent: loaded\n'
  config_host="$(/usr/bin/plutil -extract host raw -o - "$repo/data/config.json" 2>/dev/null || true)"
  if [[ -n "$config_host" && "$config_host" != '127.0.0.1' && "$config_host" != 'localhost' ]]; then
    printf 'ERROR: native config is not loopback-only.\n' >&2
    exit 1
  fi
  printf 'Native bind configuration: loopback-only\n'
else
  running="$(docker inspect -f '{{.State.Running}}' kiro-go-codex 2>/dev/null || true)"
  [[ "$running" == 'true' ]] || { printf 'ERROR: Docker container is not running.\n' >&2; exit 1; }
  port_mapping="$(docker port kiro-go-codex 8080/tcp)"
  if [[ "$port_mapping" != '127.0.0.1:8080' ]]; then
    printf 'ERROR: unexpected Docker port mapping: %s\n' "$port_mapping" >&2
    exit 1
  fi
  printf 'Docker port mapping: %s\n' "$port_mapping"
fi

listener="$(lsof -nP -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)"
[[ -n "$listener" ]] || { printf 'ERROR: no listener found on port 8080.\n' >&2; exit 1; }
printf 'Port 8080 listener: present\n'

if [[ "$expect_auth" == '--expect-auth' ]]; then
  status="$(curl -sS -o /dev/null -w '%{http_code}' \
    http://127.0.0.1:8080/v1/responses \
    -H 'Content-Type: application/json' \
    -d '{"model":"gpt-5.6-sol","input":"test"}')"
  if [[ "$status" != '401' ]]; then
    printf 'ERROR: unauthenticated Responses request returned HTTP %s, expected 401.\n' "$status" >&2
    exit 1
  fi
  printf 'API-key enforcement: OK (unauthenticated request returned 401)\n'
fi

if [[ -d '/Applications/CC Switch.app' || -d "$HOME/Applications/CC Switch.app" ]]; then
  printf 'CC Switch: installed\n'
else
  printf 'CC Switch: not found in Applications\n'
fi

printf 'Terminal verification completed. Model menu, tool calls, and credit counters still require UI/runtime acceptance.\n'
