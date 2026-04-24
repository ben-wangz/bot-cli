# proxmox-cli Prompt Suites

This directory stores OpenCode prompt cases used for action-level and workflow-level positive-path validation.

- One prompt per active action in the roadmap.
- Bootstrap workflow prompt for `bootstrap-bot-user-pool-acl`.
- Prompt files should be executable by OpenCode with minimal manual edits.

## Execution Notes

- Run `setup.md` once before suites; it prepares both `build/pve-user.env` and `build/ubuntu-24-with-agent.vm-template.id`.
- Use `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md` when bot user/pool ACL bootstrap is required.
- Only `e2e/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md` depends on `build/pve-root.env`; other prompts should use `build/pve-user.env`.
- Action prompts should resolve their own runtime inputs and remain order-independent.
- Action prompts should clean up disposable VM assets in the same prompt run.

## Layout

- `phase-1-read-task/`
- `phase-2-vm-lifecycle/`
- `phase-3-cloudinit-qga/`
- `phase-4-console-websocket/` (SSH control plane scope)
- `phase-5-privilege-root/`
- `e2e/`
