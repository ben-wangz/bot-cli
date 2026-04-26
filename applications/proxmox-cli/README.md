# proxmox-cli

Agent-facing CLI for Proxmox-related operations.

## Versioning

Version is managed via `forgekit` binary mapping.

- Binary name: `proxmox-cli`
- Version file: `applications/proxmox-cli/VERSION`
- Mapping file: `version-control.yaml`

Common commands:

- `FORGEKIT_BIN=$(bash "$PROJECT_ROOT/setup/forgekit.sh")`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version get proxmox-cli`
- `${FORGEKIT_BIN} --project-root "$PROJECT_ROOT" version bump proxmox-cli patch`

## OpenCode Skill Install

OpenCode docs confirm skill discovery locations include:

- Project-level (default): `$PROJECT_ROOT/.opencode/skills/<name>/SKILL.md`
- User-level (global): `~/.config/opencode/skills/<name>/SKILL.md`

Reference: `https://opencode.ai/docs/skills` (`Place files` section).

Install `proxmox-cli` skill to project-level path via `git clone`:

```bash
PROJECT_ROOT="/path/to/your/project"
REPO_URL="https://github.com/ben-wangz/bot-cli.git"
TMP_DIR="$(mktemp -d)"

git clone --depth 1 --filter=blob:none --sparse "$REPO_URL" "$TMP_DIR"
git -C "$TMP_DIR" sparse-checkout set applications/proxmox-cli/skills/proxmox-cli

mkdir -p "$PROJECT_ROOT/.opencode/skills"
rm -rf "$PROJECT_ROOT/.opencode/skills/proxmox-cli"
cp -R "$TMP_DIR/applications/proxmox-cli/skills/proxmox-cli" "$PROJECT_ROOT/.opencode/skills/proxmox-cli"

rm -rf "$TMP_DIR"
```

Install to user-level path (shared across projects):

```bash
mkdir -p "$HOME/.config/opencode/skills"
rm -rf "$HOME/.config/opencode/skills/proxmox-cli"
cp -R "$PROJECT_ROOT/.opencode/skills/proxmox-cli" "$HOME/.config/opencode/skills/proxmox-cli"
```

## Release

- Tag format: `proxmox-cli-v<semver>`
- Release notes directory: `applications/proxmox-cli/release/`
