# Data Model: vAssessment Phase 1

**Date**: 2026-09-15
**Feeds**: `spec.md` (Key Entities), `contracts/rest-api.md`

## Governing invariants

These four rules drive most of the design below. Every schema choice traces back to one of them.

1. **A scan is immutable.** Once written it is never updated. A rescan produces a new scan. This is what makes reports reproducible (FR-019, SC-005) and scans diffable in Phase 2.
2. **Method selections are not part of a scan.** They are user intent that must survive rescans (FR-026), so they are keyed by durable VM identity at the assessment level, not by scan.
3. **Absence is a first-class result.** A check that could not be evaluated is `not_evaluated` with a reason. There is no schema path by which missing input data can be recorded as a pass (FR-016, SC-004).
4. **Everything is multi-source from the start.** Every VM and every aggregate carries source attribution. Retrofitting this is a data migration, so it is structural from the first commit (FR-003).

---

## Entity overview

```
Assessment ──1:N──▶ Source ──1:N──▶ Scan
     │                                 │
     │                                 └──1:N──▶ VM ──1:N──▶ CheckResult
     │                                            │            │
     │                                            ├──1:N──▶ Disk
     │                                            └──1:N──▶ NIC
     │
     └──1:N──▶ MethodSelection  (keyed by VM identity, survives rescans)

Catalog ──1:N──▶ Check ──────────────────────────┘ (CheckResult references Check.id)
```

---

## Assessment

The top-level container: one engagement, one SQLite file.

| Field | Type | Notes |
|---|---|---|
| `id` | string | |
| `name` | string | Customer or engagement name |
| `createdAt` | timestamp | |
| `catalogVersion` | string | Catalog version in force; recorded per scan as well |

---

## Source

One vCenter connection or one RVTools export. Credentials are **never persisted** (FR-006) — they are held in memory for the process lifetime and re-entered on restart.

| Field | Type | Notes |
|---|---|---|
| `id` | string | |
| `type` | enum | `vcenter_api` \| `rvtools` |
| `host` | string | vCenter FQDN, or the vCenter identified inside the RVTools export |
| `username` | string | Display only; no password stored |
| `status` | enum | `connected` \| `failed` \| `never_scanned` |
| `statusDetail` | string | Failure reason when `failed` |
| `levelsAvailable` | []enum | `basic` \| `intermediate` \| `advanced` — an RVTools source is `[basic]` only |
| `lastScanAt` | timestamp | |
| `fileName` | string | RVTools sources only |

**Invariant**: source failure is isolated (FR-007). A `failed` source contributes whatever data its last successful scan produced and does not invalidate the assessment.

---

## Scan

An immutable point-in-time capture across all sources.

| Field | Type | Notes |
|---|---|---|
| `id` | string | |
| `sequence` | int | Human-facing counter — "scan #3" |
| `startedAt` / `completedAt` | timestamp | |
| `status` | enum | `running` \| `complete` \| `partial` \| `failed` |
| `catalogVersion` | string | Pinned at scan time so old scans re-render identically |
| `sourceIds` | []string | Sources that contributed |
| `sourceOutcomes` | map | Per-source success/failure for this scan |

**`partial` matters**: if one source of three fails, the scan is `partial`, not `complete`. An interrupted scan must never be presented as whole (Edge Cases).

---

## VM

A virtual machine after cross-source merge.

### Identity

Three identifiers, because no single one is sufficient:

| Field | Uniqueness | Use |
|---|---|---|
| `moref` | Stable within one vCenter | Addressing the VM for rescan |
| `instanceUuid` | Unique within one vCenter | Intra-source identity |
| `biosUuid` | **Not guaranteed unique** | Cross-source merge key, per the design |

`biosUuid` is settable and survives cloning, so collisions are real. The model records `biosUuidCollision: bool` and the colliding set rather than merging silently (FR-004). The internal primary key is a synthetic `vmId`; `(sourceId, moref)` is the natural key within a source.

### Core fields

| Field | Type | Source |
|---|---|---|
| `name`, `datacenter`, `cluster`, `host`, `folder` | string | vCenter inventory path |
| `powerState` | enum | `poweredOn` \| `poweredOff` \| `suspended` |
| `isTemplate` | bool | `config.template` |
| `hardwareVersion` | int | from `config.version` (`vmx-NN`) |
| `firmware` | enum | `bios` \| `efi` |
| `secureBoot` | bool | |
| `cbtEnabled` | *bool | nil = unknown, which is distinct from false |
| `vcpu`, `memoryMiB` | int | |
| `annotation` | string | |
| `snapshotCount` | int | |
| `vtpmPresent`, `usbDevicePresent` | bool | from `config.hardware.device` |
| `encryptionKeyId` | string | empty = not encrypted |
| `guestOsConfigured` | string | `config.guestId` — always available |
| `guestOsRunning` | string | `guest.guestFullName` — Tools only |
| `toolsStatus` | enum | Gates every Intermediate check |
| `sourceIds` | []string | Which sources contributed; >1 means merged |

### Exclusion

| Field | Type | Notes |
|---|---|---|
| `excluded` | bool | |
| `exclusionReason` | enum | `vcls` \| `template` \| `vcenter_appliance` |

**Excluded is not incompatible.** Excluded VMs are removed from migration scope and counted separately; they are not failures and must not appear in incompatible counts (FR-011).

### Classification

| Field | Type | Notes |
|---|---|---|
| `compatibility` | enum | `compatible` \| `remediation` \| `incompatible` \| `excluded` |
| `levelStatus` | map[level]enum | Per-level roll-up: `pass` \| `remediation` \| `fail` \| `not_evaluated` |

Derived purely from `CheckResult`s — never stored independently of them, so a verdict can always be explained by the checks beneath it (FR-018).

---

## Disk and NIC

Separate entities because several checks are per-device, and totals are aggregates over them.

**Disk**: `key`, `sizeGB`, `backingType` (`flat` \| `rdm_physical` \| `rdm_virtual` \| `vvol` \| `sparse`), `diskMode` (independent/persistent modes matter — see the catalog), `sharing`, `datastore`, `datastoreType`, `controllerType`, `thinProvisioned`.

**NIC**: `key`, `network`, `networkType` (`standard` \| `dvs` \| `opaque`), `macAddress`, `ipAddresses` (Tools only), `connected`.

---

## Check (catalog entry)

Reference data, versioned with the catalog. Rule IDs are a **public contract** — they appear in reports customers keep, so they are never reused or renumbered.

| Field | Type | Notes |
|---|---|---|
| `id` | string | `SRC-001`, `GST-003`, `DST-002`, … |
| `level` | enum | `basic` \| `intermediate` \| `advanced` \| `manual` \| `destination` |
| `name`, `whatItChecks`, `ifItFires` | string | Displayed verbatim on the Pre-checks page |
| `severity` | enum | `incompatible` \| `remediation` \| `informational` |
| `dataSource` | string | The property path it reads — e.g. `config.changeTrackingEnabled` |
| `remediation` | string | |

---

## CheckResult

One check against one VM in one scan.

| Field | Type | Notes |
|---|---|---|
| `scanId`, `vmId`, `checkId` | string | Composite key |
| `status` | enum | `pass` \| `remediation` \| `fail` \| `not_evaluated` |
| `evidence` | string | The observed value — `changeTrackingEnabled=false` |
| `notEvaluatedReason` | enum | `tools_not_running` \| `insufficient_privilege` \| `data_unavailable` \| `level_not_run` |

**Schema-level enforcement of invariant 3**: `status = pass` requires non-empty `evidence`. A check with no input cannot produce evidence, and therefore cannot be recorded as a pass.

---

## CopyMethod and MethodSelection

**CopyMethod** is static reference data: `id` (`standard_live`, `standard_cold`, `vjailbreak_accelerated`, `storage_accelerated`), `name`, `vddkRequired`, `migrationType`, `modelledThroughputMBs`, and its eligibility rules.

**Eligibility is computed server-side per VM** and returned with a reason for each unavailable option (FR-022) — the UI never encodes these rules:

| Method | Requires |
|---|---|
| `standard_live` | Powered on, CBT enabled |
| `standard_cold` | — (universal fallback) |
| `vjailbreak_accelerated` | A registered proxy VM; cold only |
| `storage_accelerated` | Datastore backed by a certified array; cold only |

**MethodSelection** — the mutable overlay:

| Field | Type | Notes |
|---|---|---|
| `assessmentId` | string | **Not `scanId`** |
| `vmIdentity` | string | Durable identity, not the per-scan `vmId` |
| `method` | enum | |
| `selectedAt` | timestamp | |

**This is the key modelling decision on the write path.** Keying selections by scan would silently discard every choice on rescan, violating FR-026. Keying by durable VM identity at the assessment level means a rescan re-applies prior intent, and a selection that has become invalid (a VM powered off since, so live migration no longer applies) is reported as invalidated rather than silently changed.

---

## Estimate

Derived, never stored as truth — recomputed from scan data plus current selections, so it cannot drift from its inputs.

Per VM: `copyMinutes`, `downtimeMinutes`, and the method they assume.
Per estate: `totalMigrationDays`, `baselineColdDays`, `savedDays`, `dataToMoveTB`, `storageAcceleratedPct`, `agents`.

**One canonical estimator produces every figure.** The design currently shows irreconcilable numbers in two places (`research.md` R4); a single estimator with a reconciliation test is the fix (FR-030, SC-005).

Modelled inputs, stated in all output: Standard 120 MB/s, vJailbreak Accelerated 160 MB/s, Storage-Accelerated ~800 MB/s, 15 min per-VM overhead, 4 parallel agents.

---

## Infrastructure entities

Needed for the report's Scope Summary (physical hosts, cores per region) and for storage-backed method eligibility.

**Datacenter** (`name`, `sourceId`) → treated as a *region* in the report. **Cluster** (`name`, `datacenter`, `hostCount`). **Host** (`name`, `cluster`, `cpuCores`, `esxiVersion`). **Datastore** (`name`, `type`, `capacityGB`, `freeGB`, `backingArray`, `arrayCertified`). **Network** (`name`, `type`, `vlanId`).

**Grouping key is `(sourceId, datacenter, cluster)`**, never `(datacenter, cluster)`. Datacenter names are not unique across vCenters, and the report's "one region per vSphere Datacenter" rule would otherwise silently merge two customers' datacenters that happen to share a name (Edge Cases, FR-004).

---

## Storage sketch

```sql
assessment(id, name, created_at, catalog_version)
source(id, assessment_id, type, host, username, status, status_detail,
       levels_available, last_scan_at, file_name)
scan(id, assessment_id, sequence, started_at, completed_at, status,
     catalog_version, source_outcomes)
vm(id, scan_id, source_id, moref, instance_uuid, bios_uuid,
   bios_uuid_collision, name, datacenter, cluster, host, power_state,
   is_template, hardware_version, firmware, cbt_enabled, vcpu, memory_mib,
   annotation, snapshot_count, vtpm_present, encryption_key_id,
   guest_os_configured, guest_os_running, tools_status,
   excluded, exclusion_reason, compatibility)
disk(id, vm_id, key, size_gb, backing_type, disk_mode, sharing,
     datastore, controller_type)
nic(id, vm_id, key, network, network_type, mac_address, connected)
check_result(scan_id, vm_id, check_id, status, evidence, not_evaluated_reason,
             PRIMARY KEY (scan_id, vm_id, check_id))
method_selection(assessment_id, vm_identity, method, selected_at,
                 PRIMARY KEY (assessment_id, vm_identity))
datacenter(...) cluster(...) host(...) datastore(...) network(...)
```

Indexes on `vm(scan_id, compatibility)` and `check_result(scan_id, check_id)` — the inventory filter and the "N VMs affected by check X" rollup are the two hot queries.

The catalog itself is **compiled into the binary**, not stored — it is code (`catalog/` package, evaluators are pure Go functions), and `catalogVersion` on each scan records which version produced the results.
