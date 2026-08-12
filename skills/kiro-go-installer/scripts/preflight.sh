#!/usr/bin/env bash
set -u

missing=0

printf 'Kiro-Go installer preflight\n'
printf 'OS: %s\n' "$(uname -s)"
printf 'Architecture: %s\n' "$(uname -m)"

if [[ "$(uname -s)" != "Darwin" ]]; then
  printf 'ERROR: this installer supports macOS only.\n' >&2
  exit 1
fi

for command_name in git curl brew; do
  if command -v "$command_name" >/dev/null 2>&1; then
    printf '%-12s %s\n' "$command_name:" "$(command -v "$command_name")"
  else
    printf '%-12s %s\n' "$command_name:" 'MISSING'
    missing=1
  fi
done

for command_name in go docker; do
  if command -v "$command_name" >/dev/null 2>&1; then
    printf '%-12s %s\n' "$command_name:" "$(command -v "$command_name")"
  else
    printf '%-12s %s\n' "$command_name:" 'not installed (required only by the selected deployment)'
  fi
done

if command -v lsof >/dev/null 2>&1; then
  port_output="$(lsof -nP -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)"
  if [[ -n "$port_output" ]]; then
    printf '\nPort 8080 is occupied:\n%s\n' "$port_output"
  else
    printf '\nPort 8080: available\n'
  fi
else
  printf '\nPort 8080: not checked because lsof is unavailable\n'
fi

if launchctl print "gui/$(id -u)/com.vino.kiro-go" >/dev/null 2>&1; then
  printf 'LaunchAgent com.vino.kiro-go: loaded\n'
else
  printf 'LaunchAgent com.vino.kiro-go: not loaded\n'
fi

if command -v docker >/dev/null 2>&1 && docker container inspect kiro-go-codex >/dev/null 2>&1; then
  printf 'Docker container kiro-go-codex: exists\n'
else
  printf 'Docker container kiro-go-codex: not found\n'
fi

if [[ -d '/Applications/CC Switch.app' || -d "$HOME/Applications/CC Switch.app" ]]; then
  printf 'CC Switch: installed\n'
else
  printf 'CC Switch: not found in Applications\n'
fi

if [[ "$missing" -ne 0 ]]; then
  printf '\nERROR: install the missing required commands before continuing.\n' >&2
  exit 1
fi

printf '\nPreflight completed. Resolve any existing service or port conflict before installation.\n'
