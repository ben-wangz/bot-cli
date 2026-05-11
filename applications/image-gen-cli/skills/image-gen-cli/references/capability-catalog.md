# Capability Catalog

## Commands

- `capability generate_image`
- `capability describe [<name>]`

## generate_image arguments

Required:

- `--prompt`

Optional:

- `--method` (`direct|tools`, default `direct`)
- `--model` (default `gpt-5.5`)
- `--image-model` (default `gpt-image-2`)
- `--stream` (`true`/`false`, default `true`)
- `--store` (`true`/`false`, default `false`)
- `--previous-response-id`
- `--size` (`1024x1024|1024x1536|1536x1024|auto`)
- `--quality` (`auto|high|medium|low`)
- `--output-format` (`png|jpeg|webp`)
- `--output-dir`
- `--output-name`

Compatibility notes:

- `--previous-response-id`, `--store`, and `--model` require `--method tools`.
- `diagnostics.preview_count` is deprecated and kept for backward compatibility.

For the complete argument contract (including low-frequency options), run:

```bash
image-gen-cli capability describe generate_image
```
