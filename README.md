# bot-cli

Repository for agent-first command line tools.

## Scope

- Hosts multiple CLIs under `applications/`.
- Keeps shared engineering standards (versioning, linting, release automation).
- Keeps per-CLI runtime details in each application's own docs.

## Repository Layout

- `applications/<app>/`: each CLI implementation, tests, release notes, and skill docs
- `setup/`: bootstrap utilities used by repository workflows
- `.github/workflows/`: CI and release pipelines
- `version-control.yaml`: binary-to-version-file mapping
- `lint.yaml`: repository lint commands executed by forgekit

## Working Model

1. Treat each CLI as self-contained.
2. Put tool-specific usage and operational playbooks in `applications/<app>/README.md` and skill docs.
3. Keep root README focused on repository-level conventions only.

## Version and Lint (Repository Level)

This repo uses `forgekit` for version and lint orchestration.

```bash
PROJECT_ROOT="$(pwd)"
FORGEKIT_BIN=$(bash "${PROJECT_ROOT}/setup/forgekit.sh")

"${FORGEKIT_BIN}" --project-root "${PROJECT_ROOT}" lint
"${FORGEKIT_BIN}" --project-root "${PROJECT_ROOT}" version get <app-name>
"${FORGEKIT_BIN}" --project-root "${PROJECT_ROOT}" version bump <app-name> patch
```

## Release Convention

- Tag format: `<application>-v<semver>`
- Release workflow validates app path and mapped version file before publishing artifacts.

For app-specific release readiness (tests, prompts, rollout notes), use that app's docs under `applications/<app>/`.

## OpenCode Integration

- Repository instructions: `opencode.json` -> `.opencode/AGENTS.md`
- Skill locations follow OpenCode conventions (project-level or user-level paths).
- App-specific skills should live with the app source and be installable into OpenCode skill directories.
