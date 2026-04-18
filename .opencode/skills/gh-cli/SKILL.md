---
name: gh-cli
description: |
  Work with GitHub via the gh CLI: manage pull requests, issues, releases,
  repositories, Actions workflows, and make raw API calls. Use when the user
  needs to create PRs, review issues, trigger workflows, query the GitHub API,
  or do any GitHub operation from the command line.
license: MIT
compatibility: opencode
metadata:
  audience: developers
  tool: gh
  tested-version: "2.89.0"
---

# GitHub CLI (gh) Skill

Master the GitHub CLI for pull requests, issues, releases, Actions, and more.

## Agentic Output Principles

When using gh in an agentic context, **token efficiency is critical**. Always apply these rules — they override the examples in all other sections:

### Core Rules

1. **Never use raw JSON without `--jq`.** `--json` alone produces verbose nested objects with IDs, bot flags, and unused fields. Always chain: `--json <minimal-fields> --jq '<filter>'`.

2. **Request only the fields you need.** Check the Minimal Fields table below. Adding extra fields multiplies output size with no benefit.

3. **Bound large outputs.** Commands like `gh run view --log`, `gh pr diff`, `gh api --paginate` can emit megabytes. Always pipe through `| head -n 100`, `grep <pattern>`, or use `--jq 'first(.[])'`.

4. **Prefer text output for existence/status checks.** `gh pr list` (no flags) outputs compact tab-separated text (~80 chars/line). The same with `--json` (~300+ chars/line). Use text unless you need to branch on a specific field.

5. **Set agentic environment baseline** before running gh commands:
   ```bash
   export NO_COLOR=1               # no ANSI escape codes
   export GH_NO_UPDATE_NOTIFIER=1  # suppress update banners
   export GH_PAGER=cat             # disable interactive paging
   ```
   This eliminates color codes, update notifications, and paging prompts — all add noise tokens.

### Minimal Fields by Task

| Task | Command | Notes |
|---|---|---|
| Does PR exist? | `gh pr list --head <branch> --json number --jq 'length'` | Returns 0 or 1, tiny output |
| PR list summary | `gh pr list --json number,title,state` | Omit author, labels, timestamps unless needed |
| PR merge status | `gh pr view 42 --json mergeable,reviewDecision` | Only fields needed for merge decision |
| PR author | `gh pr list --json author --jq '.[].author.login'` | Extract just login, not id/is_bot/name |
| PR labels | `gh pr list --json labels --jq '.[].labels[].name'` | Project to label names only |
| Issue exists + state | `gh issue list --json number,state` | Two fields, minimal output |
| Issue assignees | `gh issue list --json assignees --jq '.[].assignees[].login'` | Project to logins only |
| Release latest | `gh release list --json tagName,isLatest --jq 'first'` | Single object, not array |
| CI run status | `gh run list --limit 1 --json status,conclusion` | Two fields, limit results |
| Workflow enabled? | `gh workflow list --json name,state` | Minimal info for decision |

---

## Before You Start: Version Check

**This skill was generated for gh CLI v2.89.0 (released 2026-03-26).**

Always verify your installed version first:

```bash
gh --version
# Expected output: gh version 2.89.0 (2026-03-26)
```

**If your version differs,** inform the user:

> "Note: This skill was tested with gh v2.89.0. Your installed version is **X.Y.Z**.
> Some flags or subcommands may differ. To update:
> - **macOS/Linux (brew):** `brew upgrade gh`
> - **Ubuntu/Debian:** `sudo apt upgrade gh`
> - **Windows (winget):** `winget upgrade GitHub.cli`
> - Or download from: https://github.com/cli/cli/releases"

Most core commands remain stable across minor versions, so proceed with caution.

---

## Quick Start

Token-efficient patterns for common GitHub tasks:

```bash
# Pull requests: check if exists
gh pr list --head $(git branch --show-current) --json number --jq 'length > 0'

# Pull requests: list open (text form is compact)
gh pr list --state open

# Pull requests: list with filtering (JSON + jq form)
gh pr list --json number,title,state --jq '.[] | "#\(.number) \(.title) [\(.state)]"'

# Pull request: check merge readiness
gh pr view 42 --json mergeable,reviewDecision,statusCheckRollup --jq \
  '"\(.mergeable) / reviews: \(.reviewDecision)"'

# Pull request: approve
gh pr review 42 --approve

# Issues: list open (text form)
gh issue list --state open

# Issues: list with custom filter
gh issue list --json number,state --jq '.[] | select(.state == "OPEN") | "#\(.number)"'

# Issues: create and link to PR
gh issue create --title "Bug found" --label bug

# Releases: create
gh release create v1.0.0 --notes "First release"

# Releases: check latest
gh release list --limit 1 --json tagName,isLatest --jq 'first.tagName'

# Repositories: clone
gh repo clone owner/repo

# GitHub Actions: check latest run status
gh run list --workflow ci.yml --limit 1 --json status,conclusion --jq \
  '.[0] | "\(.status) / \(.conclusion)"'

# GitHub Actions: view run (bounded output)
gh run view 12345 --log | head -100

# API: list issues (with field projection)
gh api repos/{owner}/{repo}/issues --jq '.[].number'

# Search: find open PRs
gh search prs "state:open author:@me" --json number,title --jq '.[] | "#\(.number) \(.title)"'
```

---

## Core Commands

### Pull Requests (`gh pr`)

| Subcommand | Purpose |
|---|---|
| `create` | Create a new pull request |
| `list` | List pull requests in the repository |
| `view` | View details of a pull request |
| `checkout` | Check out a PR branch locally |
| `merge` | Merge a pull request |
| `close` | Close a pull request |
| `reopen` | Reopen a pull request |
| `review` | Add a review (approve/request changes/comment) |
| `diff` | View changes in a pull request |
| `checks` | Show CI status |
| `comment` | Add a comment to a PR |
| `edit` | Edit PR title/body |
| `ready` | Mark draft PR as ready for review |
| `revert` | Revert a pull request |
| `update-branch` | Update PR branch from base |
| `lock` / `unlock` | Lock/unlock PR conversation |

**Key flags:**
- `--draft` — create as draft
- `--base <branch>` — target branch (default: repository default)
- `--head <branch>` — source branch
- `--reviewer <handle>` — request review
- `--label <name>` — add labels
- `--milestone <name>` — add to milestone
- `--assignee <login>` — assign user
- `--fill` — auto-populate from commits
- `--web` — open in browser

### Issues (`gh issue`)

| Subcommand | Purpose |
|---|---|
| `create` | Create a new issue |
| `list` | List issues |
| `view` | View issue details |
| `close` | Close an issue |
| `reopen` | Reopen an issue |
| `comment` | Add a comment |
| `edit` | Edit title/body/labels/assignees |
| `delete` | Delete an issue |
| `lock` / `unlock` | Lock/unlock conversation |
| `pin` / `unpin` | Pin/unpin issue |
| `transfer` | Move issue to another repo |
| `develop` | Manage linked branches |

**Key flags:**
- `--label <name>` — add labels
- `--assignee <login>` — assign user
- `--milestone <name>` — add to milestone
- `--template <name>` — use issue template
- `--web` — open in browser

### Repositories (`gh repo`)

| Subcommand | Purpose |
|---|---|
| `create` | Create a new repository |
| `list` | List repositories |
| `clone` | Clone a repository |
| `view` | View repository info |
| `fork` | Create a fork |
| `edit` | Edit repository settings |
| `delete` | Delete a repository |
| `archive` / `unarchive` | Archive/unarchive repo |
| `sync` | Sync fork with upstream |
| `rename` | Rename repository |
| `set-default` | Set default repo for current dir |
| `deploy-key` | Manage deploy keys |
| `autolink` | Manage autolink references |
| `gitignore` | List gitignore templates |
| `license` | Explore licenses |

**Key flags:**
- `--public` / `--private` — visibility
- `--source <repo>` — fork source
- `--description <text>` — repo description

### Releases (`gh release`)

| Subcommand | Purpose |
|---|---|
| `create` | Create a new release |
| `list` | List releases |
| `view` | View release details |
| `edit` | Edit release |
| `delete` | Delete a release |
| `upload` | Upload assets to release |
| `download` | Download release assets |
| `verify` | Verify release attestation |
| `verify-asset` | Verify asset attestation |
| `delete-asset` | Remove an asset |

**Key flags:**
- `--notes <text>` — release notes
- `--draft` — create as draft
- `--prerelease` — mark as pre-release
- `--target <branch>` — target branch for tag

### Search (`gh search`)

| Subcommand | Purpose |
|---|---|
| `issues` | Search for issues |
| `prs` | Search for pull requests |
| `repos` | Search for repositories |
| `code` | Search code |
| `commits` | Search commits |

Syntax: `gh search issues "is:open label:bug"` (GitHub search syntax)

### GitHub Actions (`gh run`, `gh workflow`)

#### `gh run`
| Subcommand | Purpose |
|---|---|
| `list` | List recent workflow runs |
| `view` | View run details |
| `watch` | Watch run until completion |
| `rerun` | Rerun a failed run |
| `cancel` | Cancel a run |
| `delete` | Delete a run |
| `download` | Download run artifacts |

#### `gh workflow`
| Subcommand | Purpose |
|---|---|
| `list` | List workflows |
| `view` | View workflow details |
| `run` | Trigger a workflow (needs `workflow_dispatch`) |
| `enable` | Enable a workflow |
| `disable` | Disable a workflow |

### Secrets & Variables (`gh secret`, `gh variable`)

| Command | Purpose |
|---|---|
| `gh secret set <NAME>` | Create/update a secret |
| `gh secret list` | List secrets |
| `gh secret delete <NAME>` | Delete a secret |
| `gh variable set <NAME> --body <value>` | Create/update a variable |
| `gh variable list` | List variables |

### Authentication (`gh auth`)

| Subcommand | Purpose |
|---|---|
| `login` | Authenticate with GitHub |
| `logout` | Sign out |
| `status` | Show auth status |
| `token` | Print current auth token |
| `switch` | Switch GitHub account |
| `refresh` | Refresh stored credentials |
| `setup-git` | Configure git to use gh for auth |

---

## JSON Output & Formatting

### Basic JSON output

```bash
# Get JSON with specific fields
gh pr list --json number,title,author
gh issue view 42 --json title,state,assignees

# Filter with jq
gh pr list --json number,title,author --jq '.[].author.login'

# Format with Go templates
gh pr list --json number,title --template '{{range .}}#{{.number}} {{.title}}{{"\n"}}{{end}}'
```

### Available JSON fields

Each command supports different fields. View them without arguments:

```bash
gh pr list --json         # Lists all available fields for pr list
gh issue view --json      # Lists all available fields for issue view
gh release list --json    # Lists all available fields for releases
```

### jq filter examples

```bash
# Count open PRs
gh pr list --json number --jq 'length'

# Get PR titles that mention "bug"
gh pr list --json title --jq '.[] | select(.title | contains("bug"))'

# List PR authors
gh pr list --json author --jq '.[].author.login'
```

### Template functions

Advanced formatting with `--template`:

- `tablerow <fields>...` — align fields vertically as table
- `tablerender` — render previously added rows
- `timeago <time>` — convert timestamp to relative time
- `timefmt <format> <time>` — format timestamp with Go layout
- `color <style> <input>` — colorize output
- `autocolor <style> <input>` — color only in terminal
- `hyperlink <url> <text>` — create terminal hyperlink
- `join <sep> <list>` — join array with separator
- `pluck <field> <list>` — extract field from all items
- `truncate <length> <input>` — limit length

---

## Environment Variables

### Agentic Baseline

When running gh from an agent, always set these to eliminate noise tokens:

```bash
export NO_COLOR=1               # Disable ANSI escape codes (removes color sequences)
export GH_NO_UPDATE_NOTIFIER=1  # Suppress "new release available" banners
export GH_PAGER=cat             # Disable interactive paging (prevents less prompts)
```

Or inline per-command:

```bash
NO_COLOR=1 GH_NO_UPDATE_NOTIFIER=1 GH_PAGER=cat gh pr list
```

### All Environment Variables

| Variable | Purpose | Agentic use |
|---|---|---|
| `GH_TOKEN` | Authentication token (preferred over `GITHUB_TOKEN`) | Set in environment |
| `GH_HOST` | GitHub hostname (default: `github.com`) | Use for GitHub Enterprise |
| `GH_REPO` | Default repository (`OWNER/REPO`) | Avoids `-R` flag repetition |
| `GH_EDITOR` | Editor for composing text | Typically not used in agents |
| `GH_BROWSER` | Browser for opening links | Typically not used in agents |
| `GH_DEBUG` | Enable verbose output (`api` for HTTP details) | **Avoid in agents** (adds noise) |
| `GH_PAGER` | Paging program (e.g., `less`) | **Always set to `cat`** |
| `GH_CONFIG_DIR` | Config directory (default: `~/.config/gh`) | Use if custom location needed |
| `NO_COLOR` | Disable ANSI colors | **Always set to 1** |
| `CLICOLOR_FORCE` | Force ANSI colors when piped | **Do not use** (contradicts NO_COLOR) |
| `GH_NO_UPDATE_NOTIFIER` | Suppress update notifications | **Always set to 1** |
| `GH_FORCE_TTY` | Force terminal output | **Do not use** (adds formatting noise) |

---

## Specific Tasks

Deep dives into common workflows:

* **Agentic Patterns** — [references/agentic-patterns.md](references/agentic-patterns.md) ⭐
  - Token efficiency strategies, output bounding, GraphQL bulk queries, jq reduction patterns
  - **Read this for optimized agentic command patterns**
  
* **Pull Request Workflows** — [references/pr-workflows.md](references/pr-workflows.md)
  - Create with templates, merge strategies, review workflows, draft PR best practices
  
* **Issue Management** — [references/issue-workflows.md](references/issue-workflows.md)
  - Create with templates, triage, bulk edits, linking branches
  
* **Repository & Release Management** — [references/repo-management.md](references/repo-management.md)
  - Clone, fork, sync, create releases, upload assets, manage settings
  
* **Search & API** — [references/search-and-query.md](references/search-and-query.md)
  - Advanced search syntax, REST API calls, GraphQL queries, pagination
  
* **GitHub Actions & Workflows** — [references/actions-and-ci.md](references/actions-and-ci.md)
  - Run workflows, trigger with `workflow_dispatch`, view logs, re-run failed jobs
  
* **Output Formatting Reference** — [references/output-formatting.md](references/output-formatting.md)
  - Complete `--json` fields per command, `--jq` patterns, template functions

---

## Tips & Gotchas

- **Branch vs. repo context:** Many commands infer repo from `.git/config`. Use `-R owner/repo` to override.
- **Pagination:** Use `--paginate` with `--json` and `--slurp` to fetch all results as a single array.
- **Draft PRs:** Use `--draft` on create, then `gh pr ready <number>` to un-draft.
- **Linking issues:** Include `Fixes #N` or `Closes #N` in PR body to auto-close issue on merge.
- **Private tokens:** Always use `GH_TOKEN` or store in `gh auth login`. Never commit tokens to git.
- **Enterprise GitHub:** Set `GH_HOST=github.company.com` for GitHub Enterprise.

---

## Learn More

- Official docs: https://cli.github.com/manual
- Release notes: https://github.com/cli/cli/releases
- API docs: https://docs.github.com/en/rest
- GraphQL explorer: https://docs.github.com/en/graphql
