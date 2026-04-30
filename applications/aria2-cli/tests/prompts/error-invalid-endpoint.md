# error path: invalid endpoint

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. call any read capability with invalid endpoint, for example:
   - `--rpc-endpoint http://127.0.0.1:6809/jsonrpc`
3. verify response has `ok=false`
4. verify `error.code` is `network_error`
