# Contract: vAssessment REST API

**Date**: 2026-09-15
**Mirrors**: [vAssessment APIs (redesign)](https://docs.google.com/document/d/1aF91tcAo3Qe6zQ3nb5ttIIckNGc9MoLQdfW9cgy94ZE/edit) — that doc is the shared artifact with the UI team; this file is the in-repo copy of record and wins on conflict.

## Conventions

- Base path `/api/v1/assessment`. The binary serves this and the UI from the same port.
- **Thin client.** Every percentage, label, duration string, badge state and aggregate is computed server-side. The UI performs no classification and no arithmetic (FR-032, SC-007).
- `?scan={id}` selects a scan; omitted means the latest.
- Status vocabulary everywhere: `pass` · `remediation` · `fail` · `not_evaluated`.
- Compatibility vocabulary: `compatible` · `remediation` · `incompatible` · `excluded`.
- Every list response aggregates across sources by default and accepts `?source={id}` (FR-003).
- Errors: `{error: {code, message, detail?}}` with conventional HTTP status. Credential failures distinguish `unreachable`, `auth_failed`, and `insufficient_privilege` as separate codes (FR-001).

## Phase scope

| Endpoint group | Phase |
|---|---|
| Sources, Scans, Overview, Inventory, Pre-checks, Report | **1** |
| Guest script, Advanced level ingestion | 2 |
| Destination correlation, PCD sources | 3 |
| Bucket seeding | 4 |

---

## 1. Add Source (empty state)

| Call | Body / returns |
|---|---|
| `POST /sources/vcenter` | In `{host, username, password, insecure?}`. Validates shape, tests the live connection, creates the source, starts the first scan — the "Connect and Scan" button, one round trip. Out `{sourceId, scanId, status}`. On failure `{error:{code}}` where code ∈ `unreachable` \| `auth_failed` \| `insufficient_privilege`, with the missing privileges named. |
| `POST /sources/rvtools` | multipart: `.xlsx` / `.xlsm` / `.zip` of CSV tabs, one file per vCenter. Parses vInfo, vHost, vDisk, vUSB, vSource. Out `{sourceId, scanId, vmCount, levelsAvailable, excluded:{vcls, templates, vcenterAppliance, total}, warnings[]}`. `warnings` names every tab or column absent from the export, so the UI can say which checks will be unevaluable. |
| `GET /scans/{scanId}` | Poll: `{status, progressPct, vmsDiscovered, stage, sourceOutcomes}`. |

---

## 2. Overview

| Call | Returns |
|---|---|
| `GET /overview?scan={id}` | One call renders the page. |

```
scan     {id, label:"scan #3", capturedAtLabel:"10:42", version, catalogVersion, status}

levels   [ {level:1, key:"basic", name:"Basic", pct:100, donutDashArray,
            status:"complete", statusLabel:"Complete", statusTone:"green",
            source:"vCenter API", description, checkChips:[...],
            vmsEvaluated:324, vmsTotal:324,
            excluded:25, excludedLabel:"Excluded (vCLS, templates, vCenter)"} , ... ]
          // the selected level's detail panel is the same object — no second call

cards    {compatibleVMs:{value, caption},
          remediationItems:{value, vmCount, caption},
          incompatibleVMs:{value, caption},
          dataToMove:{valueTB, caption, storageAcceleratedPct}}

supportedEdgeCases [ {count, title, description,
                      handling:"supported|review|incompatible|plan", handlingLabel} ]

topRemediationItems [ {checkId, label, count, pctOfMax} ]   // pctOfMax = bar width
```

`levels[].pct` and `vmsEvaluated` are **weighted across sources** — an RVTools-only source can never contribute to Intermediate, so a per-source breakdown must be available behind the headline figure or it misleads (`research.md` R1, FR-005).

---

## 3. Sources

| Call | Returns |
|---|---|
| `GET /sources` | `[{id, host, username, type, typeLabel, vmCount, levelsAvailable, levelsLabel, lastScanAt, lastScanLabel, status, statusLabel, statusTone, statusDetail?}]` |
| `POST /sources/vcenter` · `POST /sources/rvtools` | As §1 — "+ Add Source" reuses them |
| `POST /sources/{id}:rescan` | vCenter sources only. Out `{scanId}` |
| `POST /scans:refresh` | Header "Refresh" — rescans every rescannable source into one new scan. Out `{scanId}` |
| `DELETE /sources/{id}` | Removes the source and its contribution |

---

## 4. Inventory

| Call | Returns |
|---|---|
| `GET /inventory?scan=&filter=all\|compatible\|remediation\|incompatible&advancedChecks=&q=&source=&page=&perPage=` | Filtering is server-side (FR-032) |

```
estimates  {totalMigration:{days, label}, baseline:{days, label},
            saved:{days, label}, dataToMove:{tb, label},
            storageAcceleratedPct, agents, caption}

groups     [ {key:"src-1/DC-Mumbai/Prod-A",       // (sourceId, datacenter, cluster)
              label:"DC-Mumbai / Prod-A", sourceLabel?, // shown only on collision
              vmCount, poweredOnCount, totalGB, caption, vms:[VM]} ]

VM         {id, name, path, badges:["RDM"|"Old OS"|"Encrypted"|...],
            guestOS, sizeLabel, diskGB, vcpu, poweredOn,
            compatibility, excluded, exclusionReason?,
            checks:{basic:{status,label}, intermediate:{...}, advanced:{...}},
            method:{selected, selectedLabel,
                    options:[{value, label, allowed, reason?}],
                    note?},
            copyLabel, downtimeLabel,
            biosUuidCollision?:{with:[vmId]}}

pagination {page, perPage, total, shownLabel, pageLabel}
```

- Incompatible or excluded VMs return `method:null` and `copyLabel`/`downtimeLabel` of `"—"`.
- **Disallowed methods are returned with `allowed:false` and a reason** so the UI greys them out without encoding eligibility rules.

| Call | Returns |
|---|---|
| `PATCH /inventory/{vmId}/method` | In `{method}`. The only routine write. Out `{vm:{method, copyLabel, downtimeLabel, note}, estimates:{...}}` — **recomputed estate estimates come back in the same response** so the header cards update in one round trip. Rejects a method that is not `allowed` for that VM. |
| `POST /inventory/{vmId}:rescan` | Per-row rescan icon. Out updated `VM` + `estimates`. |
| `GET /inventory/{vmId}` | Detail: `{header:{...}, levels:{basic:[Check], intermediate:[], advanced:[]}}`, `Check = {ruleId, name, status, evidence, notEvaluatedReason?, remediation?}` |

---

## 5. Pre-checks

| Call | Returns |
|---|---|
| `GET /pre-checks?scan={id}` | `{catalogVersion, matrixVersion, total, groups:[{key, label, count, description, checks:[{id, name, whatItChecks, ifItFires, severity, affectedCount?}]}]}` |

Groups: Basic · Intermediate · Advanced · Manual · Destination. Reference data — cacheable, and text search may stay client-side since no computation is involved. `affectedCount` is present only when a scan is supplied.

---

## 6. Guest Script — Phase 2

Recorded for completeness; **not implemented in Phase 1**.

There are **no upload endpoints**. The redesign delivers results through `guestinfo`: the script writes findings into the VM's guestinfo via VMware Tools, and vAssessment reads them back from vCenter on the next rescan. The previous design's bulk-upload and BIOS-UUID file matching are gone.

| Call | Returns |
|---|---|
| `GET /guest-script` | `{optional:true, notice, steps:[...], downloads:[{os, filename, description, url}]}` |
| `GET /guest-script/{os}/download` | The `.sh` / `.ps1` file |

Advanced-result ingestion is a side effect of `:rescan`, not its own endpoint. **This mechanism has no precedent in the repo and its read path is unverified** (`research.md` R1.6, R4).

---

## 7. Report

| Call | Returns |
|---|---|
| `POST /report?scan={id}` | In `{format:"pdf"\|"csv"\|"xlsx"}`. Out `{reportId, status}` |
| `GET /report/{reportId}` | Poll, then download |
| `GET /report/data?scan={id}` | The JSON the report renders from — an in-app preview uses it, and it is the contract for the document's sections |

```
scopeSummary  {regions:[{name, sourceLabel?, clusters, physicalHosts,
                         physicalCores, vmsToMigrate}], totals, note}
exclusions    {rows:[{reason, count}], totalExcluded, totalToMigrate, poweredOffNote}
assessmentLevels, compatibilityCards, topRemediationItems, edgeCases   // as §2
migrationPlan [ {name, cluster, guestOS, size, basic, intermediate, advanced,
                 method, copy, downtime} ]
copyMethods   [ {name, description, vddk, migration} ]
methodology   {rates:{standardMBs:120, acceleratedMBs:160, storageAcceleratedMBs:800},
               perVmOverheadMin:15, agents:4, exclusionsNote,
               criticalWorkloadRule, largeVmRule,
               disclaimer:"Planning numbers, not a quote"}
```

`csv` returns the **full estate**, not the sample shown in the document (FR-029). Every figure here must reconcile with the UI for the same scan (FR-030).

---

## Contract tests

These are the assertions the API must satisfy, and they map directly to spec requirements:

1. No endpoint returns a `pass` status where the underlying evidence field is empty (SC-004)
2. `GET /report/data` twice for the same scan and selections returns byte-identical JSON (SC-005)
3. Figures in `/overview`, `/inventory` and `/report/data` for the same scan agree (FR-030)
4. `PATCH .../method` with a method whose `allowed` is false returns 4xx and changes nothing
5. A method selection survives `POST /sources/{id}:rescan` (FR-026)
6. With two sources, every aggregate equals the sum of per-source figures, with merged VMs counted once (FR-004, SC-006)
7. A failed source leaves other sources' data readable and marks the scan `partial` (FR-007)
8. No code path issues a mutating vSphere call — asserted against `vcsim` (SC-002)
