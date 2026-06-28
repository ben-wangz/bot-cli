# Claude Agent ACP Install

## Official Sources

- Repository: `https://github.com/agentclientprotocol/claude-agent-acp`
- NPM package: `@agentclientprotocol/claude-agent-acp`
- Binary: `claude-agent-acp`

## Install

Install globally when the binary should be available on `PATH`:

```bash
npm install -g @agentclientprotocol/claude-agent-acp
```

Run without a global install when a one-off launch is enough:

```bash
npx -y @agentclientprotocol/claude-agent-acp
```

After installing, verify the binary resolves:

```bash
command -v claude-agent-acp
```

The package depends on the Claude Agent SDK, which provides the Claude Code native binary through platform-specific optional dependencies. If the native binary is missing, reinstall without omitting optional dependencies.

## Authentication

The adapter supports terminal-based Claude authentication flows through `--cli` mode.

Claude subscription login:

```bash
claude-agent-acp --cli auth login --claudeai
```

Anthropic Console login:

```bash
claude-agent-acp --cli auth login --console
```

For remote environments where browser redirects are not practical, run the CLI login flow in the terminal session exposed by the host or ACP client.

Do not print tokens or credential files. If using a gateway, pass credentials through environment variables, credential helpers, or client-side auth metadata rather than command-line literals.

## Install Verification

Running `claude-agent-acp` directly starts a stdio server and waits for ACP messages, so it is normal for the process to appear idle.

For an end-to-end check, use an ACP client that can spawn `claude-agent-acp` and send `initialize`, `session/new`, and `session/prompt`.
