# InfraSight

> A zero-config, read-only CLI agent that auto-discovers compute resources,
> services, packages, and network endpoints on a host and renders them as an
> interactive, exportable dependency graph.

**Status:** `v0.1` — early scaffold. The architecture (module interface,
concurrent engine, graph model, multi-format output) is in place and runs on
real Linux hosts. A representative set of probes is implemented; the full
catalogue from the design spec is on the roadmap below.

Every probe is **non-destructive**: files are opened read-only and only
well-known status/list commands are executed. InfraSight never modifies the host.

---

## What works today

| Module | Source | Emits |
|---|---|---|
| `hardware.cpu` | `/proc/cpuinfo` | CPU model, logical/physical cores, sockets |
| `hardware.memory` | `/proc/meminfo` | RAM/swap totals + health from utilisation |
| `os.distro` | `/etc/os-release`, `/proc/sys/kernel/osrelease` | distro, version, kernel, hostname (graph anchor) |
| `network.ports` | `ss -tlnp` | TCP listening ports + owning processes (+ their package) |
| `packages.dpkg` | `dpkg-query` | installed Debian/Ubuntu packages |
| `services.systemd` | `systemctl` | running services → main process → unit's package |
| `services.docker` | `docker ps` | running containers → published host ports |
| `web.nginx` | `/etc/nginx/**` configs | virtual hosts, listen ports, TLS certs, upstreams |
| `web.apache` | `/etc/apache2`, `/etc/httpd` configs | virtual hosts, ports, TLS certs, proxies |

TLS certificates are parsed with the standard library (no `openssl`), and graded
by expiry: **warning** under 30 days, **critical** once expired.

**Cross-linking** ties these together into a real dependency graph rather than
disconnected islands. After all modules merge, `graph.CrossLink()` infers edges
from the data the nodes already carry — no extra I/O. Hardware/processes/services
are anchored to the OS node; anything that records the package providing it (a
process binary, a systemd unit file) gets a `DEPENDS_ON` edge to that package.
The web probes reuse the `port:tcp:<n>` node IDs, so vhosts splice into the same
graph. The result is full chains:

```
process --LISTENS_ON--> port --SERVES--> website --PROXIES_TO--> port (upstream)
cert    --ENCRYPTS--> website
service --CONTAINS--> process --DEPENDS_ON--> package
```

e.g. `rsyslog.service → rsyslogd → rsyslog (pkg)`, and
`nginx → :443 → api.example.com → :3000`, with the site's cert hanging off it.

Output formats:

- **JSON** (`infrasight.json`) — schema-tagged, deterministic, machine-readable.
- **HTML** (`infrasight.html`) — single self-contained file, no CDN/fonts/network,
  works offline. An interactive **vis-network** dependency graph (shape by node
  type, colour by health), with summary cards, critical/warning banners, search,
  type filters, force-directed/hierarchical layouts, and a node detail panel.
  The vis-network bundle is vendored and embedded, so the report needs nothing
  external to open.

---

## Build & run

InfraSight targets **Linux servers**. Probes read `/proc`, `/sys`, and shell out
to `ss`, `dpkg-query`, etc., so it must be built and run on Linux. On Windows,
develop inside **WSL2**.

```bash
# Requires Go 1.26+
git clone https://github.com/Aakhri-Pastaa/infrasight.git
cd infrasight
make build            # -> ./bin/infrasight  (Linux binary)
./bin/infrasight scan # -> infrasight.json + infrasight.html
```

Or without make:

```bash
go run ./cmd/infrasight scan --format all
```

### Developing on Windows via WSL

The repo can live on the Windows filesystem and build from WSL:

```bash
# Inside WSL (Ubuntu):
cd /mnt/a/infrasight     # repo on the Windows A: drive
go build ./...
./bin/infrasight scan
```

For faster I/O you can instead clone into the WSL filesystem (`~/infrasight`).

---

## Usage

```text
infrasight scan [flags]

  --format string        json | html | all            (default "all")
  -o, --output string    output basename               (default "infrasight")
  --modules strings      only run these (name or domain, e.g. "hardware")
  --exclude-modules strings
  --parallel int         max concurrent workers        (default 50)
  --timeout duration     global scan timeout           (default 5m)
  --quiet                suppress the terminal summary
  --no-color
  --deep, --security, --open   accepted; see roadmap

infrasight version
```

Examples:

```bash
infrasight scan                          # full scan -> JSON + HTML
infrasight scan --format json            # just JSON
infrasight scan --modules hardware,network
infrasight scan --exclude-modules packages
```

Exit codes: `0` clean · `1` warnings · `2` critical findings.

---

## Architecture

```
cmd/infrasight            entrypoint
internal/
  cli/                    Cobra commands (root, scan, version)
  discovery/              Module interface + concurrent Engine
    hardware/ osinfo/     probe implementations (build-tagged per OS)
    network/ packages/
    services/             systemd + docker probes
    web/                  nginx + apache config parsers, vhost/cert builder
    pkgmap/               resolve which package owns a file (for cross-linking)
    certinfo/             parse X.509 certs (stdlib, no openssl)
  registry/               assembles the module list (no import cycle)
  graph/                  Node/Edge schema, dedup Builder, cross-linking
  output/                 Document model + json / html / terminal renderers
pkg/
  shell/                  safe, timed, read-only command execution
  version/                build-time version metadata
```

Data flow: `registry.All()` → filter by availability/flags →
`discovery.Engine.Scan()` runs probes on a bounded worker pool →
`graph.Builder` dedups & merges → `Graph.CrossLink()` adds implicit edges →
`output.BuildDocument()` → JSON / HTML.

### Adding a module

1. Create `internal/discovery/<domain>/<name>.go` implementing
   `discovery.Module` (split OS-specific logic into `*_linux.go` / `*_other.go`
   with build tags so it cross-compiles).
2. Register its constructor in `internal/registry/registry.go`.

The engine calls `Available()` to skip modules whose OS/tools are absent, then
`Probe(ctx)` under the module's own timeout.

---

## Roadmap

- **Done** — interactive vis-network graph; systemd + Docker probes;
  nginx/apache + TLS cert probes; the package/service/process and
  port → website → cert cross-link chains.
- **v0.2** — databases (Postgres/MySQL/Redis), resource profiling (CPU/mem/IO),
  language dependency trees (npm/pip/go).
- **v0.3** — security audit (`--security`), cloud metadata, CVE scan, `diff`.
- **v0.4** — WASM plugins, `--watch` daemon + live dashboard.

See the design spec for the full module catalogue and data model.

---

## License

MIT © 2026 Aakhri-Pastaa
