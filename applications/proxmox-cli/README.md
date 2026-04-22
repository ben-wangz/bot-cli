# proxmox-cli

Agent-facing CLI for Proxmox-related operations.

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
