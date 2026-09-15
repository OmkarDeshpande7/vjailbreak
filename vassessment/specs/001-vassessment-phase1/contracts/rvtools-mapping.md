# Contract: RVTools Export Mapping

**Date**: 2026-09-15
**Satisfies**: FR-002, FR-014, FR-016 · **Implements**: T030, T031, T032

An RVTools export is the alternative source for customers who cannot get a vCenter credential issued in time (User Story 2). This document defines exactly which tabs and columns map to which internal properties, and therefore exactly which checks can and cannot be evaluated from an export.

**The governing rule**: a column that is absent produces `not_evaluated` with a reason. It never produces `pass` (FR-016, SC-004). An export from an older RVTools version, or one where the operator deselected tabs, must degrade visibly rather than silently.

---

## Tab coverage

| Tab | Supplies | Required? |
|---|---|---|
| `vInfo` | Core VM configuration — the backbone of Basic level | **Required** |
| `vDisk` | Per-disk backing, mode, sharing, controller | **Required** for ~30 disk checks |
| `vHost` | Physical hosts, cores, ESXi version | **Required** for the report's Scope Summary |
| `vSource` | Identifies which vCenter produced the export | **Required** for source attribution and vCenter-appliance exclusion |
| `vUSB` | USB device presence | Recommended |
| `vNetwork` | Per-NIC network, MAC, IPv4 | Recommended — see "Intermediate from RVTools" below |
| `vPartition` | Guest filesystem capacity and free space | Recommended — same |
| `vTools` | VMware Tools status and version | Recommended — same |
| `vSnapshot` | Snapshot tree | Recommended |
| `vDatastore` | Datastore type, capacity, free | Recommended |

`vInfo`, `vDisk`, `vHost` and `vSource` are hard requirements; an export missing any of them is rejected at upload with a specific error rather than ingested partially.

---

## ⚠️ Finding: RVTools can supply more than the design assumes

The design states that an RVTools export yields "Basic checks and the Tools-reported guest OS", and that **disk free space needs a live vCenter connection**. That is not correct for a full export.

RVTools collects `vPartition` (guest filesystem free space), `vTools` (Tools status) and `vNetwork` (per-NIC IPs) directly. Those are precisely the inputs for the Intermediate level:

| Intermediate check | Needs | Available in RVTools? |
|---|---|---|
| SRC-190 root / `/boot` free space | `guest.disk[].freeSpace` | **Yes** — `vPartition.Free MiB` |
| SRC-191 Tools running status | `guest.toolsRunningStatus` | **Yes** — `vTools.Tools status` |
| SRC-192/193/194 IP and MAC checks | `guest.net[]` | **Yes** — `vNetwork.IPv4 Address`, `Mac Address` |
| SRC-005 guest OS vs matrix | `guest.guestFullName` | **Yes** — `vInfo.OS according to the VMware Tools` |

**Recommendation**: parse these three extra tabs and let an RVTools source report `levelsAvailable: [basic, intermediate]` when they are present. This materially improves the credential-less path — it is the difference between an export answering ~60 checks and answering ~72.

**Caveat that must be surfaced in the UI**: RVTools data is a point-in-time snapshot taken when the operator ran the export. Free space in particular drifts. An Intermediate result sourced from an export must be labelled with the export's timestamp, and live vCenter data must win on merge (FR-004).

This contradicts the design's Sources-page copy, which needs updating if the recommendation is accepted.

---

## `vInfo` → VM properties

| RVTools column | Internal property | Feeds |
|---|---|---|
| `VM` | `name` | identity |
| `VM UUID` | `biosUuid` | **cross-source merge key** (FR-004) |
| `VM ID` | `moref` | intra-source identity |
| `Powerstate` | `powerState` | SRC-170, method eligibility |
| `Template` | `isTemplate` | SRC-001, SRC-177, exclusions |
| `HW version` | `hardwareVersion` | SRC-003 |
| `CBT` | `cbtEnabled` | SRC-004 |
| `Firmware` | `firmware` | SRC-119, SRC-154, SRC-157 |
| `CPUs` | `vcpu` | sizing, SRC-025 |
| `Memory` | `memoryMiB` | sizing |
| `NICs` / `Disks` | counts | cross-check against `vNetwork` / `vDisk` row counts |
| `OS according to the configuration file` | `guestOsConfigured` | **SRC-150, SRC-151** — the powered-off-safe OS source |
| `OS according to the VMware Tools` | `guestOsRunning` | SRC-005 |
| `Annotation` | `annotation` | SRC-007, SRC-111 |
| `Folder`, `Resource pool`, `Cluster`, `Datacenter`, `Host` | inventory path | grouping, Scope Summary |
| `EnableUUID` | `diskEnableUuid` | SRC-114 |
| `VI SDK Server` | source vCenter | **vCenter-appliance exclusion** (FR-011) |

**Not available in `vInfo`**, and therefore `not_evaluated` from an export alone: vTPM device presence, `config.keyId` encryption flag, `efiSecureBootEnabled`, `vbsEnabled`, `faultToleranceState`. These need the live API — RVTools does not export them.

## `vDisk` → Disk properties

| RVTools column | Internal property | Feeds |
|---|---|---|
| `Disk` | `key` / label | SRC-111 (label match is case-sensitive) |
| `Capacity MiB` | `sizeGB` | sizing, estimates — note SRC-120's truncation concern |
| `Raw` | RDM flag | SRC-108, SRC-110, SRC-112 |
| `Raw LUN ID` | `lunUuid` | SRC-110 shared-RDM ownership |
| `Raw Comp. Mode` | RDM compatibility mode | **SRC-109** physical vs. virtual |
| `Disk Mode` | `diskMode` | **SRC-100** independent mode |
| `Sharing` | `sharing` | **SRC-105** multi-writer |
| `Controller` | `controllerType` | SRC-104 shared-bus SCSI, SRC-176 |
| `Thin` / `Eagerly Scrub` | provisioning | informational |
| `Path` / `Datastore` | backing location | SRC-115, SRC-130, SRC-134, SRC-136 |

`vDisk` is what makes the long tail of disk checks reachable from an export. Without it, roughly 30 checks in the catalog are unevaluable — which is why it is a hard requirement.

**Not available**: SeSparse vs. FlatVer2 concrete backing type (SRC-102, SRC-106), delta-chain parent links (SRC-113), `backing.uuid` (SRC-114 detects only via `vInfo.EnableUUID`).

## `vHost`, `vSource`, and the remaining tabs

| Tab | Column | Internal property | Feeds |
|---|---|---|---|
| `vHost` | `Host`, `Datacenter`, `Cluster` | host inventory | Scope Summary regions |
| `vHost` | `# Cores`, `# CPU`, `Cores per CPU` | `cpuCores` | **Scope Summary physical cores** |
| `vHost` | `ESX Version` | `esxiVersion` | SRC-132 proxy requirement |
| `vSource` | `VI SDK Server` | source vCenter FQDN | source attribution, appliance exclusion |
| `vUSB` | `VM`, `Device Type` | `usbDevicePresent` | SRC-176 |
| `vSnapshot` | `VM`, `Name`, `Date` | `snapshotCount`, names | SRC-117, SRC-118 |
| `vDatastore` | `Name`, `Type`, `Capacity`, `Free` | datastore entity | SRC-116, SRC-134 |
| `vPartition` | `VM`, `Disk`, `Capacity MiB`, `Free MiB` | `guest.disk[]` | **SRC-190** |
| `vTools` | `VM`, `Tools status` | `toolsStatus` | **SRC-191**, gates all Intermediate |
| `vNetwork` | `VM`, `Network`, `Mac Address`, `IPv4 Address` | `guest.net[]` | SRC-192, SRC-193, SRC-194 |

---

## Degradation behaviour

Three distinct failure modes, deliberately distinguished so the user knows which one they have:

| Situation | Behaviour |
|---|---|
| A **required tab** is missing (`vInfo`, `vDisk`, `vHost`, `vSource`) | Reject the upload with a specific error. No partial source is created (FR-002 acceptance scenario 4) |
| An **optional tab** is missing | Ingest. Every check depending on it reports `not_evaluated` with reason `data_unavailable`, naming the tab. Surface in `warnings[]` on upload so the UI can state upfront which checks will not run |
| A **column** within a present tab is missing (older RVTools version) | Same as above, at column granularity |

The upload response's `warnings[]` (see `rest-api.md` §1) is what carries this to the user. A silent partial ingest is the specific failure this contract exists to prevent.

## Version sensitivity

RVTools column names have changed across releases — columns have been added and occasionally renamed. The parser must:

1. Match columns by name, tolerating case and surrounding whitespace differences
2. Treat an unrecognised column as ignorable, never fatal
3. Record the RVTools version from the export where available, and include it in the scan record
4. **Never** match columns positionally — column order is not stable across versions

**Open item**: the exact set of RVTools versions to support has not been agreed. The mapping above reflects contemporary RVTools 4.x exports and needs verification against the oldest version encountered in the field before implementation (see `checklists/requirements.md`).
