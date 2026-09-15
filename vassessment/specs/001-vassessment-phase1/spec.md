# Feature Specification: vAssessment Phase 1 — Credentialed Assessment with Basic UI

**Feature Branch**: `2386-set-up-the-vassessment-project-folder-structure-and-basic-cli` (foundation), successors `2387`, `2388`
**Created**: 2026-09-15
**Status**: Draft
**Product doc**: [vAssessment — Discovery & Migration Assessment for vJailbreak](https://platform9.atlassian.net/wiki/spaces/vJailbreak/pages/6262587411)
**API contract**: [vAssessment APIs (redesign)](https://docs.google.com/document/d/1aF91tcAo3Qe6zQ3nb5ttIIckNGc9MoLQdfW9cgy94ZE/edit)
**Design**: vAssessment redesign canvas (4 pages) + generated "VMware to PCD Migration Scope" report

## Overview

vAssessment answers the question every migration engagement opens with — *what do you have, how much of it can vJailbreak migrate, with what risk, and in what order?* — **before** a migration is attempted and **before** a vJailbreak appliance exists in the customer environment.

Today that question is answered with RVTools exports and a hand-maintained spreadsheet. The result is slow, inconsistent between engineers, and silently incomplete: conditions that break a migration halfway through are frequently not discovered until the migration is already running at a customer site.

Phase 1 delivers a **single binary that serves its own web UI**, connects to vCenter with read-only credentials (or ingests an RVTools export), inventories the estate, evaluates every check that is derivable from vCenter alone, and produces a shareable migration-scope report. It ships as its own appliance image on its own release cadence, independent of the vJailbreak release train.

### Scope boundary

Phase 1 covers the two assessment levels that require **no in-guest agent and no destination credentials**:

| Level | Data source | In Phase 1? |
|---|---|---|
| **Basic** | vCenter API config properties, or RVTools export | Yes |
| **Intermediate** | vCenter API guest properties (requires VMware Tools running in the guest, but no script) | Yes |
| **Advanced** | Guest script results delivered via `guestinfo` | No — Phase 2 |
| **Destination (PCD)** | PCD/OpenStack API | No — Phase 3 |

Basic and Intermediate are both reachable with nothing but a read-only vCenter credential, which is why they form one coherent shippable phase.

## User Scenarios & Testing

### User Story 1 — Assess an estate from a read-only vCenter credential (Priority: P1)

A Platform9 SE, on a call with a prospect, is given a read-only vCenter credential. They enter the host, username, and password, click **Connect and Scan**, and within ten minutes have a complete inventory of the estate with every VM classified as compatible, needing remediation, or incompatible — computed on the customer's own VMs, not a generic slide.

**Why this priority**: This is the entire product. Everything else is an alternative input path, a presentation of this result, or a later enrichment of it. If only this story ships, the tool is already useful in a sales conversation and as the opening artifact of an engagement.

**Independent Test**: Point the binary at a vCenter with a read-only account, run a scan, and confirm a per-VM verdict with supporting evidence for every VM in the estate, with zero mutating API calls issued.

**Acceptance Scenarios**:

1. **Given** valid read-only vCenter credentials, **When** the user submits the connect form, **Then** the credential is validated, a source is created, and a scan begins
2. **Given** invalid credentials, **When** the user submits the form, **Then** a specific error is shown (unreachable host vs. authentication failure vs. insufficient privilege) and no source is created
3. **Given** a scan of a 1,000-VM estate, **When** the scan runs, **Then** it completes in under 10 minutes
4. **Given** a completed scan, **When** the user opens the Overview, **Then** compatible / remediation / incompatible counts, data-to-move, and the excluded-VM breakdown are shown
5. **Given** a completed scan, **When** an auditor reviews the vCenter task and event log for the scan window, **Then** no mutating operation attributable to vAssessment appears

---

### User Story 2 — Assess from an RVTools export when no credential is available yet (Priority: P2)

A partner engineer cannot get vCenter credentials issued in time, but can produce an RVTools export. They upload the workbook and get the Basic-level assessment immediately, with every check that needs live data clearly reported as *not evaluated* rather than silently passed.

**Why this priority**: Credential issuance is the single most common scheduling blocker in an engagement. This path keeps the assessment moving, and it is how most of today's manual assessments already start. It is second only to the live path because it yields a strictly smaller result.

**Independent Test**: Upload an RVTools `.xlsx` with no vCenter configured; confirm Basic checks evaluate, Intermediate checks report `not_evaluated`, and no check is reported as passing on absent data.

**Acceptance Scenarios**:

1. **Given** an RVTools `.xlsx` / `.xlsm` workbook or a `.zip` of CSV tabs, **When** the user uploads it, **Then** vInfo, vHost, vDisk, vUSB and vSource are parsed and a source is created
2. **Given** an export missing a tab or column a check depends on, **When** the assessment runs, **Then** that check reports `not_evaluated` with the reason, and never `pass`
3. **Given** an RVTools-only source, **When** the user views Sources, **Then** its available levels are shown as Basic only
4. **Given** a malformed or non-RVTools file, **When** uploaded, **Then** the upload is rejected with a specific parse error and no partial source is created

---

### User Story 3 — Browse the inventory and understand each VM's verdict (Priority: P2)

An engineer reviewing the assessment opens Inventory, filters to the VMs needing remediation, and for any VM can see exactly which checks fired, what evidence produced that result, and what to do about it.

**Why this priority**: A verdict without evidence is not actionable and will not be trusted by a customer's infrastructure team, who will and should challenge it. Equal priority with Story 2: both are required for the result to be usable in an engagement.

**Independent Test**: With a completed scan, filter the inventory by each verdict class and open a VM's detail; confirm every check contributing to the verdict is listed with its evidence value.

**Acceptance Scenarios**:

1. **Given** a completed scan, **When** the user opens Inventory, **Then** VMs are grouped by datacenter/cluster with per-group VM count, powered-on count, and total size
2. **Given** the inventory list, **When** the user selects the Compatible / Remediation / Incompatible filter, **Then** the list re-queries server-side and the result count updates
3. **Given** a VM row, **When** the user opens its detail, **Then** each check is listed under its level with status and the evidence value that produced it
4. **Given** a VM whose Intermediate checks could not run because VMware Tools is not running, **Then** those checks show `not_evaluated` with that reason, not a failure

---

### User Story 4 — Choose copy methods and see the migration window change (Priority: P3)

A solutions architect planning the engagement changes a VM's copy method from Standard to Storage-Accelerated and immediately sees the estate's estimated migration window and the savings against the cold baseline recalculate.

**Why this priority**: This is what converts an assessment into a plan, and the headline "X days, Y days saved" number is the most quoted output of the whole tool. It is P3 because it is only meaningful once the inventory and verdicts beneath it are correct.

**Independent Test**: Change one VM's method and confirm both that VM's copy/downtime estimate and the three estate-level header figures update consistently, and that the choice survives a rescan.

**Acceptance Scenarios**:

1. **Given** a VM, **When** the user opens the method dropdown, **Then** only methods valid for that VM are selectable, and invalid ones show why they are unavailable
2. **Given** a method change, **When** it is saved, **Then** the per-VM copy time, per-VM downtime, and the estate totals all recalculate
3. **Given** a powered-off VM, **When** the method list is shown, **Then** live migration is not selectable
4. **Given** a VM on a datastore with no certified array backing, **When** the method list is shown, **Then** Storage-Accelerated Copy is not selectable
5. **Given** a method selection, **When** the source is rescanned, **Then** the selection is preserved

---

### User Story 5 — Produce a shareable scope report (Priority: P3)

At the end of the session the engineer clicks **Generate Report** and gets a self-contained document they can leave with the customer: scope by region, what was excluded and why, assessment coverage, edge cases with vJailbreak's handling of each, the per-VM plan, and the methodology behind every number.

**Why this priority**: The report is the engagement deliverable and the artifact that outlives the session. P3 because it is a rendering of results that Stories 1–4 must produce first.

**Independent Test**: Generate a report from a completed scan and confirm every figure in it reconciles with what the UI displays for the same scan.

**Acceptance Scenarios**:

1. **Given** a completed scan, **When** the user generates a report, **Then** it contains scope summary, exclusions, assessment levels, compatibility counts, top remediation items, edge cases, per-VM plan, copy methods, and methodology
2. **Given** the same scan and the same method selections, **When** the report is generated twice, **Then** the two outputs are identical
3. **Given** a generated report, **When** any headline figure is compared against the UI for that scan, **Then** they agree
4. **Given** a large estate, **When** a CSV export is requested, **Then** it contains every VM, not the sampled subset shown in the document

---

### User Story 6 — Assess multiple vCenters as one estate (Priority: P3)

A customer with two vCenters adds both. The Overview, Inventory and report describe the combined estate, while each source's own connection state and coverage remain visible.

**Why this priority**: Multi-vCenter estates are common in the target customer profile, and the data model cannot be retrofitted later without migrating stored scans and rewriting every aggregate. P3 for UI polish, but its data-model consequences are P1 and are treated as such in the plan.

**Independent Test**: Add two sources, confirm all aggregate views sum across both, and confirm one source failing does not invalidate the other's results.

**Acceptance Scenarios**:

1. **Given** two vCenter sources, **When** the user views Overview or Inventory, **Then** figures aggregate across both
2. **Given** an RVTools export and a live connection for the *same* vCenter, **When** both are added, **Then** their VMs are merged rather than duplicated, and live data wins on conflict
3. **Given** two sources with identically named datacenters, **When** the inventory is grouped, **Then** the two groups remain distinct and are disambiguated by source
4. **Given** one unreachable vCenter, **When** a refresh runs, **Then** that source is marked failed and the other source's results remain valid and visible

---

### Edge Cases

- **A VM exists in two sources with the same BIOS UUID** — the design's stated merge key. BIOS UUID is settable and survives cloning, so it is not guaranteed unique. Duplicates must be detected and surfaced, never silently collapsed into one VM or double-counted.
- **VMware Tools is not running** — every Intermediate check for that VM is `not_evaluated` with that reason. It must never be reported as a pass, and must not be reported as a failure of the VM.
- **A VM is powered off** — it stays in scope (vJailbreak migrates it cold), but live migration methods are unavailable to it.
- **A VM is orphaned or inaccessible in vCenter** — it cannot be assessed and must be reported as such rather than silently dropped from counts.
- **The estate contains vCLS VMs, templates, and the vCenter appliance itself** — these are excluded automatically and reported in an exclusions breakdown. Excluded is a distinct state from incompatible; excluded VMs are not failures.
- **A scan is interrupted** (network loss, credential expiry mid-scan) — the partial scan must not be presented as complete.
- **The read-only role lacks a privilege a check needs** — that check reports `not_evaluated` naming the missing privilege, rather than failing the whole scan.
- **An RVTools export is stale relative to a live connection for the same vCenter** — live data must win, and the staleness must be visible.
- **Two vCenters have colliding datacenter names** — grouping must not merge them.
- **A rescan changes a VM's verdict** — the change must be attributable to specific checks, not just a different headline number.

## Requirements

### Functional Requirements

**Sources and credentials**

- **FR-001**: System MUST accept a vCenter source specified by host, username and password, and MUST validate reachability, authentication and privilege sufficiency as distinct, separately-reported outcomes
- **FR-002**: System MUST accept an RVTools export (`.xlsx`, `.xlsm`, or `.zip` of CSV tabs) as an alternative source, parsing at minimum the vInfo, vHost, vDisk, vUSB and vSource tabs
- **FR-003**: System MUST support N sources per assessment, of either type, added at any time
- **FR-004**: System MUST merge VMs across sources on BIOS UUID, preferring live vCenter data over RVTools data on conflict, and MUST report any BIOS UUID collision rather than silently merging
- **FR-005**: System MUST report per-source connection status and which assessment levels that source can supply
- **FR-006**: System MUST NOT persist credentials into scans, reports, or exports
- **FR-007**: System MUST isolate source failures — one unreachable source MUST NOT invalidate other sources' results

**Discovery and scanning**

- **FR-008**: System MUST issue only read operations against vCenter. No mutating vSphere call may be made under any code path
- **FR-009**: System MUST persist each scan as an immutable point-in-time record, identified and timestamped
- **FR-010**: System MUST support rescanning all sources, a single source, or a single VM
- **FR-011**: System MUST automatically exclude vCLS VMs, templates, and the vCenter appliance from the migration scope, and MUST report the count and reason for each exclusion class
- **FR-012**: System MUST keep powered-off VMs in scope
- **FR-013**: System MUST report scan progress while a scan is running

**Assessment**

- **FR-014**: System MUST evaluate every Basic-level check from vCenter configuration properties or the equivalent RVTools columns
- **FR-015**: System MUST evaluate every Intermediate-level check from vCenter guest properties where VMware Tools reports them
- **FR-016**: System MUST report a check as `not_evaluated` with a reason whenever its input data is unavailable, and MUST NOT report it as passing
- **FR-017**: System MUST classify each VM as compatible, remediation, incompatible, or excluded, derived deterministically from its check results
- **FR-018**: System MUST retain, for every check result, the evidence value that produced it
- **FR-019**: System MUST produce identical results for identical inputs — same scan data plus same catalog version yields the same verdicts
- **FR-020**: System MUST expose the check catalog itself — every check, its level, what it inspects, and what its firing means
- **FR-021**: System MUST version the check catalog and record the catalog version used for each scan

**Planning**

- **FR-022**: System MUST offer a copy method per VM, restricted to the methods valid for that VM, with a reason given for each unavailable method
- **FR-023**: System MUST estimate per-VM copy duration and downtime from the selected method
- **FR-024**: System MUST estimate the estate migration window, a Standard-Cold baseline, and the saving between them
- **FR-025**: System MUST recompute estate estimates when any method selection changes
- **FR-026**: System MUST preserve method selections across rescans
- **FR-027**: System MUST state the modelled rates and assumptions behind every estimate

**Output**

- **FR-028**: System MUST generate a self-contained scope report covering scope by region, exclusions, assessment coverage, compatibility counts, top remediation items, edge cases, per-VM plan, copy methods, and methodology
- **FR-029**: System MUST provide a full-estate export that is not truncated to the sample shown in the document
- **FR-030**: Report figures MUST reconcile with the UI for the same scan

**Interface and delivery**

- **FR-031**: System MUST ship as a single binary that serves both the API and the UI
- **FR-032**: UI MUST render server-computed values without performing its own classification, percentage, duration, or aggregate arithmetic
- **FR-033**: System MUST be operable end to end without a Kubernetes cluster
- **FR-034**: System MUST build and release independently of the vJailbreak appliance release

### Non-Functional Requirements

- **NFR-001**: A 1,000-VM estate MUST complete discovery in under 10 minutes
- **NFR-002**: The binary MUST run on linux and darwin, amd64 and arm64
- **NFR-003**: System MUST make no outbound network calls other than to the configured sources — it must be usable in an air-gapped environment
- **NFR-004**: Scan storage MUST stay at or below approximately 50 MB for a 5,000-VM estate
- **NFR-005**: Reports and exports MUST be self-contained and viewable offline
- **NFR-006**: Read-only enforcement MUST be verifiable by test, not only by convention
- **NFR-007**: All new Go code MUST have unit tests with external systems mocked, per constitution principle IV
- **NFR-008**: The `vassessment` module MUST maintain its own dependency graph and MUST NOT introduce an import cycle with other modules, per constitution principle V

### Key Entities

- **Source** — one vCenter connection or one RVTools export. Carries connection state and the set of assessment levels it can supply.
- **Scan** — an immutable point-in-time capture across all sources, with a catalog version.
- **VM** — a virtual machine as discovered, merged across sources, carrying identity, configuration, guest data where available, exclusion state, and compatibility classification.
- **Check** — a catalog entry: identity, level, what it inspects, what its firing means, and remediation guidance.
- **CheckResult** — one check evaluated against one VM in one scan: status, evidence, and reason when not evaluated.
- **CopyMethod** — an available migration transport, its eligibility rules, and its modelled throughput.
- **MethodSelection** — a user's per-VM method choice; mutable state layered over an immutable scan.
- **Estimate** — derived copy duration, downtime, estate window, baseline, and saving.
- **Report** — a generated, deterministic rendering of one scan plus its method selections.

## Success Criteria

### Measurable Outcomes

- **SC-001**: An engineer with only a read-only vCenter credential produces a complete, shareable assessment of a 1,000-VM estate in under 30 minutes end to end, of which under 10 minutes is scanning
- **SC-002**: Zero mutating vSphere API calls are issued, demonstrated by an automated test asserting the read-only property against a simulated vCenter
- **SC-003**: Every Basic and Intermediate check in the catalog is implemented, unit-tested, and reports evidence
- **SC-004**: No check can report `pass` on absent input data — enforced by test across the whole catalog
- **SC-005**: Two runs over the same scan data and catalog version produce byte-identical reports
- **SC-006**: An estate spanning two vCenters, including one added as RVTools and one live, produces correct merged totals with no double-counting
- **SC-007**: The UI performs no verdict classification or aggregate arithmetic — verified by the absence of such logic in the frontend
- **SC-008**: The binary runs and completes an assessment on a host with no Kubernetes cluster present
- **SC-009**: A release of vAssessment can be cut without building or releasing any vJailbreak appliance artifact
- **SC-010**: Every check that vJailbreak enforces at migration time and that is detectable from vCenter data is also detectable at assessment time

## Assumptions

- The customer can issue a read-only vCenter role; the minimum privilege set is documented and shipped with the product
- Intermediate-level data is available only for VMs where VMware Tools is running; partial coverage is expected and is reported, not treated as an error
- vJailbreak's own compatibility rules are the source of truth for what "compatible" means; where the codebase and the product doc disagree, the codebase behaviour wins and the discrepancy is raised
- Copy-method throughput rates are modelled planning figures, not measured guarantees, and are labelled as such in all output
- Method selection is user state layered over an immutable scan, not part of the scan itself
- The UI is deliberately minimal in Phase 1 — correctness and completeness of the assessment take precedence over visual polish
- Existing vJailbreak code is reused wherever it is not Kubernetes-coupled; where it is coupled, the reusable core is extracted rather than duplicated (see `research.md`)
- Advanced-level checks, destination (PCD) checks, scan diffing, the XLSX round-trip, and migration-bucket seeding are explicitly out of scope for Phase 1 and are specified separately
