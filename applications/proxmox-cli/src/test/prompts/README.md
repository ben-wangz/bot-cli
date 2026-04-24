# proxmox-cli Prompt Suites

This directory is the single entry for full regression coverage of `proxmox-cli` actions and workflows.

- Covers Phase 1-5 action prompts and workflow-level e2e prompts.
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

## Environment Rules

- `setup.md` prepares `build/pve-user.env` and `build/ubuntu-24-with-agent.vm-template.id` for follow-up suites.
- `setup.md` and `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md` require `build/pve-root.env`.
- All other prompts should run with `build/pve-user.env`.
- Action prompts should resolve runtime inputs locally and clean up disposable assets in the same run.

## Layout

- `setup.md`
- `phase-1-read-task/`
- `phase-2-vm-lifecycle/`
- `phase-3-cloudinit-qga/`
- `phase-4-console-websocket/`
- `phase-5-privilege-root/`
- `e2e/`
