# Implementation Plan: vAssessment Phase 1

**Branch**: `2386-set-up-the-vassessment-project-folder-structure-and-basic-cli` | **Date**: 2026-09-15 | **Spec**: [spec.md](./spec.md)
**Research**: [research.md](./research.md) | **Data model**: [data-model.md](./data-model.md) | **Contracts**: [contracts/](./contracts/)

## Summary

Deliver vAssessment as a **single self-contained binary** that serves both a REST API and an embedded web UI, connects to one or more vCenters with read-only credentials (or ingests RVTools exports), inventories the estate, evaluates every check derivable from vCenter alone (Basic and Intermediate levels), lets an engineer choose per-VM copy methods, and generates a migration-scope report.

It reuses vJailbreak's existing vCenter connectivity where that code is not Kubernetes-coupled, and extracts the reusable core where it is. It builds, versions and releases on its own cadence, sharing nothing with the appliance release except the git repository and `pkg/common`.

## Technical Context

**Language/Version**: Go 1.24.10 (matching the other four modules)
**Primary Dependencies**: `govmomi` (vCenter), `spf13/cobra` (CLI), `modernc.org/sqlite` (pure-Go SQLite), an RVTools/xlsx reader, `gophercloud` (deferred to Phase 3)
**Storage**: SQLite, one file per assessment — immutable scans plus a mutable method-selection overlay
**Testing**: `go test` with table-driven tests; `govmomi/simulator` (vcsim) for vCenter behaviour without a live environment
**Target Platform**: linux and darwin, amd64 and arm64; no Kubernetes cluster required
**Project Type**: Standalone CLI + embedded web UI (single binary)
**Performance Goals**: 1,000-VM estate discovered in under 10 minutes
**Constraints**: Read-only against vCenter; no outbound calls beyond configured sources; scan file ≤ ~50 MB at 5,000 VMs; must run air-gapped
**Scale/Scope**: Estates to ~5,000 VMs across multiple vCenters; ~23 checks in Phase 1 of a ~41-check catalog

### Decision: pure-Go SQLite, not CGO

`modernc.org/sqlite` over `mattn/go-sqlite3`. `v2v-helper` already demonstrates the cost of the alternative — its tests cannot compile on macOS without a Linux cross-compilation toolchain (`CLAUDE.md`, Common Pitfalls), which slows every contributor down. vAssessment must cross-compile to four platform/arch combinations from CI with no toolchain gymnastics, and its performance profile (bulk insert, simple reads) does not need the C library's edge. **Accepting a modest performance cost to keep the build trivially portable is the right trade here.**

### Decision: embed the UI in the binary

The UI is built to static assets and embedded with `go:embed`. FR-031 requires one artifact; shipping a separate static bundle a user must serve themselves defeats the "download one binary and run it" premise.

### Decision: the UI must support a configurable base path from day one

Phase 1 serves the UI at the root of a local port. Phase 3 serves it behind an Ingress at a sub-path (`/vassessment`), exactly as Grafana is served at `/grafana` (`research.md` R3). Retrofitting sub-path support into a built frontend is materially harder than building it in, so the base path is configurable from the first commit even though Phase 1 always sets it to `/`.

## Constitution Check

*GATE: must be resolved before implementation begins.*

| Principle | Status | Notes |
|---|---|---|
| I — Kubernetes-Native Architecture | ⚠️ **CONFLICT — needs maintainer ruling** | See below |
| II — External Documentation First | Pass | govmomi and vSphere API docs are the authority for every property path; the check catalog cites them |
| III — Generated Code Protection | Pass | vAssessment generates no CRDs and hand-edits no generated file |
| IV — Test-First Development | Pass | All new Go code gets unit tests with vCenter mocked via `vcsim`; no live systems in tests |
| V — Module Independence | Pass | `vassessment/` has its own `go.mod`/`go.sum`; imports `pkg/common` by full module path; introduces no import cycle |
| VI — AI-Assisted Development | Pass | — |
| VII — Code Reuse and Simplicity | Pass, and central to this work | The reuse mandate is the reason the module lives in this repo (`research.md` R1) |
| VIII — Migration Field Parity | Not applicable | vAssessment does not touch the Migration Form |
| IX — UI Kubernetes Access | Not applicable in Phase 1 | The Phase 1 UI talks only to its own binary and never to a Kubernetes API. **Becomes applicable in Phase 3.** |
| X — Upgrade Flow Parity | Not applicable in Phase 1 | vAssessment is not part of the appliance and is not replaced by the appliance upgrade. **Becomes mandatory in Phase 3**, when it gains a Deployment and an image inside the appliance — at which point all five sites in the Principle X table must be wired. Recorded here so it is not discovered late. |
| XI — UI Test Coverage | ⚠️ **AMBIGUOUS — needs maintainer ruling** | See below |

### Conflict 1 — Principle I vs. SQLite storage

> *"All migration state must be represented as Kubernetes Custom Resources within k3s — no external state management allowed."*

vAssessment Phase 1 stores scans in a local SQLite file, with no Kubernetes cluster present at all. Read literally, this violates Principle I.

**The argument that it does not**: Principle I governs **migration state** — the Migration, MigrationPlan and related CRs that represent in-flight work the controller must reconcile. vAssessment holds **pre-migration assessment state**, orchestrates nothing, reconciles nothing, and by design runs *before* an appliance exists in the environment. Requiring a k3s cluster to hold its state would defeat the product's central premise. The product doc specifies SQLite explicitly.

**Requested resolution**: a maintainer ruling that Principle I is scoped to migration orchestration state, ideally recorded as a clarifying amendment. **This should be settled before implementation starts, not at review time.** If the ruling goes the other way, Phase 1's architecture changes fundamentally.

### Conflict 2 — Principle XI scope

Principle XI mandates Playwright for user-visible behaviour and Vitest for helpers, with file placement under `ui/e2e/` and `ui/src/`, and states that files placed elsewhere are silently never executed. Its obligations are written as applying to "any PR touching `ui/`".

vAssessment's frontend will live at `vassessment/ui/`, not `ui/`. So the letter of the principle does not reach it, but its intent plainly does.

**Recommendation**: adopt the same two-layer discipline inside `vassessment/ui/` with its own runners and its own CI job, rather than adding vAssessment specs to the appliance UI's Playwright project — mixing them would couple the two test suites and the two release trains. **Requesting confirmation** that a second, independent Playwright project in a different module does not count as "a third E2E runner" under Principle XI, which is aimed at preventing multiple runners within one application.

## Project Structure

### Documentation (this feature)

```text
vassessment/specs/001-vassessment-phase1/
├── spec.md                        # WHAT and WHY — user stories, requirements, success criteria
├── plan.md                        # This file — HOW
├── research.md                    # Findings: reuse, packaging, UI exposure
├── data-model.md                  # Entities and their relationships
├── quickstart.md                  # Build and run
├── contracts/
│   ├── rest-api.md                # HTTP contract, per UI page
│   └── pre-check-catalog.md       # The check catalog — rule IDs are a public contract
├── checklists/
│   └── requirements.md            # Pre-implementation quality gate
└── tasks.md                       # Ordered, dependency-aware work breakdown
```

### Source code

```text
vassessment/
├── cmd/vassessment/main.go        # Entry point (exists)
├── cli/                           # Cobra commands (exists: root, version, validate)
│   ├── serve.go                   # NEW — starts API + UI
│   └── scan.go                    # NEW — headless scan for CI/scripted use
├── preflight/                     # Offline validation (exists)
├── catalog/                       # Check catalog + evaluators (exists: 10 Basic rules)
│   ├── rules_basic.go             # Basic level (extends existing rules_src.go)
│   ├── rules_intermediate.go      # NEW — guest-property checks
│   └── exclusions.go              # NEW — vCLS / template / vCenter appliance
├── source/                        # NEW — source abstraction
│   ├── vcenter/                   #   live vCenter, via pkg/common
│   └── rvtools/                   #   xlsx/csv ingest
├── scan/                          # NEW — scan lifecycle, merge across sources
├── store/                         # NEW — SQLite persistence
├── plan/                          # NEW — copy methods, eligibility, estimates
├── report/                        # NEW — report generation
├── api/                           # NEW — REST handlers
├── ui/                            # NEW — frontend, built and go:embed-ed
└── specs/                         # This directory
```

**Structure Decision**: flat, purpose-named packages inside the existing module, matching the convention already established by `catalog/` and `preflight/`. The `source/` package is deliberately an abstraction over vCenter and RVTools so that FR-003 (N sources of either type) is structural rather than bolted on later.

### Shared-code changes (separate PRs, specified in `research.md` R1.5)

```text
pkg/common/vmware/
├── connect.go                     # C1 — govmomi connect/login/retry, no k8s
├── provider.go                    # C2 — CredentialProvider: Secret-backed | static
└── inventory.go                   # C4 — ContainerView + PropertyCollector bulk read
pkg/common/openstack/
└── connect.go                     # C3 — deferred to Phase 3
```

## Architecture

```
                    ┌──────────────────────────────────────────┐
  vCenter ──────────▶  source/vcenter  ┐                       │
  (read-only creds)  │  (pkg/common)   │                       │
                     │                 ├──▶ scan ──▶ store     │
  RVTools .xlsx ─────▶  source/rvtools ┘    (merge   (SQLite,  │
                     │                       by BIOS  immutable│
                     │                       UUID)    scans)   │
                     │                         │               │
                     │                         ▼               │
                     │                      catalog            │
                     │                 (rules → results)       │
                     │                         │               │
                     │            ┌────────────┼───────────┐   │
                     │            ▼            ▼           ▼   │
                     │          plan        report       api   │
                     │       (methods,    (scope doc,  (REST)  │
                     │       estimates)     CSV)          │    │
                     └─────────────────────────────────────┼───┘
                                                           ▼
                                                   ui (go:embed)
```

**Data flow invariant**: a scan is immutable once written. Method selections are a separate, mutable overlay keyed by scan and VM. Everything the API returns is derived from those two inputs plus the catalog version — which is what makes FR-019 (determinism) and SC-005 (identical reports) achievable.

## Phasing

| Phase | Scope | Issues |
|---|---|---|
| **1 — this spec** | Sources (vCenter + RVTools), scans, Basic + Intermediate checks, exclusions, inventory, copy methods and estimates, report, embedded UI, independent release | #2386 (done), #2387, #2388, + new |
| **2** | Advanced level via guest script and `guestinfo`; scan diff surfaced in UI | #2388 partially |
| **3** | Destination (PCD) checks; embedded mode inside the appliance (Ingress, Principle IX and X obligations) | — |
| **4** | Migration-bucket seeding into the Planner | — |

Phase 1 deliberately absorbs both Basic and Intermediate because both need only a read-only vCenter credential — the split between them is about what the guest reports, not about what vAssessment must be given.

## Release and packaging

Per `research.md` R2, entirely separate from `packer.yml`:

- `.github/workflows/vassessment.yml` — existing; build + test on every PR; publishes nothing
- `.github/workflows/vassessment-release.yml` — new; cross-compiles binaries, builds the container image and QCOW2, publishes to `quay.io/platform9/vassessment` and to S3 under a `vassessment/` prefix
- **Prerequisite actions outside the codebase**: create the `platform9/vassessment` Quay repository and grant the existing `QUAY_ROBOT_*` credentials push access; add a `vassessment/` entry to `cleanup-old-builds.yaml` so dev builds are pruned
- `latest` semantics mirror the appliance's exactly — `releases/latest` moves on GitHub release only, `nightly/latest` on the nightly build. This is a deliberate choice to match what support engineers already expect, and it is **not** inherited automatically (`research.md` R2.2)

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| Local SQLite state instead of Kubernetes CRs (Principle I) | The product must run before any appliance or cluster exists in the target environment — that is its entire reason to exist | Requiring k3s would make the tool unusable in exactly the pre-appliance scenario it is built for |
| A sixth Go module's worth of vCenter code | The reuse mandate is satisfied by extracting shared cores into `pkg/common` (C1, C2), not by importing k8s-coupled code | Importing `pkg/common/vmware/credentials.go` as-is is impossible: every function requires a `client.Client` |
| A second frontend and a second Playwright project | vAssessment ships independently of the appliance; sharing the appliance UI's test project would couple the release trains | Adding vAssessment pages into `ui/` would make independent release impossible |
