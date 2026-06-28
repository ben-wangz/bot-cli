---
name: claude-agent-acp
description: Install, run, verify, or integrate the official @agentclientprotocol/claude-agent-acp adapter so ACP clients can call Claude Code through the Claude Agent SDK.
metadata:
  skill-version: "1.0.0"
  npm-package: "@agentclientprotocol/claude-agent-acp"
  github-repo: "https://github.com/agentclientprotocol/claude-agent-acp"
---

# Claude Agent ACP

Use this skill when installing, running, verifying, or integrating the official Claude Agent ACP adapter.

`claude-agent-acp` is an ACP stdio server backed by the Claude Agent SDK. It is not an MCP server and not an opencode skill provider.

## Quick Task Routing

- Installation, authentication, and binary verification: `references/install.md`.
- Startup, environment loading, model selection, permission modes, and smoke checks: `references/usage.md`.
- ACP client integration and prompt helper usage: `references/acp-client.md`.

## Operating Principles

- Use the official repository URL: `https://github.com/agentclientprotocol/claude-agent-acp`.
- Treat the adapter as a long-running stdio ACP process when used by a client.
- Keep credentials out of command output; do not print API keys, auth tokens, gateway headers, or raw `.env` values.
- Use project-local `.env` files only as private launcher input, and never commit them.
- Do not assume `.env` is loaded automatically by `claude-agent-acp`; the parent process must load environment variables before spawning it.
- Verify the installed binary with `command -v claude-agent-acp` before documenting a path.
- Do not guess ACP JSON-RPC shapes; inspect the installed SDK examples and generated types before building a client.
- Prefer environment variables or Claude settings for process defaults, and ACP `session/set_config_option` for per-session model switches.

## Output Checklist

- State whether the work covered installation, runtime configuration, or ACP client integration.
- Name the official package, binary, and repository when relevant.
- Summarize verification without printing credentials or raw environment values.
- If this skill content changed, mention that opencode may need restart/reload before the new content is available.
