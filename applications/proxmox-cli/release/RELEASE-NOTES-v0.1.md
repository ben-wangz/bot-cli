# proxmox-cli v0.1 Release Notes

## Version

- Application: `proxmox-cli`
- Version: `0.1.0`
- Scope: v0.1 convergence closeout (M0-M5)
- Publish status: documentation complete; tag/release publishing intentionally skipped in this cycle

## Highlights

- Convergence milestones M0-M4 are completed, including duplicate prompt cleanup and out-of-scope legacy path removal.
- Regression entry is unified under prompt suites (`setup -> workflows -> composed-v0.1`).
- Lint CI path for Go setup is fixed and green on latest run.
- Release gate and checklist are documented for repeatable future publishing.

## Key Changes

1. Version management and release conventions
   - `forgekit` binary mapping is used as the version source of truth for `proxmox-cli`.
   - GitHub release tag rule is standardized as `proxmox-cli-v<semver>`.

2. Test asset convergence
   - Action-level prompt duplicates covered by workflows are removed.
   - Uncovered actions are consolidated into composed virtual-workflow chains.
   - Release regression gate defines mandatory execution order and pass criteria.

3. Workflow and CI hardening
   - `provision-template-from-artifact` runner script root resolution is stabilized for moved prompt paths.
   - Lint workflow `setup-go` target is corrected to existing module `go.mod`.
   - Formatting regressions were fixed with `gofmt` for `proxmox-cli` Go sources.

## Regression and Validation Status

- Full regression sequence: completed and accepted for v0.1 closeout.
- Required regression artifacts: present (`build/pve-user.env`, `build/ubuntu-24-with-agent.vm-template.id`).
- Latest lint workflow status: success.

## Compatibility and Operational Notes

- Main execution path remains user-scope VM lifecycle operations.
- Root scope remains limited to bootstrap-style setup operations.
- Historical root-shell path is removed and not part of v0.1 acceptance scope.

## Deferred for Future Publish Action

- Actual tag creation and push.
- GitHub Release workflow execution and artifact publication.
- These are intentionally deferred and must be completed when doing a public release.
