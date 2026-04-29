# error path: gid not found

1. call `capability tell_status --gid deadbeefdeadbeef`
2. verify response has `ok=false`
3. verify `error.code` is `rpc_error`
