# Codex-assisted mode

Use the highest safe automation available while keeping secrets under user control.

## Tool routing

| Work | Tool |
| --- | --- |
| Environment, Git, build, test, Docker, launchd, backups, HTTP checks | Terminal |
| Kiro-Go admin at `http://127.0.0.1:8080/admin` and browser SSO | Browser Use |
| CC Switch native macOS application | Computer Use |

Do not use Computer Use for the Kiro-Go webpage. Do not use Browser Use for CC Switch.

## Execution rules

1. Perform terminal work directly instead of asking the user to copy commands.
2. Use Browser Use to navigate Kiro-Go pages, select non-sensitive options, test accounts, refresh models, create the named key entry, and enable API authentication.
3. Hand control to the user before password entry, SSO authorization, profile selection when identity-sensitive, or viewing/copying a generated key.
4. Use Computer Use to configure CC Switch, but pause at the API-key field for the user to paste the key.
5. When Browser Use is unavailable, guide the Kiro-Go webpage one field at a time. When Computer Use is unavailable, guide CC Switch one field at a time.
6. Never claim a click or saved setting succeeded unless the resulting state is visible or independently verified.

## Restart boundary

Because quitting Codex ends the active task, save all configuration first. Tell the user to restart Codex and invoke `$kiro-go-installer 继续验收`; verify persisted state in the new task.
