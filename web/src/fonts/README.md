# Vendored font

## material-symbols.woff2

- **What:** a **subset** of Google's *Material Symbols Outlined* — only the ~39
  icons this UI uses (variable weight 100–700). ~44 KB.
- **Why vendored:** the report must work offline / air-gapped. This file is
  bundled (inlined as a base64 data: URI by Vite) so the HTML makes **no network
  requests**. Replaces the previous `fonts.googleapis.com` CDN links.
- **Source:** fetched once, at build time, from the Google Fonts CSS2 endpoint
  with `icon_names=` subsetting:
  `https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=<used icons>`
- **License:** Apache-2.0 (Material Symbols).

To regenerate after adding/removing icons: re-fetch with the updated
`icon_names` list, save here, and rebuild the frontend (`make frontend`).
