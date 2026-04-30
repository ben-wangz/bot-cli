# daemon idempotent with secret

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. set env `ARIA2_RPC_SECRET` to a fixed value
3. call `capability ensure_daemon_started --rpc-endpoint http://127.0.0.1:6820/jsonrpc`
4. call the same command again with same endpoint and secret source
5. verify first response has `ok=true`
6. verify second response has `ok=true` and `result.already_running=true`
