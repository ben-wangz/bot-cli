# workflow smoke

Use URL:

- `https://github.com/ben-wangz/bot-cli/releases/download/v0.2.4/proxmox-cli-linux-amd64`

1. call `capability ensure_daemon_started`
2. call `workflow queue_add_and_wait --uri <url>`
2. call `workflow cleanup_completed`
3. verify `ok=true` for each response
