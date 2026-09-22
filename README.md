# InfraSight

[![CI](https://github.com/Aakhri-Pastaa/infrasight/actions/workflows/ci.yml/badge.svg)](https://github.com/Aakhri-Pastaa/infrasight/actions/workflows/ci.yml)

**A zero-config, read-only Go CLI that discovers the hardware, OS, services, containers, ports, websites, TLS certificates, databases and packages on a Linux host and cross-links them into a dependency graph, written as JSON and as a single self-contained, offline HTML report.**

> [!NOTE]
> **Status: pre-1.0, early but working.** The scanner — 11 discovery modules, cross-linking, drift detection, `--redact` and `--security` — is tested and runs on real Linux hosts. The HTML report is not finished: its Security tab currently shows sample findings instead of the scan's, and its Deploy Agent tab is a stub. See [Status and limitations](#status-and-limitations).
>
> Designed and directed by Kunal Patil; developed with AI coding assistants. See [AI disclosure](#ai-disclosure).

## What it does

One command scans the host it runs on. Each discovery module reads a source —
`/proc`, `/etc` configuration, or the output of a read-only command such as
`ss -tlnp` or `systemctl` — and emits typed nodes (`OS`, `HARDWARE`,
`SERVICE`, `PROCESS`, `CONTAINER`, `DATABASE`, `PORT`, `WEBSITE`,
`CERTIFICATE`, `PACKAGE`). A cross-linking pass then joins them into chains
such as *service → process → package* and *port → website → certificate*.

Every probe is **non-destructive**: files are opened read-only and only
well-known status and list commands are executed. InfraSight never modifies
the host. It is not a vulnerability scanner — there is no CVE lookup — and it
scans one host per run.

## Quick start

```bash
# Requires Linux and Go 1.26+ (on Windows, use WSL2)
git clone https://github.com/Aakhri-Pastaa/infrasight.git
cd infrasight
make build                                  # -> ./bin/infrasight
./bin/infrasight scan                       # -> infrasight.json + infrasight.html
./bin/infrasight scan --security --redact   # audit, and strip hostname/versions/addresses
```

Real output of the last command:

```text
wrote infrasight.json
note: written with --redact (hostname, versions and bind addresses removed)
wrote infrasight.html
InfraSight  host=[redacted]  modules=8  2889ms
  nodes=707 edges=95  [707 healthy  0 warning  0 critical]
    HARDWARE       10
    OS             1
    PACKAGE        660
    PORT           2
    PROCESS        17
    SERVICE        17
  skipped (unavailable): [database web.apache web.nginx]
security: no findings
```

Conditions: WSL2 running Ubuntu 26.04, commit `01442a5`. Three modules were
skipped because no web server or database is installed on that host.

## How it works

| Module | Source | Emits |
|---|---|---|
| `hardware.cpu` | `/proc/cpuinfo` | CPU model, logical/physical cores, sockets |
| `hardware.memory` | `/proc/meminfo` | RAM/swap totals + health from utilisation |
| `os.distro` | `/etc/os-release`, `/proc/sys/kernel/osrelease` | distro, version, kernel, hostname (graph anchor) |
| `network.ports` | `ss -tlnp` | TCP listening ports + owning processes (+ their package) |
| `packages` | dpkg / rpm / apk / pacman | installed packages (backend auto-selected by distro) |
| `services.systemd` | `systemctl` | running services → main process → unit's package |
| `services.docker` | `docker ps` | running containers → published host ports |
| `web.nginx` | `/etc/nginx/**` configs | virtual hosts, listen ports, TLS certs, upstreams |
| `web.apache` | `/etc/apache2`, `/etc/httpd` configs | virtual hosts, ports, TLS certs, proxies |
| `database` | postgres/mysql/redis configs + binaries | DATABASE nodes → listening port + owning package |
| `resources.system` | `/proc/stat`, `/proc/mounts` + statfs | CPU utilisation/load + per-filesystem disk usage |

The engine calls each module's `Available()` and skips modules whose OS or
tools are absent, then runs `Probe(ctx)` on a bounded worker pool, each under
its own timeout. TLS certificates are parsed with the standard library (no
`openssl`) and graded by expiry: **warning** under 30 days, **critical** once
expired.

**Cross-linking** turns the module outputs into one dependency graph rather
than disconnected islands. After all modules merge, `graph.CrossLink()`
infers edges from the data the nodes already carry — no extra I/O.
Hardware, processes and services are anchored to the OS node; anything that
records the package providing it (a process binary, a systemd unit file)
gets a `DEPENDS_ON` edge to that package. The web probes reuse the
`port:tcp:<n>` node IDs, so virtual hosts splice into the same graph:

```text
process --LISTENS_ON--> port --SERVES--> website --PROXIES_TO--> port (upstream)
cert    --ENCRYPTS--> website
service --CONTAINS--> process --DEPENDS_ON--> package
```

For example `rsyslog.service → rsyslogd → rsyslog (pkg)`, and
`nginx → :443 → api.example.com → :3000` with the site's certificate attached.

Data flow: `registry.All()` → filter by availability and flags →
`discovery.Engine.Scan()` → `graph.Builder` deduplicates and merges →
`Graph.CrossLink()` adds implicit edges → `output.BuildDocument()` → JSON and
HTML.

**Output formats:**

- **JSON** (`infrasight.json`) — schema-tagged, deterministic, machine-readable.
- **HTML** (`infrasight.html`) — one self-contained file that makes zero
  network calls. It is a React 19 + Vite + Tailwind single-file bundle with
  Dashboard, Graph View, Security Findings, Packages Catalog, Settings and
  Deploy Agent tabs. The graph is drawn by the vendored vis-network bundle,
  embedded as a `data:` script, and the scan data travels inside the page as
  a JSON island.

## Install

| Requirement | Version | Why |
|---|---|---|
| Linux host | — | Probes read `/proc`, `/sys` and `/etc`; on Windows, develop inside WSL2 |
| Go | 1.26+ | Build (`go.mod`) |
| `ss`, `systemctl`, `docker`, a package manager | whatever the host has | Used when present; modules without their tool are skipped |
| Node.js | 24 | Only to rebuild the web report (`make frontend`); the built bundle is committed |

Without `make`:

```bash
go run ./cmd/infrasight scan --format all
```

The repository can live on the Windows filesystem and build from WSL
(`cd /mnt/<drive>/infrasight && go build ./...`); cloning into the WSL
filesystem gives faster I/O.

## Usage

`infrasight scan --help` (verbatim):

```text
Usage:
  infrasight scan [flags]

Flags:
      --deep                      deep inspection (current modules already probe fully)
      --exclude-modules strings   skip these modules
      --format string             output format: json, html, all (default "all")
  -h, --help                      help for scan
      --modules strings           only run these modules (full name or domain, e.g. hardware)
      --no-color                  disable coloured output
      --open                      open the HTML report in a browser (not yet implemented)
  -o, --output string             output path or basename (extension added per format) (default "infrasight")
      --parallel int              max concurrent probe workers (default 50)
      --quiet                     suppress the terminal summary
      --redact                    redact hostname, versions and bind addresses (safe to share)
      --save string               also save this scan as a named baseline for 'infrasight diff'
      --security                  enable the security audit module over the graph
      --timeout duration          global scan timeout (default 5m0s)
```

The other commands are `infrasight diff <baseline> <current>`, a drift
report between two scans, and `infrasight version`.

```bash
infrasight scan                          # full scan -> JSON + HTML
infrasight scan --format json            # just JSON
infrasight scan --modules hardware,network
infrasight scan --exclude-modules packages

# Drift detection: capture a baseline, compare later
infrasight scan --save baseline          # saved under ~/.infrasight/scans/
infrasight scan --save now
infrasight diff baseline now             # arguments are a file path or a saved name
```

**Drift detection** (`diff`) reports nodes added and removed, version changes
and health changes between two scans — for example a package upgrade, a
service that stopped, or a disk that crossed into critical. It compares by
stable node ID, so back-to-back scans show *no* drift (no false positives
from changing metrics).

**Exit codes** — `scan`: `0` clean, `1` warnings, `2` critical findings.
`diff`: `0` no drift, `1` drift, `2` a regression to critical health, which
makes it usable as a CI gate.

**Security audit.** `infrasight scan --security` runs an audit pass over the
graph (no extra host I/O) and reports findings ranked by severity: expired or
expiring TLS certificates, weak keys (for example RSA under 2048 bits), and
ports reachable from any interface. Findings appear in the terminal and in
the JSON (`security` array), and they raise the exit code (high or critical
→ `2`), so it doubles as a CI security gate. The HTML report's Security tab
does not show them yet — see [Status and limitations](#status-and-limitations).

**The output is sensitive.** The probes are read-only and safe to run
anywhere, but the report they produce is not: `infrasight.json` and `.html`
contain the hostname, every open port and bind address, every running
service, and the full package inventory with versions — effectively a
reconnaissance sheet for the host. Store the output like a configuration
dump and keep it out of public issues. To share a report, use `--redact`,
which removes the hostname, all versions and bind addresses while keeping
the topology (the JSON is marked `"redacted": true`). Every scan prints a
one-line reminder of this.

## Testing

```bash
go vet ./... && go test -race ./...
```

27 tests across 10 packages, all passing at commit `01442a5` (run under WSL2,
Go 1.26.4). CI ([`ci.yml`](.github/workflows/ci.yml)) runs on every push to
`main` and every pull request: vendored-asset checksum, `gofmt`, `go vet`,
`go test -race`, and a cross-compile matrix for linux/amd64, linux/arm64 and
darwin/arm64. Configuration-parsing modules are tested against fixtures;
`INFRASIGHT_NGINX_CONF`, `_APACHE_CONF`, `_PG_CONF`, `_MYSQL_CONF` and
`_REDIS_CONF` point config discovery at them.

## Status and limitations

| Area | Status | Notes |
|---|---|---|
| Discovery modules (11) | Works | Verified on real Linux hosts |
| Cross-linking, JSON output | Works | Deterministic, schema-tagged |
| Drift detection (`diff`, `--save`) | Works | Compares by stable node ID |
| `--redact`, `--security` (terminal, JSON, exit code) | Works | |
| HTML report: graph and packages | Works | Offline, zero network calls |
| HTML report: Security tab | Partial | Shows hardcoded sample findings, not the scan's — known bug, next fix |
| HTML report: other dashboard panels | Partial | Not yet audited for hardcoded sample data |
| HTML report: Deploy Agent tab | Not built | Stub for the impact engine in [`docs/IMPACT_ENGINE_SPEC.md`](docs/IMPACT_ENGINE_SPEC.md) |
| `--deep`, `--open` | Not built | Accepted; `--open` is marked "not yet implemented", `--deep` changes nothing because the modules already probe fully |
| Language dependency trees, cloud metadata, CVE lookup, `--watch` | Not built | Planned |

- **Linux only.** Other platforms compile (CI cross-compiles darwin/arm64)
  through build-tagged stubs, but they are not a supported target.
- **One host per scan.** There is no remote or agent mode.
- **One OS node per graph.** `CrossLink()` anchors everything to a single OS
  node.

## Repository layout

```text
cmd/infrasight            entrypoint
internal/
  cli/                    Cobra commands (root, scan, diff, version)
  discovery/              Module interface + concurrent Engine
    hardware/ osinfo/     probe implementations (build-tagged per OS)
    network/ packages/
    services/             systemd + docker probes
    web/                  nginx + apache config parsers, vhost/cert builder
    database/             postgres/mysql/redis detection + config port parsing
    resources/            CPU utilisation/load + filesystem usage
    pkgbackend/           package-manager abstraction (dpkg/rpm/apk/pacman)
    certinfo/             parse X.509 certs (stdlib, no openssl)
  registry/               assembles the module list (no import cycle)
  graph/                  Node/Edge schema, dedup Builder, cross-linking
  output/                 Document model + json / html / terminal renderers
  diff/                   drift comparison + report between two scans
  security/               audit rules over the graph (certs, exposed ports)
  store/                  saved scans under ~/.infrasight/scans (baselines)
pkg/
  shell/                  safe, timed, read-only command execution
  version/                build-time version metadata
web/                      React + Vite + Tailwind source of the HTML report
docs/                     dev log and design specs
```

**Adding a module:** create `internal/discovery/<domain>/<name>.go`
implementing `discovery.Module` (split OS-specific logic into
`*_linux.go` / `*_other.go` with build tags so it cross-compiles), then
register its constructor in `internal/registry/registry.go`.

## Documentation

| Need | Start here |
|---|---|
| Build environment, current state, next steps | [`docs/DEVLOG.md`](docs/DEVLOG.md) |
| What changed, milestone by milestone | [`CHANGELOG.md`](CHANGELOG.md) |
| The proposed impact engine (not built) | [`docs/IMPACT_ENGINE_SPEC.md`](docs/IMPACT_ENGINE_SPEC.md) |

## AI disclosure

I designed InfraSight and made its decisions: what it discovers and how the
graph is modelled, the rule that every probe is read-only, the offline
single-file report, the redaction and security-gate behaviour, and what
ships when. AI coding assistants (Claude Code) did much of the development
under my direction: writing the Go probes and the React report, drafting
tests and documentation, and running builds and scans.

| Area | Who |
|---|---|
| Idea, scope, architecture, design decisions | Me |
| Trade-offs, what to cut, when to release | Me |
| Code, tests, documentation drafts, tooling runs | AI assistants, directed by me |
| Review, verification, approval | Me |

AI output is checked, not trusted. Every change is vetted, tested and built
before it is pushed, and CI must pass. Checking scan output against ground
truth (`df`, `dpkg -l`) has caught real errors: a disk-usage calculation that
ignored reserved blocks, since fixed, and the Security tab's sample
findings, which are still open.

## Credits

- [vis-network](https://github.com/visjs/vis-network) 9.1.9 — vendored in
  `internal/output/html/assets/`, dual-licensed Apache-2.0 / MIT,
  © Almende B.V. and the visjs contributors.
- [Cobra](https://github.com/spf13/cobra) for the CLI; React, Vite,
  Tailwind CSS and `vite-plugin-singlefile` for the report; a subset of
  Google's Material Symbols font (Apache-2.0).

## License

MIT © 2026 Kunal Patil
