# Vendored assets

## vis-network.min.js

- **Library:** [vis-network](https://visjs.github.io/vis-network/) (standalone UMD bundle,
  includes vis-data; exposes the global `vis` with `vis.Network` and `vis.DataSet`).
- **Version:** 9.1.9 (pinned)
- **Source:** https://unpkg.com/vis-network@9.1.9/standalone/umd/vis-network.min.js
- **Size:** 688911 bytes
- **SHA-256:** `f53f833ddb9bf97efe856bb0637d4fe88f39e39999c7e94a4b8afc8de8a1a2e5`
- **License:** dual MIT / Apache-2.0 (© vis.js contributors).

It is vendored (not fetched at runtime) so generated reports are fully self-contained and work
in air-gapped environments. The renderer embeds it via `//go:embed` and inlines it into each
HTML report as a base64 `data:` URI.

### Integrity

The pinned SHA-256 is recorded in `vis-network.min.js.sha256`. CI verifies it on every run, and
you can check it locally:

```bash
make verify-assets            # sha256sum -c against the recorded hash
make vendor-vis               # re-fetch from upstream and verify it still matches
```

### Upgrading

```bash
# bump VIS_VERSION in the Makefile, then:
make vendor-vis-update        # re-fetch and rewrite the .sha256
```

Then review the diff, update the **Version/Size/SHA-256** above, and commit.
