# proxmox-cli

Agent-facing CLI for Proxmox-related operations.

## Auth Strategy

- Default execution path uses `pve user` credentials for VM management capabilities and workflows.
- `root` credentials are not required for normal day-to-day VM lifecycle operations.
- Root scope is reserved for one-time bootstrap tasks (for example, creating/assigning a least-privilege user), then execution should return to user scope.
- Root-scope planning is limited to root-assisted pool/user ACL bootstrap (create pool/user, get ACL binding, grant ACL, revoke ACL).
- ACL "change" operations are modeled as `revoke + grant`; no standalone capability is required.
- Historical root-shell path is removed in v0.1 cleanup; keep using serial/websocket and SSH/QGA capabilities on the main path.

## Code Layout

- Go source root: `applications/proxmox-cli/src`
- Entrypoint: `applications/proxmox-cli/src/cmd/proxmox-cli`
- Internal packages: `applications/proxmox-cli/src/internal`
- Prompt tests: `applications/proxmox-cli/tests/prompts`
- Versioned workflow assets: `applications/proxmox-cli/assets`

### Internal Package Guide

- `internal/cli`: CLI entrypoints, global flags, command dispatch, and rendering wiring.
- `internal/capability`: capability handlers grouped by stable domains (`vm`, `console`, `ssh`, `storage`, `guest`, `access`).
- `internal/workflow`: workflow orchestration and shared step runner helpers.
- `internal/taskwait`: reusable poll/wait primitives for task and session waits.
- `internal/pveapi`: Proxmox HTTP/websocket client adapters.
- `internal/auth`, `internal/output`, `internal/policy`, `internal/redact`: cross-cutting support modules.

## Versioning

Version is managed via `forgekit` binary mapping.

- Binary name: `proxmox-cli`
- Version file: `applications/proxmox-cli/VERSION`
- Mapping file: `version-control.yaml`

Common commands:

- `FORGEKIT_BIN=$(bash /root/code/github/bot-cli/setup/forgekit.sh)`
- `${FORGEKIT_BIN} --project-root /root/code/github/bot-cli version get proxmox-cli`
- `${FORGEKIT_BIN} --project-root /root/code/github/bot-cli version bump proxmox-cli patch`

## Build

Build from source root:

`cd /root/code/github/bot-cli/applications/proxmox-cli/src && go build -o ../build/bin/proxmox-cli ./cmd/proxmox-cli`

Output binary path:

`applications/proxmox-cli/build/bin/proxmox-cli`

Do not build to `applications/proxmox-cli/src/proxmox-cli`.

## Release

- Tag format: `proxmox-cli-v<semver>`
