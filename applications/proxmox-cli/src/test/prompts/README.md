# proxmox-cli Prompt Suites

This directory is the single entry for full regression coverage of `proxmox-cli` actions and workflows.

- Covers Phase 1-5 action prompts, composed virtual-workflow prompts, and workflow-level e2e prompts.
- Includes suite runners (`RUN-PHASE-*.prompt.md`) and standalone workflow prompts under `e2e/`.
- Prompts are designed to run with minimal manual edits.

## Full Regression Entry

Run prompts in this order:

1. `setup.md`
2. `phase-1-read-task/RUN-PHASE-1-SUITE.prompt.md`
3. `phase-2-vm-lifecycle/RUN-PHASE-2-SUITE.prompt.md`
4. `phase-3-cloudinit-qga/RUN-PHASE-3-SUITE.prompt.md`
5. `phase-4-console-websocket/RUN-PHASE-4-SUITE.prompt.md`
6. `phase-5-privilege-root/RUN-PHASE-5-SUITE.prompt.md`
7. `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md`
8. `e2e/PROVISION-TEMPLATE-FROM-ARTIFACT.prompt.md`
9. `composed-v0.1/RUN-COMPOSED-SUITE.prompt.md`

## Environment Rules

- `setup.md` prepares `build/pve-user.env` and `build/ubuntu-24-with-agent.vm-template.id` for follow-up suites.
- `setup.md`, `phase-5-privilege-root/RUN-PHASE-5-SUITE.prompt.md`, and `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md` require `build/pve-root.env`.
- All other prompts should run with `build/pve-user.env`.
- Action prompts should resolve runtime inputs locally and clean up disposable assets in the same run.

## Layout

- `setup.md`
- `phase-1-read-task/`
- `phase-2-vm-lifecycle/`
- `phase-3-cloudinit-qga/`
- `phase-4-console-websocket/`
- `phase-5-privilege-root/`
- `composed-v0.1/`
- `e2e/`

## Run Record Template

Use this lightweight template for each full regression run:

```text
Regression Run Record
- date:
- operator:
- env:
  - pve-root.env: ok/missing
  - pve-user.env: ok/missing
  - template-id file: ok/missing
  - source iso: ok/missing

Execution (README order)
1) setup.md -> pass/fail (note)
2) phase-1 RUN -> pass/fail (note)
3) phase-2 RUN -> pass/fail (note)
4) phase-3 RUN -> pass/fail (note)
5) phase-4 RUN -> pass/fail (note)
6) phase-5 RUN -> pass/fail (note)
7) e2e bootstrap -> pass/fail (note)
8) e2e provision-template -> pass/fail (note)
9) composed-v0.1 RUN -> pass/fail (note)

Summary
- overall: pass/fail
- first_failed_step:
- key artifacts:
  - user env:
  - template id:
  - serial log:
  - e2e log file:
```
