# change global option

Goal: cover all args described by `capability describe change_global_option`.

1. ensure runtime dependency exists: `command -v aria2c || sudo dnf install -y aria2`
2. call `capability ensure_daemon_started`
3. call `capability change_global_option --options '{"max-concurrent-downloads":"1","split":"1"}'`
4. call `capability change_global_option --option max-tries=1 --option timeout=60 --option connect-timeout=30`
5. call `capability get_global_option`

Verify:

- every response has `ok=true`
- step 3 and step 4 return `diagnostics.applied_keys`
- step 5 result includes expected values for:
  - `max-concurrent-downloads=1`
  - `split=1`
  - `max-tries=1`
  - `timeout=60`
  - `connect-timeout=30`
