#!/usr/bin/env bash
set -euo pipefail

readonly required_commit='795b2ca'
readonly label='com.vino.kiro-go'

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

if [[ "$(uname -s)" != 'Darwin' ]]; then
  printf 'ERROR: native deployment requires macOS.\n' >&2
  exit 1
fi
if [[ "$(git -C "$repo" branch --show-current)" != 'main_vino' ]] || ! git -C "$repo" merge-base --is-ancestor "$required_commit" HEAD; then
  printf 'ERROR: source must be main_vino and contain %s.\n' "$required_commit" >&2
  exit 1
fi
if [[ -n "$(lsof -nP -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)" ]]; then
  printf 'ERROR: port 8080 is already occupied. Resolve it before installation.\n' >&2
  lsof -nP -iTCP:8080 -sTCP:LISTEN >&2 || true
  exit 1
fi

plist="$HOME/Library/LaunchAgents/$label.plist"
if [[ -e "$plist" ]] || launchctl print "gui/$(id -u)/$label" >/dev/null 2>&1; then
  printf 'ERROR: an existing %s LaunchAgent was found. Refusing to replace it.\n' "$label" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  command -v brew >/dev/null 2>&1 || { printf 'ERROR: Homebrew is required to install Go.\n' >&2; exit 1; }
  brew install go
fi

mkdir -p "$repo/data" "$repo/logs" "$HOME/Library/LaunchAgents"
config="$repo/data/config.json"
if [[ -e "$config" ]]; then
  /usr/bin/plutil -lint "$config" >/dev/null || { printf 'ERROR: existing config is not valid JSON.\n' >&2; exit 1; }
  config_host="$(/usr/bin/plutil -extract host raw -o - "$config" 2>/dev/null || true)"
  if [[ -n "$config_host" && "$config_host" != '127.0.0.1' && "$config_host" != 'localhost' ]]; then
    printf 'ERROR: existing config is not loopback-only. Change only its host setting after review.\n' >&2
    exit 1
  fi
fi

template="$repo/deploy/macos/com.vino.kiro-go.plist.template"
if [[ ! -f "$template" ]]; then
  printf 'ERROR: launchd template is missing: %s\n' "$template" >&2
  exit 1
fi

(
  cd "$repo"
  go test ./...
  go build -o kiro-go.new .
  mv kiro-go.new kiro-go
)

if [[ ! -e "$config" ]]; then
  umask 077
  printf '%s\n' '{' \
    '  "password": "changeme",' \
    '  "port": 8080,' \
    '  "host": "127.0.0.1",' \
    '  "requireApiKey": false,' \
    '  "accounts": []' \
    '}' > "$config"
  chmod 600 "$config"
fi

temp_plist="$(mktemp "${TMPDIR:-/tmp}/kiro-go-launchd.XXXXXX")"
cleanup_temp() {
  rm -f "$temp_plist"
}
trap cleanup_temp EXIT

while IFS= read -r line || [[ -n "$line" ]]; do
  printf '%s\n' "${line//__KIRO_GO_DIR__/$repo}"
done < "$template" > "$temp_plist"

/usr/bin/plutil -lint "$temp_plist"
mv "$temp_plist" "$plist"
chmod 600 "$plist"
trap - EXIT

launchctl bootstrap "gui/$(id -u)" "$plist"
launchctl enable "gui/$(id -u)/$label"
launchctl kickstart -k "gui/$(id -u)/$label"

healthy=0
for _ in {1..20}; do
  if curl -fsS http://127.0.0.1:8080/health >/dev/null 2>&1; then
    healthy=1
    break
  fi
  sleep 0.5
done

if [[ "$healthy" -ne 1 ]]; then
  printf 'ERROR: Kiro-Go did not become healthy. Cleaning up the newly installed LaunchAgent.\n' >&2
  launchctl bootout "gui/$(id -u)/$label" >/dev/null 2>&1 || true
  rm -f "$plist"
  tail -n 30 "$repo/logs/kiro-go.error.log" >&2 2>/dev/null || true
  exit 1
fi

printf 'Native Kiro-Go installed successfully.\n'
printf 'Admin: http://127.0.0.1:8080/admin\n'
printf 'Next: change the default admin password before adding an account.\n'
