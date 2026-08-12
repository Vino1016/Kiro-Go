---
name: kiro-go-installer
description: Install, configure, verify, or repair Kiro-Go and CC Switch on macOS so Codex can use Kiro models and credits through the native Responses API. Use when a user asks to install Kiro-Go, configure CC Switch, connect Codex to Kiro, choose Docker versus native macOS deployment, or continue post-restart acceptance. Supports both Codex-assisted GUI automation and Kiro-assisted command-line installation.
---

# Kiro-Go Installer

Install the fixed `Vino1016/Kiro-Go` `main_vino` branch, configure it as a local Responses provider, and preserve the user's official Codex provider and existing configuration.

## Fixed requirements

- Use `https://github.com/Vino1016/Kiro-Go.git`, branch `main_vino`.
- Require commit `795b2ca` to be an ancestor of `HEAD`.
- Bind access to the local machine only.
- Use native Responses at `http://127.0.0.1:8080/v1`; keep CC Switch Local Routing off.
- Never expose, print, store, or commit passwords, SSO credentials, API keys, tokens, or the contents of `data/config.json`.
- Preserve the official `default` Codex provider, authentication, MCP servers, plugins, hooks, and unrelated files.

## 1. Select the assistant mode

Determine the mode from the current harness and available tools. State the detected mode before acting.

- **Codex-assisted**: use when terminal tools plus Browser Use and Computer Use are available. Read `references/codex-assisted.md`.
- **Kiro-assisted**: use when running in Kiro or Browser Use/Computer Use are unavailable. Read `references/kiro-assisted.md`.
- If detection is ambiguous, ask the user to choose Codex-assisted or Kiro-assisted.
- Honor an explicit user choice even when more tools are available.

Do not confuse the UI tools: Kiro-Go admin is a webpage for Browser Use; CC Switch is a native macOS app for Computer Use.

## 2. Select the deployment mode

Before any mutating installation command, ask the user to choose exactly one deployment:

1. **macOS native (recommended)** — Go binary with launchd.
2. **Docker** — locally built container with loopback-only port mapping.

If the user already chose in the invocation, do not ask again. If the user says “recommended”, “default”, or “随便”, select native. Never run both modes on port 8080.

Also confirm the installation directory only when needed. If this Skill is being used directly from a checked-out Kiro-Go repository, prefer that repository root. Otherwise default to `~/IdeaProjects/Kiro-Go`; if it already exists, inspect it and preserve it rather than replacing it.

## 3. Preflight and source preparation

Resolve `SKILL_DIR` to the directory containing this `SKILL.md`, then:

1. Run `scripts/preflight.sh` and report conflicts.
2. Stop if port 8080 is owned by an unrelated process or the other deployment mode. Explain the exact owner and ask before stopping anything.
3. Run `scripts/prepare-source.sh <absolute-target-directory>` after obtaining any required network/write approval.
4. Confirm `main_vino`, commit ancestry, and a clean source state before installation.

Never repoint an unrelated Git remote, reset a dirty worktree, delete `data/`, or overwrite an existing LaunchAgent/container.

## 4. Install the selected deployment

- For native deployment, read `references/native-deployment.md`, then run `scripts/install-native.sh <source-directory>`.
- For Docker deployment, read `references/docker-deployment.md`, then run `scripts/install-docker.sh <source-directory>`.

Treat the scripts as guarded installers. If one refuses because of existing state, inspect that state and ask the user before any replacement or migration.

## 5. Configure Kiro-Go and CC Switch

Read `references/configuration.md` and the selected assistant-mode reference.

Complete these checkpoints in order:

1. Change the default Kiro-Go admin password.
2. Add and test the Kiro account; complete SSO and select the correct enterprise profile when applicable.
3. Refresh models.
4. Create the `codex-local` local gateway API key and enable API-key authentication.
5. Back up `~/.codex/config.toml` and `~/.codex/auth.json` before CC Switch changes.
6. Install CC Switch and preserve the `default` provider and official authentication.
7. Create `Kiro-Go Local`, use native Responses, apply common configuration, disable Local Routing, and add the documented model mappings.
8. Verify MCP, plugin, and hook configuration remains present.

Pause for the user to enter or copy every secret. Do not inspect a screen while a secret is visible. Resume only after the user says the sensitive step is complete.

## 6. Restart and acceptance

Do not quit the application hosting the current task. After all settings are saved, tell the user to quit and restart Codex, start a new task, and invoke this skill with “继续验收”.

On resumed acceptance:

1. Run `scripts/verify.sh <native|docker> <source-directory> --expect-auth`.
2. Confirm `/model` contains the mapped Kiro display names.
3. Run a read-only Shell/file tool task and, when configured, a read-only MCP task.
4. Confirm Kiro-Go request, token, or credit counters increase.
5. Confirm CC Switch can return to `default` without losing official history/configuration.

Do not claim success based only on configuration. Distinguish installed, configured, manually completed, runtime verified, and not yet verified.

## 7. Final report

Report:

- assistant mode and deployment mode;
- source directory, branch, and required-commit result;
- service and health status;
- Kiro-Go account/API authentication status without secret values;
- CC Switch provider/model mapping status;
- Codex tool-call and credit-metering results;
- manual or unverified items;
- backup locations and safe maintenance commands.
