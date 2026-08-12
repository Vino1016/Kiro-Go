#!/usr/bin/env bash
set -euo pipefail

readonly repo_url='https://github.com/Vino1016/Kiro-Go.git'
readonly required_commit='795b2ca'

if [[ "$#" -ne 1 ]]; then
  printf 'Usage: %s <target-directory>\n' "$0" >&2
  exit 2
fi

target="$1"
case "$target" in
  '~/'*) target="$HOME/${target#\~/}" ;;
esac
if [[ "$target" != /* ]]; then
  target="$(pwd)/$target"
fi

if [[ -e "$target" && ! -d "$target" ]]; then
  printf 'ERROR: target exists and is not a directory: %s\n' "$target" >&2
  exit 1
fi

if [[ -d "$target/.git" ]]; then
  origin_url="$(git -C "$target" remote get-url origin 2>/dev/null || true)"
  case "$origin_url" in
    'https://github.com/Vino1016/Kiro-Go.git'|'https://github.com/Vino1016/Kiro-Go'|'git@github.com:Vino1016/Kiro-Go.git') ;;
    *)
      printf 'ERROR: existing repository has an unexpected origin: %s\n' "$origin_url" >&2
      printf 'Refusing to repoint or overwrite it.\n' >&2
      exit 1
      ;;
  esac

  if [[ -n "$(git -C "$target" status --porcelain)" ]]; then
    printf 'ERROR: existing Kiro-Go worktree is dirty: %s\n' "$target" >&2
    printf 'Preserve or commit those changes before updating.\n' >&2
    exit 1
  fi

  git -C "$target" fetch origin '+refs/heads/main_vino:refs/remotes/origin/main_vino'
  if git -C "$target" show-ref --verify --quiet refs/heads/main_vino; then
    git -C "$target" switch main_vino
  else
    git -C "$target" switch -c main_vino --track origin/main_vino
  fi
  git -C "$target" pull --ff-only origin main_vino
elif [[ -d "$target" && -n "$(ls -A "$target")" ]]; then
  printf 'ERROR: target directory is non-empty and is not a Git repository: %s\n' "$target" >&2
  exit 1
else
  mkdir -p "$(dirname "$target")"
  git clone --branch main_vino --single-branch "$repo_url" "$target"
fi

if [[ "$(git -C "$target" branch --show-current)" != 'main_vino' ]]; then
  printf 'ERROR: expected main_vino branch.\n' >&2
  exit 1
fi

if ! git -C "$target" merge-base --is-ancestor "$required_commit" HEAD; then
  printf 'ERROR: required fix commit %s is not present.\n' "$required_commit" >&2
  exit 1
fi

printf 'Source ready\n'
printf 'Directory: %s\n' "$target"
printf 'Branch: %s\n' "$(git -C "$target" branch --show-current)"
printf 'Commit: %s\n' "$(git -C "$target" log -1 --oneline)"
