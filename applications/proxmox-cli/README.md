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

Install `proxmox-cli` skill to project-level path via release asset (`.tar.gz`):

```bash
PROJECT_ROOT="/path/to/your/project"
GH_REPO="ben-wangz/bot-cli"
VERSION="0.1.1"
TAG="proxmox-cli-v${VERSION}"
ASSET="proxmox-cli_skills_${VERSION}.tar.gz"

mkdir -p "$PROJECT_ROOT/build" "$PROJECT_ROOT/.opencode/skills"
curl -fsSL -o "$PROJECT_ROOT/build/proxmox-cli-skills.tar.gz" \
  "https://github.com/${GH_REPO}/releases/download/${TAG}/${ASSET}"
rm -rf "$PROJECT_ROOT/.opencode/skills/proxmox-cli"
tar -xzf "$PROJECT_ROOT/build/proxmox-cli-skills.tar.gz" -C "$PROJECT_ROOT/.opencode/skills"
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
- Release assets include platform binaries, `checksums.txt`, and skills bundle `proxmox-cli_skills_<version>.tar.gz`
