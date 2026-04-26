# proxmox-cli v0.1 Release Checklist

Use this checklist before creating a release tag.

## Scope

- Application: `proxmox-cli`
- Tag format: `proxmox-cli-v<semver>`
- Gate policy: `applications/proxmox-cli/tests/prompts/RELEASE-REGRESSION-GATE.md`

## 1) Version Check

- [x] `applications/proxmox-cli/VERSION` contains the target semver (`0.1.0`).
- [x] `version-control.yaml` still declares binary `proxmox-cli`.
- [x] Local version command returns expected value:

```bash
REPO_ROOT="$(pwd)"
FORGEKIT_BIN=$(bash "${REPO_ROOT}/setup/forgekit.sh")
"${FORGEKIT_BIN}" --project-root "${REPO_ROOT}" version get proxmox-cli
```

## 2) Regression Gate Check

- [x] Full regression sequence is green in this order:
  1. `applications/proxmox-cli/tests/prompts/setup.md`
  2. `applications/proxmox-cli/tests/prompts/workflows/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md`
  3. `applications/proxmox-cli/tests/prompts/workflows/PROVISION-TEMPLATE-FROM-ARTIFACT.prompt.md`
  4. `applications/proxmox-cli/tests/prompts/composed-v0.1/RUN-COMPOSED-SUITE.prompt.md`
- [x] Required artifacts exist after run:
  - `build/pve-user.env`
  - `build/ubuntu-24-with-agent.vm-template.id`
- [x] No unresolved failures remain in referenced logs.

## 3) Change Notes Check

- [x] Release notes summarize what changed and why (not only file list).
- [x] Notes include impact scope (user-facing behavior, tests/regression, compatibility).
- [x] Notes exclude unrelated work.
- [x] Release notes file exists: `applications/proxmox-cli/release/RELEASE-NOTES-v0.1.md`.

## 4) Tag and Release Check (Skipped in this cycle)

- [x] This v0.1 closeout does not perform actual tag push or GitHub Release publishing.
- [ ] If/when publishing later, create and push release tag with app prefix:

```bash
git tag proxmox-cli-v<semver>
git push origin proxmox-cli-v<semver>
```

- [ ] If/when publishing later, GitHub `Release` workflow succeeds for this tag.
- [ ] If/when publishing later, release artifacts are uploaded (`linux/darwin` for `amd64/arm64` + `checksums.txt`).

## 5) Final Decision

- [x] For current v0.1 M5 closeout, release readiness documentation is accepted without executing tag/release publishing.
- [x] For any later public release, Section 4 must be completed before tag/release is allowed.
