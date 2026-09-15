# Requirements Checklist: vAssessment Phase 1

**Purpose**: Quality gate to clear before implementation starts. Items are questions about the *specification*, not the code — an unchecked item means we would be building on an unresolved assumption.
**Created**: 2026-09-15
**Feature**: [spec.md](../spec.md)

## Blocking — must be resolved before any implementation

- [ ] CHK001 Maintainer ruling obtained on **Constitution Principle I vs. local SQLite state** (`plan.md`, Conflict 1). A ruling against the proposed scoping changes Phase 1's architecture fundamentally
- [ ] CHK002 Maintainer ruling obtained on **Principle XI scope** — whether a second Playwright project under `vassessment/ui/` is permitted (`plan.md`, Conflict 2)
- [ ] CHK003 Decision recorded on what **`latest` means for vAssessment artifacts**. It is not inherited — no `:latest` container tag exists in this repo today (`research.md` R2.2)
- [ ] CHK004 `quay.io/platform9/vassessment` repository created and `QUAY_ROBOT_*` granted push access
- [ ] CHK005 Agreement that `pkg/common` changes C1 and C2 land as **separate PRs** before vAssessment's source layer depends on them

## Scope and requirement completeness

- [ ] CHK006 Every Basic and Intermediate check in `contracts/pre-check-catalog.md` is traceable to a functional requirement
- [ ] CHK007 Each of the four copy methods has stated eligibility rules and a modelled throughput
- [ ] CHK008 The exclusion set (vCLS, templates, vCenter appliance) is confirmed complete — is anything else auto-excluded?
- [ ] CHK009 The set of VM badges surfaced in the inventory (RDM, Old OS, Encrypted, …) is agreed and finite
- [ ] CHK010 Phase 1 / Phase 2 boundary confirmed: Advanced level and `guestinfo` are out
- [ ] CHK011 Confirmed that no Phase 1 requirement depends on PCD credentials

## Unresolved design questions

- [ ] CHK012 **What state are the remaining VMs in?** The design shows 120 compatible + 39 remediation VMs + 10 incompatible against an estate of 324. The other ~155 are unaccounted for, and the answer determines what `filter=` must support
- [ ] CHK013 **Which estate figures are canonical?** The report shows 6.8 days / 14.6 baseline / 7.8 saved; the inventory header shows 2.0 / 2.1 / 0.1. One estimator must own both
- [ ] CHK014 Is the agent count (4) user-adjustable, given it scales the headline number directly?
- [ ] CHK015 Terminology settled: the API uses `scan`, the design's header chip still says "snapshot"
- [ ] CHK016 Is `critical workload candidate` (≥2 of: ≥8 vCPU, active memory ≥16 GiB, CPU Ready ≥5%) a check, a badge, or both — and is CPU Ready actually retrievable at scan time without a performance-counter query?
- [ ] CHK017 Behaviour defined for a VM whose method selection becomes invalid after a rescan (e.g. powered off since, so live migration no longer applies)

## Data and correctness

- [ ] CHK018 BIOS UUID collision handling specified end to end — detection, reporting, and what the UI shows (FR-004)
- [ ] CHK019 Confirmed that no schema path allows `pass` on absent data (SC-004)
- [ ] CHK020 Every `not_evaluated` reason is enumerated and each maps to user-facing text
- [ ] CHK021 Grouping key is `(sourceId, datacenter, cluster)` everywhere, including the report's region table (FR-004)
- [ ] CHK022 Determinism defined precisely: which inputs must be equal for two reports to be byte-identical
- [ ] CHK023 Partial-scan semantics defined — what the UI shows when one of three sources failed

## Non-functional

- [ ] CHK024 The 10-minute / 1,000-VM target has a measurement method and a test estate
- [ ] CHK025 Read-only enforcement has a test strategy that can fail the build (SC-002)
- [ ] CHK026 Minimum vCenter privilege set enumerated and verified against every property the catalog reads
- [ ] CHK027 Behaviour specified when the role lacks a privilege mid-scan (FR-016)
- [ ] CHK028 Air-gapped operation confirmed — no telemetry, no font/CDN fetch from the embedded UI (NFR-003)
- [ ] CHK029 Credential handling reviewed: never persisted, never logged, never passed as a process argument

## Delivery

- [ ] CHK030 `vassessment-release.yml` scope agreed: which artifacts, which triggers, which S3 prefixes
- [ ] CHK031 `cleanup-old-builds.yaml` given a `vassessment/` prefix entry so dev builds are pruned
- [ ] CHK032 Confirmed a vAssessment release requires no vJailbreak appliance build (SC-009)
- [ ] CHK033 Phase 3 embedded-mode obligations recorded against Principles IX and X so they are not discovered late

## Notes

- Check items off as completed: `[x]`
- CHK001–CHK005 are hard blockers; the rest can proceed in parallel with early implementation
- CHK012 and CHK013 are questions about the design mockups themselves and need the designer, not an engineer
