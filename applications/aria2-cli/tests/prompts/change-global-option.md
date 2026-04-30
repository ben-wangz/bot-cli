# change global option

Goal: cover all args described by `capability describe change_global_option`.

1. call `capability ensure_daemon_started`
2. call `capability change_global_option --options '{"max-concurrent-downloads":"1","split":"1"}'`
3. call `capability change_global_option --option max-tries=1 --option timeout=60 --option connect-timeout=30`
4. call `capability get_global_option`

Verify:

- every response has `ok=true`
- step 2 and step 3 return `diagnostics.applied_keys`
- step 4 result includes expected values for:
  - `max-concurrent-downloads=1`
  - `split=1`
  - `max-tries=1`
  - `timeout=60`
  - `connect-timeout=30`
