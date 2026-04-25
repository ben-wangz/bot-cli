# proxmox-cli Prompt Suites

This directory is the single entry for full regression coverage of `proxmox-cli` workflows.

- Covers composed virtual-workflow prompts and workflow-level e2e prompts.
- Includes composed suite runner and standalone workflow prompts under `e2e/`.
- Prompts are designed to run with minimal manual edits.

## Full Regression Entry

Run prompts in this order:

1. `setup.md`
2. `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md`
3. `e2e/PROVISION-TEMPLATE-FROM-ARTIFACT.prompt.md`
4. `composed-v0.1/RUN-COMPOSED-SUITE.prompt.md`

## Environment Rules

- `setup.md` prepares `build/pve-user.env` and `build/ubuntu-24-with-agent.vm-template.id` for follow-up suites.
- `setup.md` and `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md` require `build/pve-root.env`.
- All other prompts should run with `build/pve-user.env`.
- Prompts should resolve runtime inputs locally and clean up disposable assets in the same run.

## Layout

- `setup.md`
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
2) e2e bootstrap -> pass/fail (note)
3) e2e provision-template -> pass/fail (note)
4) composed-v0.1 RUN -> pass/fail (note)

Summary
- overall: pass/fail
- first_failed_step:
- key artifacts:
  - user env:
  - template id:
  - serial log:
  - e2e log file:
```
