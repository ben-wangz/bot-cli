# proxmox-cli Prompt Suites

This directory stores OpenCode prompt cases used for action-level and workflow-level positive-path validation.

- One prompt per action in A01-A43.
- One bootstrap workflow prompt for `ubuntu24-with-agent-template`.
- Prompt files should be executable by OpenCode with minimal manual edits.

## Execution Notes

- Bootstrap once per suite run via `e2e/BOOTSTRAP-UBUNTU24-WITH-AGENT-TEMPLATE.prompt.md`.
- Action prompts should resolve their own runtime inputs and remain order-independent.
- Action prompts should clean up disposable VM assets in the same prompt run.

## Layout

- `phase-1-read-task/`
- `phase-2-vm-lifecycle/`
- `phase-3-cloudinit-qga/`
- `phase-4-console-websocket/`
- `phase-5-privilege-root/`
- `phase-6-policy-cleanup/`
- `e2e/`
