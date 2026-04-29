# workflow smoke

1. call `workflow queue_add_and_wait --uri <url>`
2. call `workflow cleanup_completed`
3. verify `ok=true` for each response
