# Tasks: vAssessment Phase 1

**Input**: Design documents in `vassessment/specs/001-vassessment-phase1/`
**Prerequisites**: `spec.md`, `plan.md`, `research.md`, `data-model.md`, `contracts/`

**Format**: `[ID] [P?] [Story] Description`
**[P]** = parallelisable (different files, no dependency on an unfinished task)
**[Story]** = the user story from `spec.md` this serves

**Tests are mandatory here**, not optional — constitution principle IV requires unit tests for all new Go code with external systems mocked.

**Scope note**: the current PR delivers **specs and directory structure only**. Everything below is subsequent work.

---

## Phase 0 — Blockers (no implementation until these clear)

- [ ] T001 Obtain maintainer ruling on Principle I vs. local SQLite state (`plan.md` Conflict 1, CHK001)
- [ ] T002 Obtain maintainer ruling on Principle XI scope for `vassessment/ui/` (`plan.md` Conflict 2, CHK002)
- [ ] T003 Resolve the two design inconsistencies with the designer: unaccounted VM states (CHK012) and irreconcilable estate figures (CHK013)
- [ ] T004 Create the `quay.io/platform9/vassessment` repository and grant `QUAY_ROBOT_*` push access (CHK004)

---

## Phase 1 — Shared-code extraction (`pkg/common`, separate PRs)

These land **before** vAssessment's source layer, and each is behaviour-preserving with existing callers unchanged.

- [ ] T005 [P] Extract govmomi connect/login/retry from `pkg/common/validation/vmware/validate.go:123-195` into `pkg/common/vmware/connect.go`, accepting a plain credentials struct (`research.md` C1)
- [ ] T006 Add `CredentialProvider` interface in `pkg/common/vmware` with Secret-backed and static implementations; repoint `Validate()` at it (C2)
- [ ] T007 [P] Unit tests for T005/T006 against `vcsim`, plus a regression test that existing appliance callers are unaffected
- [ ] T008 [P] Run both `pkg/common` and `k8s/migration` test suites to prove no appliance regression (`research.md` R2.4)

---

## Phase 2 — Foundation (blocks every user story)

- [ ] T009 [P] Add dependencies: `govmomi`, `modernc.org/sqlite`, xlsx reader; `go mod tidy`
- [ ] T010 Define core entities in `vassessment/store/models.go` per `data-model.md`
- [ ] T011 SQLite schema and migrations in `vassessment/store/` — immutable scans, separate `method_selection` overlay keyed by assessment + durable VM identity
- [ ] T012 [P] Unit tests for the store, including the invariant that `pass` cannot be written with empty evidence (SC-004)
- [ ] T013 Scan lifecycle in `vassessment/scan/` — create, progress, `complete` vs `partial`, per-source outcomes
- [ ] T014 `Source` abstraction in `vassessment/source/` with a single interface over vCenter and RVTools

---

## Phase 3 — User Story 1: assess from a vCenter credential (P1) 🎯 MVP

- [ ] T015 [US1] vCenter source in `vassessment/source/vcenter/`, using `pkg/common/vmware` connect + session cache
- [ ] T016 [US1] Privilege pre-flight per `contracts/vcenter-privileges.md` — distinguish unreachable / auth failed / insufficient privilege, naming the missing privileges and the checks they disable (FR-001)
- [ ] T016a [US1] Settle the privilege contract empirically: scan with a bare Read-only role, record every `NoPermission`, add the smallest unblocking privilege, repeat on vCenter 6.7/7.0/8.0. Ship `vassessment-role-minimum.json` and `vassessment-role-complete.json` (CHK026)
- [ ] T017 [US1] Bulk inventory read: `ContainerView` + one `property.Collector.Retrieve` over the required property set (`research.md` C4; NFR-001)
- [ ] T018 [P] [US1] Map raw `mo.VirtualMachine` to the `VM` entity, including disks, NICs, vTPM, USB, encryption — **all new code, no precedent in repo** (`research.md` R1.6)
- [ ] T019 [P] [US1] Infrastructure collection: datacenters, clusters, hosts (cores), datastores (type, backing array), networks
- [ ] T020 [US1] Exclusion rules in `vassessment/catalog/exclusions.go` — vCLS, templates, vCenter appliance via vSource matching (FR-011)
- [ ] T021 [US1] Extend `catalog/` with the remaining Basic checks per `contracts/pre-check-catalog.md`
- [ ] T022 [US1] Intermediate checks in `catalog/rules_intermediate.go` reading guest properties, each degrading to `not_evaluated` with reason when Tools is absent (FR-015, FR-016)
- [ ] T023 [US1] Compatibility roll-up: check results → `compatible` / `remediation` / `incompatible` / `excluded` (FR-017)
- [ ] T024 [P] [US1] Table-driven unit tests for every check, covering pass, fire, and not-evaluated paths
- [ ] T025 [US1] Read-only assertion test — run a full scan against `vcsim` and fail if any mutating method is invoked (SC-002)
- [ ] T026 [US1] Performance test against a 1,000-VM `vcsim` estate, asserting < 10 min (NFR-001)
- [ ] T027 [US1] `GET /overview` and `GET /scans/{id}` per `contracts/rest-api.md`
- [ ] T028 [US1] `POST /sources/vcenter`, `POST /scans:refresh`, `GET /sources`
- [ ] T029 [US1] `vassessment serve` and `vassessment scan` commands; credentials via stdin/env only, never a process argument

**Checkpoint**: a read-only credential yields a complete, evidenced assessment. Independently shippable.

---

## Phase 4 — User Story 2: RVTools ingest (P2)

- [ ] T030 [P] [US2] RVTools parser in `vassessment/source/rvtools/` per `contracts/rvtools-mapping.md`; reject an upload missing a required tab (vInfo, vDisk, vHost, vSource)
- [ ] T031 [US2] Column-to-property mapping per that contract, matching columns by name never by position, with absent columns producing `not_evaluated` — never `pass` (FR-002, FR-016)
- [ ] T032 [US2] Report parse warnings naming every missing tab/column so the UI can say which checks are unevaluable
- [ ] T033 [P] [US2] Tests with a real export, a partial export, and a malformed file
- [ ] T034 [US2] `POST /sources/rvtools`

---

## Phase 5 — User Story 6: multi-source (P3 UI, P1 data model)

Sequenced early despite its P3 label — the data model cannot absorb it later without migrating stored scans.

- [ ] T035 [US6] Cross-source merge on BIOS UUID, live data winning over RVTools on conflict (FR-004)
- [ ] T036 [US6] BIOS UUID collision detection and reporting — never a silent merge (FR-004, CHK018)
- [ ] T037 [US6] `(sourceId, datacenter, cluster)` grouping with source disambiguation on name collision (CHK021)
- [ ] T038 [US6] Per-source failure isolation and `partial` scan semantics (FR-007)
- [ ] T039 [P] [US6] Tests: two vCenters; same vCenter via both paths; colliding datacenter names; one source failing
- [ ] T040 [US6] `POST /sources/{id}:rescan`, `DELETE /sources/{id}`, `?source=` filter on list endpoints

---

## Phase 6 — User Story 3: inventory and evidence (P2)

- [ ] T041 [US3] `GET /inventory` with server-side filter, search, grouping and pagination (FR-032)
- [ ] T042 [US3] `GET /inventory/{vmId}` detail with per-level check results and evidence
- [ ] T043 [US3] `GET /pre-checks` catalog endpoint
- [ ] T044 [P] [US3] Contract tests 1–3 from `contracts/rest-api.md`

---

## Phase 7 — User Story 4: copy methods and estimates (P3)

- [ ] T045 [US4] Method eligibility engine in `vassessment/plan/` — returns allowed methods plus a reason per disallowed one (FR-022)
- [ ] T046 [US4] Single canonical estimator: per-VM copy and downtime, estate window, cold baseline, saving (FR-023, FR-024)
- [ ] T047 [US4] Method selection persistence keyed by assessment + durable VM identity, surviving rescans (FR-026)
- [ ] T048 [US4] Invalidation handling when a rescan makes a stored selection ineligible (CHK017)
- [ ] T049 [US4] `PATCH /inventory/{vmId}/method`, returning recomputed estate estimates in the same response
- [ ] T050 [P] [US4] Tests: eligibility matrix, estimate arithmetic, selection survival, invalid-method rejection

---

## Phase 8 — User Story 5: report (P3)

- [ ] T051 [US5] Report data assembly in `vassessment/report/` — scope by region, exclusions, levels, compatibility, top remediation, edge cases, per-VM plan, copy methods, methodology
- [ ] T052 [US5] PDF renderer matching the approved layout
- [ ] T053 [P] [US5] Full-estate CSV export, untruncated (FR-029)
- [ ] T054 [US5] `POST /report`, `GET /report/{id}`, `GET /report/data`
- [ ] T055 [US5] Determinism test — identical bytes for identical inputs (SC-005)
- [ ] T056 [US5] Reconciliation test — report figures equal UI figures for the same scan (FR-030, SC-005)

---

## Phase 9 — UI

- [ ] T057 Scaffold `vassessment/ui/` matching the appliance UI's stack, with a **configurable base path** from the first commit (`plan.md`)
- [ ] T058 [P] Add Source page — connect vCenter, upload RVTools
- [ ] T059 [P] Overview page — levels, cards, edge cases, top remediation
- [ ] T060 [P] Sources page — list, add, rescan, remove
- [ ] T061 Inventory page — grouped list, filters, VM detail, method dropdown, per-row rescan
- [ ] T062 [P] Pre-checks page — catalog grouped by level
- [ ] T063 Generate Report action in the shared header
- [ ] T064 `go:embed` the built assets into the binary (FR-031)
- [ ] T065 [P] Unit tests for UI helpers; E2E specs for each page, backend stubbed — pending the T002 ruling
- [ ] T066 Verify the UI performs no classification or aggregate arithmetic (SC-007)

---

## Phase 10 — Release

- [ ] T067 `.github/workflows/vassessment-release.yml` — cross-compile linux/darwin × amd64/arm64, container image, QCOW2
- [ ] T068 Publish to `quay.io/platform9/vassessment` (image + ORAS artifact) and S3 under a `vassessment/` prefix
- [ ] T069 Implement the agreed `latest` semantics (T003/CHK003) — not inherited (`research.md` R2.2)
- [ ] T070 [P] Add a `vassessment/` prefix entry to `cleanup-old-builds.yaml`
- [ ] T071 [P] vAssessment version file, independent of the appliance's
- [ ] T072 Prove SC-009: cut a vAssessment release with no appliance artifact built

---

## Dependencies

```
T001-T004 (blockers)
    └─▶ T005-T008 (pkg/common)
            └─▶ T009-T014 (foundation)
                    ├─▶ T015-T029  US1  ── MVP checkpoint
                    │       ├─▶ T030-T034  US2
                    │       ├─▶ T035-T040  US6  (early: data model)
                    │       ├─▶ T041-T044  US3
                    │       │       └─▶ T045-T050  US4
                    │       │               └─▶ T051-T056  US5
                    │       └─▶ T057-T066  UI  (needs its endpoints per page)
                    └─▶ T067-T072  Release (parallel with UI)
```

## Parallelisation notes

- T005/T007/T008 (`pkg/common`) run parallel to T009–T012 (store) — different modules
- T018/T019 (VM mapping, infrastructure) are independent files
- Within Phase 9 most pages are independent; T061 (Inventory) is the largest and depends on T045–T049
- Release work (Phase 10) needs no UI and can run alongside it

## Suggested MVP cut

**T001–T029** delivers User Story 1 end to end: connect a read-only credential, scan, evidenced verdicts, headless output. That alone is demonstrable to a customer and is the point at which the product's core claim is proven.
