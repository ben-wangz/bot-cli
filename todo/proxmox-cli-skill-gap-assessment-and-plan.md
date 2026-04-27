# proxmox-cli Skill Gaps Assessment and Retrofit Plan

Date: 2026-04-27
Input source: `/root/code/geekcity-k8s/pve/todo/proxmox-cli-skill-gaps.md`

## Scope

This note validates each reported gap against current `bot-cli` code/docs and proposes implementation and documentation changes.

## Validation Results

### 1) `get_next_vmid` vs operation range mismatch

Status: **Confirmed**

Evidence:

- `get_next_vmid` directly returns `/cluster/nextid` without policy filtering (`applications/proxmox-cli/src/internal/capability/inventory.go:100`).
- VM mutation paths enforce operation range through `RequiredOperationVMID` / `RequiredOperationInt` (`applications/proxmox-cli/src/internal/capability/vm/request.go:110`, `applications/proxmox-cli/src/internal/capability/vm/request.go:133`).
- Default allowed range is `1001..2000` (`applications/proxmox-cli/src/internal/policy/operation_vmid.go:12`).

Impact:

- Agent can receive a valid cluster next-id that is later rejected by mutation capability with `vmid is outside allowed operation range`.

### 2) No explicit destroy capability in catalog

Status: **Confirmed**

Evidence:

- Capability registry has no VM destroy/delete operation (`applications/proxmox-cli/src/internal/capability/registry.go:47`).
- Skill docs mention cleanup as "destroy/dispose" (`applications/proxmox-cli/skills/proxmox-cli/references/vm-lifecycle.md:13`, `applications/proxmox-cli/skills/proxmox-cli/references/quickstart.md:94`).

Impact:

- Cleanup chain is not fully closed within capability catalog; users may fall back to direct API calls.

### 3) `ssh_inject_pubkey_qga` may report success for non-existent user

Status: **Confirmed**

Evidence:

- Injection script falls back to `/home/$USER_NAME` if `getent` misses user, then continues and ignores `chown` failure (`applications/proxmox-cli/src/internal/capability/ssh/support.go:176`, `applications/proxmox-cli/src/internal/capability/ssh/support.go:184`).
- Bridge returns `ok: true` if QGA command exits 0, without explicit user existence validation (`applications/proxmox-cli/src/internal/capability/ssh_bridge.go:80`).

Impact:

- False-positive "key injected" signal is possible when username does not exist in guest.

### 4) SSH capability auth behavior docs gap

Status: **Confirmed**

Evidence:

- SSH implementation forces batch mode (`BatchMode=yes`) and supports key-based options only (`applications/proxmox-cli/src/internal/capability/ssh/support.go:71`).
- Skill docs list SSH capabilities but do not provide argument-complete examples for `ssh_check_service`/`ssh_exec` nor explicit auth-mode statement (`applications/proxmox-cli/skills/proxmox-cli/references/capability-workflow-catalog.md:55`, `applications/proxmox-cli/skills/proxmox-cli/references/troubleshooting.md:57`).

Impact:

- Users cannot infer required args and auth expectations quickly; password-mode assumptions are likely wrong.

### 5) Capability-level help discoverability

Status: **Confirmed**

Evidence:

- `capabilityHelp()` prints grouped capability names only, no per-capability args (`applications/proxmox-cli/src/internal/cli/help.go:49`).
- `runCapabilityCommand()` treats any `--help` under `capability` as top-level capability help, not capability-specific help (`applications/proxmox-cli/src/internal/cli/app_commands.go:60`).

Impact:

- Discovering required flags depends on trial-and-error and source/doc lookup.

## Retrofit Plan

### A. Range-aware VMID resolution

1. Change existing `get_next_vmid` semantics to be policy-compliant:
   - read cluster next-id,
   - enforce operation range,
   - if cluster next-id is outside range, scan only within allowed range and return first free ID.
2. Add diagnostics fields:
   - `allowed_range`,
   - `cluster_next_vmid`,
   - `selected_vmid`.

Acceptance:

- `get_next_vmid` always returns a usable VMID inside allowed range.
- Returned VMID can be consumed directly by mutation capabilities without range error.

### B. Add explicit VM destroy capability

1. Add `destroy_vm` capability in VM group and catalog.
2. Behavior:
   - required: `--node`, `--vmid`.
   - optional: `--purge 1|0`, `--destroy-unreferenced-disks 1|0`, `--if-missing ok|fail` (default `ok`).
   - pre-check `RequiredOperationVMID` for range safety.
3. API call:
   - `DELETE /nodes/{node}/qemu/{vmid}` (form-encoded options if provided).
4. Idempotency:
   - treat not-found as success when `if-missing=ok`.

Acceptance:

- Disposable VM lifecycle can be completed with capability-only chain.

### C. Tighten `ssh_inject_pubkey_qga` success semantics

1. Add pre-validation in inject script:
   - fail hard if `getent passwd "$USER_NAME"` returns empty (except `root` path logic as applicable).
2. Remove silent ownership fallback (`chown ... || true`) for normal path; convert to explicit error.
3. Return extra diagnostics:
   - `user_exists=true/false`,
   - `home_dir_resolved`.
4. Add optional strict switch (if needed for compatibility): `--allow-missing-user 0|1` default `0`.

Acceptance:

- Injecting to non-existent guest user returns `ok=false` with clear invalid-args/network classification and message.

### D. SSH docs and auth-mode clarity

1. Update skill references with concrete examples for:
   - `ssh_check_service` (key mode),
   - `ssh_exec` (key mode with timeout),
   - `ssh_scp_transfer` (upload/download).
2. Add explicit statement:
   - current SSH capabilities are key-oriented batch SSH (`BatchMode=yes`),
   - password interactive prompts are not supported in capability path.
3. Add "args quick table" in `capability-workflow-catalog.md` for SSH capabilities.

Acceptance:

- A new user can run SSH capabilities using docs alone without trial-and-error.

### E. Per-capability help system

1. Introduce capability metadata schema:
   - `Summary`, `RequiredArgs`, `OptionalArgs`, `Examples`, `Async`, `WaitPolicy`.
2. Extend registry entries to include `HelpMeta`.
3. CLI behavior:
   - `proxmox-cli capability --help` => grouped catalog (existing behavior).
   - `proxmox-cli capability <name> --help` => detailed per-capability help.
4. Add machine-readable option if useful:
   - `proxmox-cli capability describe <name> --output json`.

Acceptance:

- `capability <name> --help` clearly shows required/optional args and executable examples.

## Suggested Delivery Phases

Single phase (recommended):

1. `get_next_vmid` range-compliant behavior.
2. `destroy_vm` capability.
3. `ssh_inject_pubkey_qga` strict user existence.
4. SSH docs examples and auth-mode statement.
5. Per-capability help metadata + CLI rendering.

Execution order inside this phase:

1. Capability semantics fixes (`get_next_vmid`, `destroy_vm`, `ssh_inject_pubkey_qga`).
2. CLI help metadata and rendering.
3. Skill/docs synchronization.

## Risks and Compatibility Notes

- Changing `get_next_vmid` behavior may affect external callers that expect raw `/cluster/nextid`; inside this repo, current prompts already avoid relying on raw next-id for mutations.
- Tightening `ssh_inject_pubkey_qga` may break flows that relied on implicit home-dir fallback for absent users; consider feature flag during transition.
- `destroy_vm` is destructive; keep operation-range enforcement mandatory and document scope expectations.
- Per-capability help introduces metadata maintenance overhead; reduce drift by colocating metadata with registry entries.

## Test Plan (High Level)

1. Prompt/e2e updates (primary focus):
   - lifecycle cleanup uses `destroy_vm` instead of raw API.
   - `get_next_vmid` is consumed directly in a mutation path to verify no range mismatch.
   - add a negative check for `ssh_inject_pubkey_qga` with non-existent user (must fail).
   - add one doc-driven SSH capability smoke path.
2. CLI help snapshots:
   - verify `capability <name> --help` output for 2-3 representative capabilities.
   - verify `capability describe <name>` JSON shape for machine-readability.

Note:

- Unit tests are optional for this change set; prioritize prompt/e2e and CLI help validation.
