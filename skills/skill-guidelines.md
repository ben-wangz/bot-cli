# Skill Guidelines

This document defines the conventions for developing OpenClaw and OpenCode skills in this repository.

## Core Principles

- Keep each skill focused on one responsibility boundary.
- Skills may depend on other skills, but dependencies must not form cycles.
- `openclaw/skills` and `opencode/skills` are separate skill sets. A skill in one set must not depend on, route to, or reference a skill in the other set.
- Put reusable operational knowledge in the skill that owns the operation.
- Do not duplicate commands or procedures across skills. Reference the owning skill instead.
- Keep `SKILL.md` short. Move detailed procedures, command references, and long explanations into `references/`.
- Prefer explicit cross-skill routing over copying implementation details.

## Dependency Rules

Skills can depend on one or more other skills.

Dependencies are only valid inside the same skill set. OpenClaw skills may depend only on skills available to OpenClaw. OpenCode skills may depend only on skills available to OpenCode.

Declare dependencies in `SKILL.md` with a `## Dependencies` section:

```md
## Dependencies

Load and use these skills when executing this workflow:

- `gitea-tea`: Gitea repository, issue, pull request, comment, review, attachment, and `tea` operations.
- `openclaw-manage-opencode`: OpenCodeInstance lifecycle, readiness, endpoint discovery, and auth discovery.
```

Dependency rules:

- Dependencies must be directional and acyclic.
- Dependencies must not cross the `openclaw/skills` and `opencode/skills` boundary.
- A workflow skill may depend on lower-level capability skills.
- A lower-level capability skill should not depend on a higher-level workflow skill.
- If two skills need the same operation, move that operation into the skill that owns the domain and reference it from the other skill.
- If a dependency is optional, say when it should be loaded.

## Loading Skills

When instructing another agent or skill in the same skill set to load a skill, use this exact format:

```md
load skill `skill-name`
```

Examples:

- Need Gitea issue attachment upload? -> load skill `gitea-tea`.
- Need OpenCode worker lifecycle commands? -> load skill `openclaw-manage-opencode`.
- Need OpenCode session submission and wait logic? -> load skill `opencode-workmate`.

Avoid vague wording such as "use the Gitea skill" when an executable instruction is needed.

Do not use `load skill ...` to reference a skill from another skill set. For example, an OpenClaw workflow may delegate repository work to an OpenCode worker, but it must not tell that worker to load a skill from `opencode/skills`.

## `SKILL.md` Boundary

`SKILL.md` is the skill entry point. It should contain only the information needed to select and route the skill.

Good content for `SKILL.md`:

- Frontmatter `name` and `description`.
- One short purpose statement.
- `## Dependencies`, when the skill relies on other skills.
- `## Quick Task Routing` that points to files under `references/`, `templates/`, or other skills.
- Core operating principles and safety rules.
- A short output checklist.

Avoid putting these in `SKILL.md`:

- Long command sequences.
- Full troubleshooting guides.
- Large examples.
- Repeated knowledge already owned by another skill.
- Environment-specific manifests or generated artifacts.

If `SKILL.md` becomes long, split content into `references/` and route to it.

## `references/` Boundary

Use `references/` for detailed knowledge that explains how to perform work.

Good content for `references/`:

- Command references.
- Lifecycle procedures.
- Decision rules.
- Domain models.
- Safety rules that need detail.
- Troubleshooting notes.
- Cross-skill responsibility boundaries.

Rules for `references/`:

- A reference file may contain concrete commands only for operations owned by the current skill.
- For operations owned by another skill, say which skill to load and which reference section to use.
- For operations owned by another skill in the same skill set, say which skill to load and which reference section to use.
- For operations performed by another runtime with a separate skill set, describe the required outcome and delegate the work without naming that runtime's skills.
- Keep reference files organized by task area, not by arbitrary length.
- Prefer stable headings so other skills can reference them, such as `references/collaboration-commands.md > ### Attachments`.

## `templates/` Boundary

Use `templates/` for reusable files that are intended to be copied, rendered, or adapted.

Good content for `templates/`:

- Kubernetes manifests.
- Config skeletons.
- Prompt skeletons.
- Issue or PR body templates.
- Script templates that are meant to be customized before use.

Rules for `templates/`:

- Templates should be generic and parameterized.
- Do not put live secrets, tokens, cluster-specific values, or user-specific paths in templates.
- Use obvious placeholder names such as `<worker-name>`, `${NAMESPACE}`, or `${REQUIREMENT_ID}`.
- Do not use `templates/` for command documentation; use `references/` instead.
- Do not use `templates/` for completed examples; use `examples/` instead.

## `scripts/` Boundary

Use `scripts/` for executable helper programs that perform reusable operational work for the skill.

Good content for `scripts/`:

- Test or debugging helpers.
- Local automation wrappers.
- Server lifecycle helpers.
- Data collection helpers.
- Small utilities that are meant to be invoked, not copied into the user's codebase.

Rules for `scripts/`:

- Scripts must be generic and reusable across projects.
- Scripts must be invokable via an absolute path and must not depend on the caller's current working directory.
- Document script usage in `references/`, not only in comments inside the script.
- Prefer safe defaults and clear failure messages.
- Do not embed live secrets, tokens, cluster-specific values, or user-specific paths.
- If a script accepts shell commands or user-provided executable input, document the trust boundary clearly.

## `examples/` Boundary

Use `examples/` for concrete samples that demonstrate expected usage or output.

Good content for `examples/`:

- Example prompts.
- Example issue descriptions.
- Example PR summaries.
- Example sanitized outputs.
- Example workflow transcripts with secrets removed.

Rules for `examples/`:

- Examples must be safe to read and share.
- Do not include real credentials, tokens, raw Secret data, or authenticated Git remotes.
- Make examples clearly non-authoritative. The implementation guidance belongs in `references/`.
- Keep examples small enough to illustrate behavior without becoming a second copy of the reference documentation.

## Cross-Skill References

Use cross-skill references when another skill owns an operation.

Cross-skill references are allowed only within the same skill set.

Example:

```md
For the upload operation, load skill `gitea-tea` and use its issue attachment guidance (`references/collaboration-commands.md > ### Attachments`).
```

Do not repeat the concrete upload command in the calling skill. This keeps command ownership clear and prevents documentation drift.

Do not write cross-set references such as an OpenClaw skill telling a worker to load an OpenCode skill, or an OpenCode skill telling OpenClaw to load an OpenClaw skill.

## Ownership Examples

- Gitea repositories, issues, pull requests, comments, reviews, attachments, and `tea` commands belong to `gitea-tea`.
- OpenCodeInstance lifecycle, Kubernetes readiness, endpoint discovery, and auth discovery belong to `openclaw-manage-opencode`.
- OpenCode session creation, prompt submission, waiting, blocker handling, and output collection belong to `opencode-workmate`.
- Multi-role process orchestration belongs to workflow skills such as `openclaw-standard-team-workflow`.

Workflow skills should describe what needs to happen and delegate the how to the owning capability skill.

## Safety

- Never document commands that print tokens, passwords, raw Secret data, or authenticated Git remotes.
- Prefer Secret names, login names, repository names, issue links, PR links, and artifact names in outputs.
- When a command requires credentials, prefer credential files, Secret references, helpers, or other mechanisms that avoid exposing values in command output. If the underlying tool commonly requires credentials through arguments or environment variables, document the safest practical usage, warn not to print values, and avoid embedding literal credential values in examples.
- Mark destructive actions clearly and require explicit user intent for delete, force-push, close, merge, or credential rotation operations.

## Versioning

Use semantic versions for project skills: `MAJOR.MINOR.PATCH`.

Each skill stores its version in `SKILL.md` frontmatter:

```yaml
---
name: latex-writing
description: Use when ...
metadata:
  skill-version: "1.0.0"
---
```

`metadata` is supported by opencode skill frontmatter and keeps version data close to the skill content without changing runtime behavior.

Version bump rules:

- `PATCH`: wording fixes, examples, typo fixes, clarifications, or references that do not change expected behavior.
- `MINOR`: new capabilities, new reference files, new tool workflows, or expanded trigger coverage that remains backward-compatible.
- `MAJOR`: renamed skill, removed guidance, incompatible workflow changes, or behavior that may make old prompts produce materially different results.

## Review Checklist

Before adding or changing a skill, check:

- Does the skill have one clear responsibility boundary?
- Are dependencies declared in `SKILL.md`?
- Are dependency directions acyclic?
- Are cross-skill operations referenced instead of duplicated?
- Is `SKILL.md` short and route-focused?
- Are detailed procedures under `references/`?
- Are reusable skeleton files under `templates/`?
- Are concrete safe samples under `examples/`?
- Are executable helpers under `scripts/` and documented from `references/`?
- Are secrets and destructive operations handled safely?
