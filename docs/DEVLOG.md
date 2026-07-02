# InfraSight — Dev Log & Resume Guide

Working notes for picking the project back up. Pairs with `CHANGELOG.md`
(what changed) — this is **where we are, how to build it, and what's next**.

Last updated: after commit `8de9507` (visual layer offline/graph/hygiene fixes).

---

## 1. What InfraSight is

Zero-config, **read-only** Linux host-discovery CLI (Go). One command scans a
host and renders a **dependency graph** — hardware, OS, processes, services,
containers, ports, websites, TLS certs, databases, packages — cross-linked into
real chains, output as JSON + a single self-contained interactive HTML report.
Every probe is non-destructive (files O_RDONLY; only read-only commands).

- **Repo:** `A:\infrasight` (Windows) = `/mnt/a/infrasight` (WSL).
- **GitHub:** https://github.com/Aakhri-Pastaa/infrasight (private). Author: Kunal
  Patil <kunalrpatil324@gmail.com> only — **no AI attribution** in commits/docs.

---

## 2. Build / dev environment (READ THIS FIRST — it's split)

Two toolchains, two OSes. This trips people up.

| Task | Where | Command |
|---|---|---|
| **Go build/test** | **WSL** (Ubuntu) | `go` is `/usr/local/go/bin/go` (Kunal's, on PATH in an interactive shell). go1.26.4. |
| **Frontend build** | **Windows** | Node **v24** + npm 11 (Windows only; WSL has **no node**). |
| **Git push** | WSL | `wsl -d Ubuntu -- git -C /mnt/a/infrasight push origin main` |
| **gh / CI watch** | WSL | `gh` authed as Aakhri-Pastaa in WSL only |

### Gotchas (hard-won)
- **Non-interactive WSL shells don't load `.bashrc`.** `wsl -- go ...` or
  `bash -lc` → "command not found". Use the **full path** `/usr/local/go/bin/go`,
  or `bash -lic` (interactive). Go's `-C <dir>` flag avoids `cd`.
- **PowerShell → wsl → bash nested quoting breaks constantly.** For anything with
  quotes/`$`/`()`, write a script file or use direct `wsl -- <prog> <args>`
  (each arg separate, no shell). The Windows `$PATH` has spaces/parens that
  corrupt `export PATH=...`.
- There's a **second, unused** Go at `/home/kunal/.local/go` (installed session 1).
  Ignore it; `/usr/local/go` is the one on PATH.
- `git` on Windows works too, but **push via WSL git** (gh credentials live there).
- `Remove-Item` is sandbox-blocked on some paths; prefer `wsl -- rm`.
- Piping a here-string to `git commit -F -` in PowerShell prepends a BOM — commit
  via the Bash tool heredoc instead.

### Full build + run
```powershell
# 1. Frontend (Windows) — only when web/ changed:
cd A:\infrasight\web ; npm run build ; cd ..
Copy-Item web\dist\index.html internal\output\html\assets\index.html -Force
# 2. Go (WSL):
wsl -d Ubuntu -- /usr/local/go/bin/go -C /mnt/a/infrasight build -o bin/infrasight ./cmd/infrasight
# 3. Run:
wsl -d Ubuntu -- /mnt/a/infrasight/bin/infrasight scan --security -o /mnt/a/infrasight/infrasight
```
Or in an interactive WSL shell: `make build` (Go), `make frontend` (Node UI).

### View the HTML report
Serve + open in Chrome (headless Edge / file:// are blocked in this sandbox):
```powershell
wsl -d Ubuntu -- python3 -m http.server 8099 --directory /mnt/a/infrasight --bind 127.0.0.1
# then navigate a browser to http://localhost:8099/infrasight.html
```
Verified by driving the page via the Chrome MCP `javascript_tool` (click tabs,
read DOM, cross-check against the `#infrasight-data` JSON island).

---

## 3. Architecture snapshot

```
cmd/infrasight            entrypoint
internal/
  cli/                    Cobra: scan, diff, version (+ filterModules)
  discovery/              Module interface + concurrent Engine
    hardware/ osinfo/ network/ packages/ services/ web/ database/ resources/
    pkgbackend/           dpkg|rpm|apk|pacman abstraction (List/Owner/Detect)
    certinfo/             X.509 parse (stdlib)
  registry/               assembles the 11 modules (no import cycle)
  graph/                  Node/Edge schema, dedup Builder, CrossLink()
  output/                 Document + json/html/terminal renderers, Redact, security wiring
  diff/  security/  store/
pkg/  shell/  version/
web/                      React 19 + Vite + Tailwind single-file HUD (frontend)
  src/App.jsx (~1509 lines), src/fonts/material-symbols.woff2 (subset)
docs/  IMPACT_ENGINE_SPEC.md, DEVLOG.md
.github/workflows/ci.yml  gofmt, vet, test -race, cross-compile matrix
```

- **11 modules:** hardware.cpu, hardware.memory, os.distro, network.ports,
  packages, services.systemd, services.docker, web.nginx, web.apache, database,
  resources.system.
- **Node types (emitted):** OS, HARDWARE, SERVICE, PROCESS, CONTAINER, DATABASE,
  PORT, WEBSITE, CERTIFICATE, PACKAGE.
- **Edges:** RUNS_ON, LISTENS_ON, SERVES, PROXIES_TO, ENCRYPTS, DEPENDS_ON,
  CONTAINS. Cross-linked in `graph.CrossLink()` (pure, post-merge, no I/O).
- **Data model detail:** see `internal/graph/schema.go`.

### Adding a discovery module
Implement `discovery.Module` under `internal/discovery/<domain>/` (split
OS-specific logic into `*_linux.go` / `*_other.go`), then append its constructor
in `internal/registry/registry.go`. Engine calls `Available()` then `Probe(ctx)`.

---

## 4. CLI surface
```
infrasight scan [--format json|html|all] [-o base] [--modules ..] [--exclude-modules ..]
                [--security] [--redact] [--save NAME] [--quiet] [--parallel N] [--timeout d]
infrasight diff <baseline> <current>      # 0 none / 1 drift / 2 critical regression
infrasight version
```
Scan exit codes: 0 clean / 1 warnings / 2 critical. `--security` and `diff` also
gate exit code — usable as CI checks. `INFRASIGHT_NGINX_CONF`, `_APACHE_CONF`,
`_PG_CONF`, `_MYSQL_CONF`, `_REDIS_CONF` override config discovery (also used by
fixture tests).

---

## 5. Current state (verified)

- Backend: **solid**. 11 modules, cross-linking, `go vet` + `go test -race` +
  cross-compile (linux amd64/arm64, darwin arm64) all green in CI. External code
  review scored the pre-visual codebase 8/10; all its P1 items (F1-F5, E1, E2)
  are done.
- Visual layer: React HUD, **offline-correct** (0 CDN refs, subset font inlined),
  Graph view renders (vis canvas), Packages tab = real 638-row table, no JS
  errors across tabs. Committed + CI green.

---

## 6. Open issues / next steps (priority order)

1. **[BUG, do first] Security tab shows MOCK data.** `web/src/App.jsx` renders
   hardcoded sample findings, not the real `doc.security` array (verified: real
   findings = `[]` on a clean host, UI shows a fake "Weak RSA / CRITICAL"). Wire
   it to `doc.security` + add an empty-state. **Sweep App.jsx for other hardcoded
   demo data** (Dashboard cards, diff, etc.) — assume more exists.
2. **Deploy Agent tab** = stub. Read `docs/IMPACT_ENGINE_SPEC.md` and scope it.
3. **Visual design polish** — owner/designer pass (HUD is functional).
4. Small polish: `--open` via xdg-open; hide `--deep`; comment the single-OS
   anchor assumption in `CrossLink`; `-v` verbose logging.
5. Roadmap (later): language dep trees (npm/pip/go), cloud metadata, CVE feed,
   `--watch` daemon.

---

## 7. Workflow conventions
- **Commits:** author = Kunal Patil only, **no AI/co-author trailers**. Clear,
  descriptive messages.
- **Every change:** gofmt + `go vet` + `go test -race` + build locally, then
  commit, push (WSL git), and watch CI to green:
  `wsl -d Ubuntu -- gh run watch <id> --repo Aakhri-Pastaa/infrasight --exit-status`.
- **Verify, don't assume:** run on the WSL host, cross-check output against
  ground truth (`df`, `dpkg -l`, the JSON island). This has caught real bugs
  (disk reserved-block math; the security mock data).
- Fixtures with generated keys/certs go under `.fixture/` (gitignored); delete
  after. `web/node_modules` + `web/dist` are gitignored.
