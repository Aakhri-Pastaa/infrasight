# Vendored assets

## vis-network.min.js

- **Library:** [vis-network](https://visjs.github.io/vis-network/) (standalone UMD bundle,
  includes vis-data; exposes the global `vis` with `vis.Network` and `vis.DataSet`).
- **Version:** 9.1.9
- **Source:** https://unpkg.com/vis-network@9.1.9/standalone/umd/vis-network.min.js
- **License:** dual MIT / Apache-2.0 (© vis.js contributors).

It is vendored (not fetched at runtime) so generated reports are fully self-contained and work
in air-gapped environments. The renderer embeds it via `//go:embed` and inlines it into each
HTML report as a base64 `data:` URI.

To upgrade: replace this file with a newer standalone UMD bundle and bump the version above.
