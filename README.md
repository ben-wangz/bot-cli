# bot-cli

Repository for building **agent-native CLIs**.

This repository is used to design and maintain multiple small CLI tools that are built for AI agents first, while still being usable by humans.

## What Is an Agent CLI?

An agent CLI is a command-line interface designed for reliable machine use:

- **Structured output**: supports deterministic JSON output for programmatic parsing.
- **Discoverable commands**: clear `--help`, stable command groups, predictable flags.
- **Composable workflows**: works in pipelines, scripts, and autonomous agent loops.
- **Real backend integration**: wraps real software/APIs/services, not toy mocks.

## Project Goals

- Build practical CLI tools for daily workflows.
- Standardize command design so agents can switch tools with low friction.
- Keep each CLI small, testable, and easy to iterate.
- Reuse common patterns (errors, output schema, auth, retries, pagination).

## Directory Layout

All agent CLIs are placed under:

- `applications/<app-name>/`

Current app scaffold:

- `applications/proxmox-cli/`

## Version Management (forgekit)

This project uses `forgekit` for version management.

### Binary mapping

`forgekit version` can manage standalone binaries through `version-control.yaml`.

For `proxmox-cli`:

- binary name: `proxmox-cli`
- binary path: `applications/proxmox-cli`
- version file: `applications/proxmox-cli/VERSION`

### Bootstrap forgekit binary

```bash
REPO_ROOT="$(pwd)"
FORGEKIT_BIN=$(bash "${REPO_ROOT}/setup/forgekit.sh")
"${FORGEKIT_BIN}" --project-root "${REPO_ROOT}" version get proxmox-cli
```

### Common commands

```bash
REPO_ROOT="$(pwd)"

# Get current semver
forgekit --project-root "${REPO_ROOT}" version get proxmox-cli

# Bump version
forgekit --project-root "${REPO_ROOT}" version bump proxmox-cli patch
forgekit --project-root "${REPO_ROOT}" version bump proxmox-cli minor
forgekit --project-root "${REPO_ROOT}" version bump proxmox-cli major
```

## Core Conventions (Planned Baseline)

Each CLI in this repo should follow these conventions unless explicitly overridden:

1. **Command shape**
   - Use verb-first subcommands where possible.
   - Keep naming consistent across tools (`list`, `get`, `create`, `update`, `delete`, `run`, `check`).

2. **Output modes**
   - Human-readable output by default.
   - `--json` for machine-readable output.
   - Exit code `0` for success, non-zero for failure.

3. **Error contract**
   - Clear, short error messages.
   - When `--json` is enabled, return structured error payloads.

4. **Safety defaults**
   - Read-only by default for dangerous targets.
   - Require explicit flags/arguments for destructive operations.

5. **Docs and discoverability**
   - Every CLI must provide meaningful `--help` output.
   - Every CLI should ship or reference a `SKILL.md` for agent usage patterns.

## Suggested CLI Development Checklist

For each new agent CLI:

1. Define scope (what backend/system it controls).
2. Define command groups and minimal JSON schema.
3. Implement MVP commands.
4. Add unit tests for parser + core logic.
5. Add at least one end-to-end command test.
6. Add/update usage docs and agent skill notes.

## OpenCode Integration

This repo is configured to load local rules via:

- `opencode.json` -> `instructions: [".opencode/AGENTS.md"]`

Local project skills can be installed under:

- `.opencode/skills/<skill-name>/SKILL.md`

## Notes

- Priorities are reliability, speed, and practical usefulness over framework complexity.
- The repository will gradually expand as more agent-focused CLIs are added.
