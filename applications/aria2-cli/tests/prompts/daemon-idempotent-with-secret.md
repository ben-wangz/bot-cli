# daemon idempotent with secret

1. set env `ARIA2_RPC_SECRET` to a fixed value
2. call `capability ensure_daemon_started --rpc-endpoint http://127.0.0.1:6820/jsonrpc`
3. call the same command again with same endpoint and secret source
4. verify first response has `ok=true`
5. verify second response has `ok=true` and `result.already_running=true`
