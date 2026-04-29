# basic capability chain

1. call `capability add_uri` with a test URL
2. capture gid from result
3. call `capability tell_status` for the gid
4. call `capability pause` then `capability resume`
5. call `capability remove`
6. call `capability purge_download_result`
