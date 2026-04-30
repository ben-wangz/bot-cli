# workflow smoke

Use URL:

- `https://github.com/ben-wangz/bot-cli/releases/download/v0.1.0/aria2-cli-linux-amd64`

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. call `capability ensure_daemon_started`
3. call `workflow queue_add_and_wait --uri <url>`
4. call `workflow cleanup_completed`
5. verify `ok=true` for each response
