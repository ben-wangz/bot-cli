# error path: invalid secret

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. ensure daemon is running with a known secret
3. call `capability get_global_stat` using a wrong `--rpc-secret`
4. verify response has `ok=false`
5. verify `error.code` is `rpc_error`
