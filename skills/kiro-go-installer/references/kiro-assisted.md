# Kiro-assisted mode

Use Kiro for terminal installation and deterministic verification. Treat graphical configuration as a guided human workflow unless the current Kiro environment demonstrably exposes an appropriate UI tool.

## Kiro can complete

- environment and conflict checks;
- cloning/updating the required source branch;
- Go tests and native builds;
- Docker image build and container start;
- launchd installation;
- health, process, port, branch, and commit verification;
- Homebrew installation of CC Switch.

## User must complete

- Kiro-Go admin login and password change;
- Kiro account type selection, browser SSO, authorization, and enterprise profile selection;
- account test and model refresh when Kiro cannot operate the page;
- creation and secure storage of the `codex-local` gateway key;
- CC Switch provider, common configuration, API key, and model mappings;
- Codex restart, model selection, tool-call acceptance, and credit-counter confirmation.

Give one UI step at a time and wait for “已完成”. After each response, verify what can be checked from the terminal before continuing. Never infer that a graphical step succeeded.
