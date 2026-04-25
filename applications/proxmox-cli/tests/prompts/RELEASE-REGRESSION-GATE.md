# proxmox-cli Release Regression Gate

## Purpose

Define the mandatory regression checks that must pass before creating and pushing a release tag.

## Hard Rule

- No release tag is allowed unless the full regression sequence passes.
- For `proxmox-cli`, release tag format is `proxmox-cli-v<semver>`.

## Required Regression Sequence

Run prompts in this order:

1. `applications/proxmox-cli/tests/prompts/setup.md`
2. `applications/proxmox-cli/tests/prompts/workflows/BOOTSTRAP-BOT-USER-POOL-ACL.prompt.md`
3. `applications/proxmox-cli/tests/prompts/workflows/PROVISION-TEMPLATE-FROM-ARTIFACT.prompt.md`
4. `applications/proxmox-cli/tests/prompts/composed-v0.1/RUN-COMPOSED-SUITE.prompt.md`

## Pass Criteria

- Every step above returns success (`ok == true` or suite `success == true`).
- No unresolved failure remains in logs.
- Required artifacts are present:
  - `build/pve-user.env`
  - `build/ubuntu-24-with-agent.vm-template.id`
  - Serial/e2e logs referenced by the run outputs.

## Run Record (Required)

For each release candidate, store one run record with:

- date/operator
- commit SHA under test
- each step result (pass/fail)
- first failure point (if any)
- key log/artifact paths

Use the template in `applications/proxmox-cli/tests/prompts/README.md`.

## Release Decision

- If all required steps pass, release tag creation is allowed.
- If any step fails, release tag creation is blocked until the regression run is green.
