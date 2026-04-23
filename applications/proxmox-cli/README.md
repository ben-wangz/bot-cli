# proxmox-cli

Agent-facing CLI for Proxmox-related operations.

## Auth Strategy

- Default execution path uses `pve user` credentials for VM management actions and workflows.
- `root` credentials are not required for normal day-to-day VM lifecycle operations.
- Root scope is reserved for one-time bootstrap tasks (for example, creating/assigning a least-privilege user), then execution should return to user scope.
- Phase 5 planning is limited to root-assisted user/ACL bootstrap (create user, get ACL binding, grant ACL, revoke ACL).
- ACL "change" operations are modeled as `revoke + grant`; no standalone update action is required.
- Historical root-shell action support remains in code for compatibility, but is not part of the recommended main path.

## Code Layout

- Go source root: `applications/proxmox-cli/src`
- Entrypoint: `applications/proxmox-cli/src/cmd/proxmox-cli`
- Internal packages: `applications/proxmox-cli/src/internal`
- Prompt tests: `applications/proxmox-cli/src/test/prompts`
- Versioned workflow assets: `applications/proxmox-cli/assets`

## Versioning

Version is managed via `forgekit` in module mode.

- Module name: `applications/proxmox-cli`
- Version file: `applications/proxmox-cli/container/VERSION`

The `container/` directory here is used as a compatibility path for `forgekit version` conventions. It does not imply this project builds container images for this module.

## Build

Use the module build script:

`/root/code/github/bot-cli/applications/proxmox-cli/build.sh`

Output binary path:

`applications/proxmox-cli/build/bin/proxmox-cli`

Do not build to `applications/proxmox-cli/src/proxmox-cli`.
