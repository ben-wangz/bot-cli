# proxmox-cli managed assets

This directory stores versioned, reviewable assets used by `proxmox-cli` workflows.

## Stable seed assets

- `single-disk-nocloud/user-data`
- `single-disk-nocloud/meta-data`

These files are the canonical baseline for the single-disk shim NoCloud installer.

## build/ directory policy

`build/` is runtime workspace/output. Keep only environment files and active run artifacts there.

- Keep: `pve-user.env`, `pve-root.env`, source ISO, latest generated shim ISO, active serial logs.
- Do not treat `build/` files as canonical configuration; update assets in this directory first, then copy into `build/` for execution.
