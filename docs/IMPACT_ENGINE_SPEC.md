# InfraSight — Impact Engine (Blast-Radius Reasoning)
## Engineering specification & build brief

**Status:** proposed · **Owner:** engineering agent · **Target:** InfraSight v0.3
**Prerequisite reading:** `internal/graph/schema.go`, `internal/graph/builder.go`, `internal/diff/diff.go`
**This document is the single source of truth for the feature. Build from it. Do not re-derive the strategy.**

---

## 0. TL;DR for the agent

Build a new package, `internal/impact`, that answers one question against the existing
discovery graph: **"if this node fails, what else is affected, how badly, and why?"**

The naive version ("traverse dependency edges from X") is *wrong* and will produce
confidently-incorrect answers that cause the very outages the feature claims to prevent.
The correct version is a **reverse traversal over a per-relation failure-propagation model**,
with substrate suppression, redundancy awareness, severity classes, and an explainable
evidence chain stamped with scan freshness.

The intellectual core of this feature is a small, reviewable **propagation table** (Section 4),
not the traversal code. Get that table right and tested first; everything else is plumbing.

**Definition of done:** `infrasight impact <node-id>` and an MCP tool `impact_of` both return a
severity-classed, evidence-backed, freshness-stamped impact set; the propagation table is
fixture-tested across ≥5 diverse host profiles; no answer is ever a bare yes/no.

---

## 1. Why this feature exists (do not skip — it governs every design choice)

### 1.1 The strategic problem
Every tool in this market — osquery, runZero, Faddom, Device42, Netdata — is a **camera**.
They tell you *what is*. None tell you *what would happen if you changed something*.
"What breaks if I touch this?" is the single most expensive question in operations, it is
answered today from human memory, and getting it wrong is a leading cause of self-inflicted
outages ("I didn't know anything else used that").

InfraSight already owns the one asset required to answer it that nobody else has in a free,
single-binary tool: **a typed, deterministic, explainable dependency graph.** This feature
converts InfraSight from an *inventory* into a *reasoner*. That is a category jump, and it is
the thing that makes the project worth releasing rather than "another GitHub project."

### 1.2 Why we specifically can build it (and a SaaS competitor cannot)
- Faddom et al. infer dependencies from **network traffic**. Traffic ≠ causal dependency
  (two hosts chatting doesn't mean one *needs* the other). Our graph is built from **config and
  process facts**, which carry real "needs" semantics.
- Our graph is **typed and deterministic**, so reasoning can be **explainable** (show the edges),
  which is the only way reasoning is safe to ship.
- We are read-only and offline, so the reasoning runs anywhere, including air-gapped, with no
  account, no agent, no traffic mirror.

### 1.3 The non-negotiable principle: trustworthy or nothing
A wrong impact answer is **worse** than no answer, because it will be used to make a change.
Therefore every output MUST:
1. show its evidence (the exact edge chain that produced the conclusion),
2. state its confidence and the age of the underlying scan, and
3. never collapse to a bare boolean ("safe to reboot: yes"). We present consequences and let a
   human decide.

If a design choice trades explainability for cleverness, choose explainability.

---

## 2. What "impact" means (precise definitions to prevent scope drift)

We model three distinct **failure scenarios** because the same node fails differently depending
on the verb:

| Scenario | Meaning | Example trigger |
|---|---|---|
| `stop`   | The node stops running but stays installed | `systemctl stop`, crash, OOM-kill, reboot |
| `remove` | The node is uninstalled / deleted | `apt remove`, disk wipe, `docker rm` |
| `expire` | The node loses validity but keeps running | TLS cert expiry, license lapse |

A query is always `(node, scenario)`. Default scenario is `stop`.

We classify each affected node into a **propagation severity** (distinct from a node's *health*):

| Severity | Meaning | Operator reading |
|---|---|---|
| `SEVERED`  | Cannot function at all without the failed node | hard outage |
| `DEGRADED` | Loses redundancy, encryption, or a non-sole dependency | works, but impaired/at-risk |
| `WARNED`   | Cosmetic or advisory effect | note it, no action forced |
| `NONE`     | Reachable in the graph but not actually affected | excluded from output |

**Out of scope for v1 (state explicitly, resist scope creep):** probabilistic/ML inference,
cross-host impact, performance/latency impact, "auto-remediation," predicting *when* something
will fail. v1 is deterministic graph reasoning over a single host's current scan.

---

## 3. The four failures the naive approach makes (proven against real scan data)

These were found by running the idea against an actual `infrasight.json`. The fixes ARE the design.

**F-1 — Edges point toward dependencies, so reachability must be REVERSED.**
Edges are emitted as `dependent --REL--> dependency` (e.g. `process:175 --DEPENDS_ON--> package:apt:chrony`).
Naively traversing *outbound* from a node yields its dependencies, not its dependents. Impact =
"who depends on X" = traverse the **reverse** graph. Build a reverse adjacency index once.

**F-2 — Relations have different failure semantics; treating them uniformly yields garbage.**
`RUNS_ON` to the OS is universal (everything runs on the OS) → suppress as substrate.
`ENCRYPTS` only `WARNS`/`DEGRADES`, never `SEVERS`. `DEPENDS_ON` a package only matters in the
`remove` scenario, not `stop` (a package isn't a running thing). See the propagation table (§4).

**F-3 — Real graphs are shallow on lateral edges; the killer demo needs edges we don't yet emit.**
On a bare host the graph is a star around the OS. The juicy edge — a service/app `CONNECTS_TO`
a database — is defined in the schema (`RelConnectsTo`) but **emitted by no probe today**. So the
feature's value is gated on collecting lateral edges. This is a feature: impact analysis becomes
the forcing function that tells us which edges to collect next (see §7, Phase 3).

**F-4 — Ubiquitous low-level nodes explode the result set.**
`package:apt:systemd` already has 8 dependents in the sample; `libc6` would have hundreds. Without
**dampening** (suppress propagation through substrate-tier nodes, rank by distance + criticality),
every interesting query returns "the whole machine." This is the #1 reason the feature would *feel*
broken even when technically correct.

Two more that bite after the above are fixed:

**F-5 — Affected is not binary.** Without severity classes (§2) operators can't triage.
**F-6 — No redundancy model → over-reporting.** Two upstreams, one dies = `DEGRADED`, not `SEVERED`.
Model redundancy crudely: *N edges of the same relation into the same role on a node = N-redundant.*

---

## 4. The propagation model — THE GEM (build and test this first)

A single table drives all reasoning. Each row says: for this relation, when the **target** of the
edge fails under a given scenario, what happens to the **source** (the dependent), and does this
relation count as substrate (suppressed unless explicitly requested).

> Reading: edge is `source --relation--> target`. Impact flows from a failed `target` back to its
> `source`. "Severity" is what the *source* suffers when the *target* fails.

| Relation | source → target meaning | `stop` | `remove` | `expire` | Substrate? | Redundancy-aware? |
|---|---|---|---|---|---|---|
| `DEPENDS_ON` | dependent → thing it needs | — | SEVERED | — | no | yes |
| `LISTENS_ON` | process/db → port it binds | SEVERED¹ | SEVERED | — | no | no |
| `SERVES` | port → website it fronts | SEVERED | SEVERED | — | no | yes |
| `PROXIES_TO` | website → upstream port | DEGRADED² | DEGRADED² | — | no | yes |
| `CONNECTS_TO` | client → service it calls | SEVERED³ | SEVERED | — | no | yes |
| `ENCRYPTS` | cert → website it secures | — | — | DEGRADED⁴ | no | no |
| `CONTAINS` | service → its process | SEVERED⁵ | SEVERED | — | no | no |
| `RUNS_ON` | runtime → OS/host | SEVERED | SEVERED | — | **YES** | no |
| `PARENT_OF` | parent proc → child proc | DEGRADED | DEGRADED | — | no | no |
| `SCHEDULES` | timer/cron → job it triggers | WARNED⁶ | SEVERED | — | no | no |
| `MOUNTS` | consumer → filesystem | SEVERED | SEVERED | — | no | yes |
| `INCLUDES` | config → included config | WARNED | DEGRADED | — | no | no |

**Footnotes (the nuance that makes it correct):**
1. A port failing under `stop` means the binder is what stopped; treat the listener relationship
   as severed for anything *downstream* of the port (websites, clients), not the binder itself.
2. `PROXIES_TO` is `DEGRADED` by default because a reverse proxy with a dead upstream returns 502
   but the proxy itself is up; promote to `SEVERED` only if it is the *sole* upstream (redundancy).
3. `CONNECTS_TO` severity depends on whether the client can run without the dependency. Default
   `SEVERED`; downgrade to `DEGRADED` if redundancy or `metadata.optional=true` is present.
4. `ENCRYPTS` + `expire`: site keeps serving but browsers reject it → `DEGRADED`, escalates to
   higher rank as `daysUntilExpiry` goes negative. Never `SEVERED` (the bytes still flow).
5. `CONTAINS` is near-identity (a service *is* its main process); failure is total but should be
   de-duplicated in output so we don't double-count the service and its process.
6. A stopped timer means the job won't fire next time → `WARNED` now (no immediate breakage);
   `remove` of the job target means it can never fire → `SEVERED`.

**Substrate suppression rule:** edges marked Substrate (currently `RUNS_ON`) are **excluded by
default** from impact output, because "if you destroy the OS, all 24 things die" is true but
useless. They are included only when the queried node *is itself* substrate, or with `--include-substrate`.

**Dampening rule (for F-4):** maintain a small, configurable set of "ubiquitous" package names
(`libc6`, `systemd`, `dbus`, `glibc`, …) tagged substrate-tier. Propagation **stops** at these
nodes' dependents boundary unless the query targets them directly. Also cap transitive depth
(default 4) and rank results by `(severity, distance, target-criticality)` so signal sorts first.

**This table is the spec's crown jewel. It must be:**
- expressed in code as data (a `map[RelationType]propagationRule`), not scattered `if` statements,
- unit-tested row-by-row with synthetic graphs,
- documented inline with the footnotes above so reviewers can argue each cell.

---

## 5. Public surfaces

### 5.1 CLI
```text
infrasight impact <node-id-or-label> [flags]

  --scenario string     stop | remove | expire           (default "stop")
  --depth int           max propagation hops             (default 4)
  --min-severity string warned | degraded | severed      (default "warned")
  --include-substrate   include OS/hardware/ubiquitous-pkg propagation
  --format string       text | json                      (default "text")
  --from <scan>         analyze a saved/!live scan instead of scanning now
```
- If `<node>` is ambiguous (label matches several IDs), print the candidates and exit 3.
- Exit codes: `0` no severed impact · `1` something degraded · `2` something severed.
  (Mirrors the existing scan/diff exit-code convention so it slots into CI.)

### 5.2 Text output (the demo — must look exactly this clear)
```text
impact of  website:api.example.com   (scenario: stop)

SEVERED (2) — cannot function without it:
  ✗ port :443            the listener that serves it
  ✗ nginx.service        its reason to run

DEGRADED (1):
  ⚠ cert *.example.com   encrypts this site · expires in 9 days

evidence:
  website:api.example.com  ←SERVES─  port:tcp:443  ←LISTENS_ON─  nginx[pid 812]
                                                    └DEPENDS_ON→  package:apt:nginx
confidence: HIGH · based on scan from 4 minutes ago (scan_20260625_…)
```
The **evidence block** and the **confidence line** are mandatory, not optional polish. They are
what make the feature trustworthy instead of alarming.

### 5.3 JSON output (stable, schema-tagged, deterministic — same discipline as the rest of the repo)
```json
{
  "$schema": "https://infrasight.dev/schema/impact/v1",
  "query": { "node": "website:api.example.com", "scenario": "stop", "depth": 4 },
  "affected": [
    { "nodeId": "port:tcp:443", "type": "PORT", "label": ":443",
      "severity": "severed", "distance": 1,
      "via": [{ "from": "website:api.example.com", "relation": "SERVES", "to": "port:tcp:443" }] }
  ],
  "summary": { "severed": 2, "degraded": 1, "warned": 0 },
  "confidence": { "level": "high", "scanAgeSeconds": 240, "scanId": "scan_20260625_131000" }
}
```

### 5.4 MCP server (the multiplier — see §8)
Expose `impact_of(node, scenario)`, `explain_node(node)`, and `riskiest_nodes()` as MCP tools so
an AI agent can ask the host real questions grounded in the typed graph. This is what turns the
tool into infrastructure other systems depend on.

---

## 6. Internal design

```
internal/impact/
  model.go        propagationRule table (THE GEM) + scenario/severity types
  engine.go       reverse-index build, traversal, dampening, redundancy, ranking
  evidence.go     reconstruct + format the edge chain that justified each result
  confidence.go   map scan age -> confidence level
  render.go       text renderer (mirrors internal/security/render.go style)
  engine_test.go  table-driven tests over synthetic graphs (one per rule + integration)
internal/output/impact/
  renderer.go     json renderer (mirrors internal/output/json)
internal/cli/
  impact.go       cobra command, arg resolution, exit codes
```

**Key engineering rules (consistency with the existing codebase):**
- Pure functions for all reasoning; no I/O in `internal/impact`. It consumes a `graph.Graph` and a
  scan timestamp, returns a result struct. (Same separation that makes the parsers testable.)
- Deterministic output: sort `affected` by `(severityRank desc, distance asc, nodeId asc)`.
- Reuse `graph.Node`/`graph.Edge`; do not fork the model.
- No new third-party dependencies. Standard library only, like the rest of the repo.
- Build with no build tags — this is platform-independent graph logic.

**Algorithm (reference):**
1. Build reverse adjacency: `target -> [edges]`.
2. BFS from the queried node over reverse edges, depth-bounded.
3. For each traversed edge, look up `propagationRule[relation]` for the scenario → severity.
   Skip if `NONE`; skip substrate edges unless requested; stop at dampening boundaries.
4. Apply redundancy: if a node has N≥2 edges of the same `(relation, role)` and only one path is
   broken, downgrade `SEVERED`→`DEGRADED` for that node.
5. Keep the **worst** severity per affected node (a node reached two ways takes the higher).
6. Record the shortest justifying path per node for evidence.
7. Rank, classify confidence from scan age, render.

---

## 7. Delivery phases (ship value at each step; never a 6-week dark tunnel)

**Phase 0 — Validation spike (before real code).** Implement the propagation table + a throwaway
traversal. Run it against **≥5 diverse real scans**: a web host, a DB host, a Docker host, a bare
host, a busy multi-service host. For each, answer: *does the output tell a competent admin
something they didn't already know?* Record results in `docs/impact_validation.md`.
- Gate: if it only ever returns OS/systemd/libc trivia → jump to Phase 3 (edge collection) first.
- Expected outcome: underwhelming on bare hosts, genuinely useful on real service hosts. That
  result simultaneously confirms the audience (consultants/SREs on real boxes) and the next edge
  to collect (`CONNECTS_TO`).

**Phase 1 — Core engine + CLI (`stop` scenario only).** Reverse index, table, severity, evidence,
confidence, text + json renderers, `infrasight impact`. Full test suite on synthetic graphs.

**Phase 2 — `remove` and `expire` scenarios + redundancy + dampening.** Complete the table, add the
ubiquitous-package suppression list and depth ranking.

**Phase 3 — Lateral edge collection (`CONNECTS_TO`).** This is what makes impact *non-obvious*.
Emit `CONNECTS_TO` edges by correlating established outbound sockets to listening ports:
- Read `/proc/<pid>/net/tcp` (+`tcp6`) or `ss -tnp state established`, map remote `ip:port` that
  is a local listener back to the owning process/service. Read-only, no new privilege.
- Wire into `graph.CrossLink` so app→db/app→cache chains appear. Re-run Phase 0 validation; the
  demo should now light up on real hosts.

**Phase 4 — MCP server.** Expose `impact_of`, `explain_node`, `riskiest_nodes`. See §8.

**Phase 5 — Report integration.** Add an "Impact" interaction to the HTML report: click a node →
highlight its severed/degraded set in the existing vis-network graph, colored by severity. Reuses
the graph that's already rendered; pure front-end.

---

## 8. The MCP multiplier (why this is the worthiness lever)

Shipping the impact engine as an **MCP server** lets any AI assistant ask the host grounded
questions: *"is this server safe to reboot?"*, *"what's the riskiest thing running here?"*,
*"explain this box to me — I just got paged."* The agent answers from the **typed graph**, not from
hallucinated `ps aux` parsing. We become the trustworthy eyes of AI ops on a Linux host.

Guardrails carry over: tools are **read-only and advisory**. `impact_of` returns consequences and
evidence; it never executes the change. The MCP layer must surface the confidence/freshness fields
so the calling agent can refuse to answer on stale data.

---

## 9. Acceptance criteria (the build is done when ALL pass)

**Correctness**
- [ ] Every relation in `schema.go` has a row in the propagation table, with the scenario matrix filled.
- [ ] A unit test exists per propagation rule, asserting severity per scenario on a synthetic graph.
- [ ] Reverse traversal proven: querying a leaf returns its dependents, not its dependencies.
- [ ] Substrate suppression: querying a normal node never lists the OS/hardware unless `--include-substrate`.
- [ ] Dampening: a graph with `libc6` having 200 dependents does NOT return 200 nodes for a normal query.
- [ ] Redundancy: a website with 2 upstreams, 1 down → that site is `DEGRADED`, not `SEVERED`.
- [ ] Worst-severity-wins when a node is reached by multiple paths.

**Trust / safety**
- [ ] Every affected node carries a non-empty `via` evidence chain.
- [ ] Output always includes confidence + scan age; no bare yes/no anywhere.
- [ ] Engine performs zero host I/O (verified: package imports no os/exec).
- [ ] Determinism: same scan in → byte-identical impact JSON out (golden-file test).

**Surfaces**
- [ ] `infrasight impact <node>` text + `--format json` both work; exit codes 0/1/2 correct.
- [ ] Ambiguous node label prints candidates, exit 3.
- [ ] `--from <saved-scan>` analyzes without re-scanning.
- [ ] MCP tools `impact_of`, `explain_node`, `riskiest_nodes` return the JSON schema above.

**Polish (the repo's existing bar)**
- [ ] gofmt-clean, `go vet` clean, `go test -race ./...` green, cross-compiles (linux/darwin).
- [ ] README gains an "Impact analysis" section with the demo output verbatim.
- [ ] JSON schema URL added; output marked deterministic like the scan document.
- [ ] `docs/impact_validation.md` records Phase 0 results on the 5 host profiles.

---

## 10. Risks & how to defend against them

| Risk | Consequence | Defense |
|---|---|---|
| Confidently-wrong impact answer | Causes the outage it promised to prevent | Mandatory evidence chain + confidence/freshness; never a boolean verdict |
| Result-set explosion via libc/systemd | Feature feels broken; users distrust it | Substrate tier + dampening + depth cap + ranking (§4) |
| Shallow graph on bare hosts | "It just says the OS" → looks useless | Phase 0 gate; prioritize `CONNECTS_TO` collection (Phase 3) |
| Stale scan reasoning | Right logic, wrong-because-old answer | Confidence degrades with scan age; MCP can refuse on stale data |
| Scope creep into ML/prediction | Never ships; loses the explainability moat | §2 out-of-scope list is binding for v1 |
| Over-modeling redundancy | Complexity without payoff | Crude N-edges heuristic only in v1; revisit with real data |

---

## 11. The one-line north star (put it in the README when shipped)

> **InfraSight: the read-only graph that lets you — and your AI — ask a server what would happen
> *before* you touch it.**

Stop being a camera. Become a reasoner. This feature is the difference between "another GitHub
project" and a piece of infrastructure people fight to have.

---

*Spec authored as a build brief. It modifies no existing source; it defines new packages under
`internal/impact`, `internal/output/impact`, `internal/cli/impact.go`, and an MCP entrypoint.*
