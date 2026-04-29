# error path: invalid secret

1. ensure daemon is running with a known secret
2. call `capability get_global_stat` using a wrong `--rpc-secret`
3. verify response has `ok=false`
4. verify `error.code` is `rpc_error`
