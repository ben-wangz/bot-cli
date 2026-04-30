# basic capability chain

Use aria2-cli self-release binary URL for reproducible validation:

- `https://github.com/ben-wangz/bot-cli/releases/download/v0.1.0/aria2-cli-linux-amd64`

Execution chain:

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. call `capability ensure_daemon_started`
3. call `capability add_uri --uri <url>`
4. capture `result` as gid
5. call `capability tell_status --gid <gid>`
6. call `capability pause --gid <gid>`
7. call `capability resume --gid <gid>`
8. call `capability remove --gid <gid>`
9. call `capability purge_download_result`

Verify:

- every response has `ok=true`
- pause/resume/remove response `result` equals input gid
