# minimal two-call chain

Prerequisites:

1. binary built at `applications/image-gen-cli/src/image-gen-cli`
2. env configured:
   - `IMAGE_API_BASE_URL`
   - `IMAGE_API_KEY`

Execution chain:

1. call `capability generate_image --prompt "A cozy cabin in snowy forest" --stream true`
2. verify response `ok=true`
3. capture `result.response_id` as `first_response_id`
4. verify `result.output_file` exists and can be opened as image
5. call `capability generate_image --prompt "Same cabin at night with northern lights" --stream false --store true --previous_response_id <first_response_id>`
6. verify response `ok=true`
7. verify second `result.output_file` exists and can be opened as image

Pass criteria:

- two calls both succeed
- second call explicitly uses `previous_response_id`
- both output files are generated from final result path
