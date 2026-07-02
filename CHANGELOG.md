# Changelog

All notable changes to InfraSight. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/); the project is pre-1.0 and
versioned by milestone rather than tag.

## [Unreleased]

### In progress / known gaps
- **Security tab renders mock data.** The React Security panel shows hardcoded
  sample findings (e.g. "Weak RSA Key / CRITICAL") instead of the real
  `doc.security` array. On a clean host (0 findings) it still shows fake alerts.
  Must wire to real data + add an empty-state. Audit `web/src/App.jsx` for other
  hardcoded/demo data while at it.
- **Deploy Agent tab** is a stub (button only) tied to the unbuilt
  `docs/IMPACT_ENGINE_SPEC.md`.
- **Visual design polish** — HUD is functional and offline-correct; final design
  pass is owner/designer work.
- Smaller backlog: `--open` (xdg-open), hide unimplemented `--deep`, comment the
  single-OS-anchor assumption in `CrossLink`, `-v` verbose logging.

## Visual layer — React HUD (`8de9507`)
### Added
- New frontend under `web/`: **React 19 + Vite + Tailwind**, single-file bundle
  (`vite-plugin-singlefile`) — "InfraSight Command Center" HUD. Tabs: Dashboard,
  Graph View, Security Findings, Packages Catalog, Settings, Deploy Agent.
- Packages now render as a searchable **table** (638 rows) instead of as canvas
  nodes — fixes the old "hairball".
- `make frontend` target (Node build) split from `make build` (Go only).
### Fixed
- **Air-gap restored.** Removed Google Fonts CDN links (Geist + Material
  Symbols). Material Symbols is now a self-hosted **~44 KB subset** woff2
  (`web/src/fonts/`, only the ~39 icons used), inlined by Vite; Geist falls back
  to the system stack. Report makes **zero network calls**.
- **Graph view** — the new shell had dropped vis-network (global `vis`
  undefined). Re-injected the embedded vis-network as a base64 `data:` script.
- **Hygiene** — gitignore `web/node_modules` + `web/dist`; `renderer.go` slimmed
  to embed only the built shell + vis + data.

## Databases — E2 (`a122c77`)
### Added
- `database` module: detects PostgreSQL / MySQL-MariaDB / Redis from config
  files + binaries (no connection, no creds, no root). Parses listening port per
  config style; emits `DATABASE -LISTENS_ON-> PORT` (merges with the live
  listener) and `-DEPENDS_ON-> PACKAGE`. Overridable via
  `INFRASIGHT_PG_CONF` / `_MYSQL_CONF` / `_REDIS_CONF`. Pure port parsers tested.

## Multi-distro packages + tests — F5/F4 (`fc02781`)
### Added
- `pkgbackend`: `Backend` interface (Key/List/Owner) with **dpkg, rpm, apk,
  pacman**, runtime `Detect()`. `packages` module is now generic (replaced
  `packages.dpkg`); `pkgmap` folded into `pkgbackend.OwnerNode`. Pure parsers
  per manager, unit-tested with fixture output (validates rpm/apk/pacman without
  those distros).
- Tests for `Engine.Scan` concurrency (fake modules) and `filterModules`.

## Security audit — E1 (`9e26134`)
### Added
- `--security`: audit pass over the graph (no extra I/O) — expired/expiring
  certs, weak RSA (<2048) keys, ports bound to any interface. Findings in
  terminal + JSON (`security` array) + HTML; high/critical raise exit code to 2
  (CI gate).

## Output safety + supply chain — F1/E6/F2/F3 (`4497892`)
### Added
- `--redact`: scrubs hostname, versions, bind addresses (`"redacted": true`);
  every scan prints a one-line sensitivity notice.
- Vendored vis-network pinned with SHA-256 (`.sha256` + asset README), verified
  in CI via `make verify-assets`.
### Fixed
- Documented the vis-network tooltip as plain-text (must not be `esc()`'d).

## CI (`796f503`, `8e3a70e`)
### Added
- GitHub Actions: gofmt gate, `go vet`, `go test -race`, build, and a
  cross-compile matrix (linux amd64/arm64, darwin arm64). Status badge in README.

## Drift detection (`5401df5`)
### Added
- `infrasight diff <baseline> <current>`: added/removed/version/health changes
  between two scans; `scan --save <name>` baselines under `~/.infrasight/scans`.
  Exit codes: 0 none / 1 drift / 2 regression to critical.

## Resource profiling (`b6010be`)
### Added
- `resources.system`: CPU utilisation/load (`/proc/stat`, `/proc/loadavg`) and
  per-filesystem disk usage (`statfs`), with health thresholds. Disk % uses df's
  reserved-block-aware math. Graph merge now escalates health to the worst signal.

## Authorship (`bddf618`)
### Changed
- All commits authored solely as **Kunal Patil <kunalrpatil324@gmail.com>**
  (history rewritten; AI co-author trailers removed). LICENSE/README credit Kunal.

## Web servers + TLS certs (`a0b2f7f`)
### Added
- `web.nginx` / `web.apache`: vhosts, listen ports, TLS certs, upstreams.
  `certinfo` parses X.509 with the standard library (no openssl). Chain:
  `port -SERVES-> website`, `cert -ENCRYPTS-> website`,
  `website -PROXIES_TO-> upstream port`.

## Cross-linking + systemd/docker (`d3f12c3`)
### Added
- `graph.CrossLink()` — generic post-merge edge inference (packageId →
  DEPENDS_ON, OS anchoring). `services.systemd` (service → process → package)
  and `services.docker` (container → published ports).

## Interactive graph (`e2842ab`)
### Added
- Self-contained offline HTML report on vendored vis-network (pre-React).

## Initial scaffold — v0.1 (`f24e472`)
### Added
- Concurrent module engine, typed graph model, deterministic output. First
  probes: hardware (cpu/memory), os.distro, network.ports, packages. JSON output,
  Cobra CLI, Makefile, GoReleaser, MIT license.
