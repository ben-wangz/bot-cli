# error path: invalid endpoint

1. call any read capability with invalid endpoint, for example:
   - `--rpc-endpoint http://127.0.0.1:6809/jsonrpc`
2. verify response has `ok=false`
3. verify `error.code` is `network_error`
