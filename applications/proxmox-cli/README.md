# proxmox-cli

Agent-facing CLI for Proxmox-related operations.

## Auth Strategy

- Default execution path uses `pve user` credentials for VM management actions and workflows.
- `root` credentials are not required for normal day-to-day VM lifecycle operations.
- Root scope is reserved for one-time bootstrap tasks (for example, creating/assigning a least-privilege user), then execution should return to user scope.
- Phase 5 planning is limited to root-assisted pool/user ACL bootstrap (create pool/user, get ACL binding, grant ACL, revoke ACL).
- ACL "change" operations are modeled as `revoke + grant`; no standalone update action is required.
- Historical root-shell action path is removed in v0.1 M4 cleanup; keep using Phase 4 serial/websocket and SSH/QGA actions on the main path.

## Code Layout

- Go source root: `applications/proxmox-cli/src`
- Entrypoint: `applications/proxmox-cli/src/cmd/proxmox-cli`
- Internal packages: `applications/proxmox-cli/src/internal`
- Prompt tests: `applications/proxmox-cli/tests/prompts`
- Versioned workflow assets: `applications/proxmox-cli/assets`

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

Use the module build script:

`/root/code/github/bot-cli/applications/proxmox-cli/build.sh`

Output binary path:

`applications/proxmox-cli/build/bin/proxmox-cli`

Do not build to `applications/proxmox-cli/src/proxmox-cli`.
