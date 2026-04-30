# error path: gid not found

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. call `capability tell_status --gid deadbeefdeadbeef`
3. verify response has `ok=false`
4. verify `error.code` is `rpc_error`
