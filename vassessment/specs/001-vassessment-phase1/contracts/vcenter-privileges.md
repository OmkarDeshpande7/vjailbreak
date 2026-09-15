# Contract: Minimum vCenter Privilege Set

**Date**: 2026-09-15
**Satisfies**: FR-001, FR-006, FR-008, FR-016 · **Closes**: CHK026 · **Feeds**: `vassessment/preflight/role.go`

The customer must create a vCenter role for vAssessment. This document defines what that role needs and, just as importantly, what it must **not** have.

It is the source for the required-privileges JSON that `vassessment validate role` diffs against, which today deliberately ships empty because the privilege list could not be fixed before the properties being read were known. The catalog now enumerates those properties, so the list can be derived.

## Design principles

1. **Read-only, provably.** No privilege that permits a state change may appear in the role. This is not a convention — a privilege granting write access is a defect (FR-008, SC-002).
2. **Least privilege.** A customer security team will review this list. Every entry must be justifiable by a specific check.
3. **Missing privilege degrades, never fails.** If the role lacks something, the dependent check reports `not_evaluated` with reason `insufficient_privilege`, naming the privilege. The scan continues (FR-016).
4. **Detected at connect time, not mid-scan.** The privilege pre-flight runs during `POST /sources/vcenter` so the user learns immediately which checks will not run (FR-001).

---

## Base: the built-in Read-only role

vSphere's built-in **Read-only** role, granted at the root folder with propagation, carries `System.Anonymous`, `System.View` and `System.Read`. These are implicit in every role and cover reading the great majority of managed-object properties.

**Most of the Basic and Intermediate catalog is satisfied by Read-only alone.** The table below covers what it does and does not reach.

## Coverage by property group

Status: ✅ covered by Read-only · ➕ needs an explicit extra privilege · ❓ needs verification against a live vCenter before the list is finalised

| Property group | Example properties | Checks | Privilege | Status |
|---|---|---|---|---|
| VM configuration | `config.template`, `config.version`, `config.firmware`, `config.changeTrackingEnabled`, `config.annotation`, `config.guestId` | SRC-001, 003, 004, 007, 150–156, 177 | `System.Read` | ✅ |
| VM runtime | `runtime.powerState`, `runtime.host`, `runtime.faultToleranceState` | SRC-170, 173, 175 | `System.Read` | ✅ |
| VM hardware devices | `config.hardware.device[]` — disks, NICs, controllers, vTPM, USB, CD-ROM, serial | SRC-100–114, 140–143, 176 | `System.Read` | ✅ |
| Disk backing detail | `backing.diskMode`, `.sharing`, `.lunUuid`, `.compatibilityMode`, `.parent` | SRC-100, 105, 109, 110, 113 | `System.Read` | ✅ |
| Snapshots | `snapshot.rootSnapshotList[]`, `layoutEx` | SRC-101, 117, 118 | `System.Read` | ✅ |
| Guest properties | `guest.guestFullName`, `guest.disk[]`, `guest.net[]`, `guest.toolsRunningStatus` | SRC-005, 190–195 | `System.Read` | ✅ — these are properties of the VM object, **not** Guest Operations, which are not required |
| Host & cluster | `HostSystem.hardware.cpuInfo`, `config.product.version`, `ClusterComputeResource.summary` | SRC-132, 174, 175; Scope Summary | `System.Read` | ✅ |
| Datastore summary | `Datastore.summary.type/capacity/freeSpace`, `info.vmfs.version/blockSizeMb/extent` | SRC-116, 130, 134 | `System.Read` | ✅ |
| Network | portgroups, DVS, `OpaqueNetworkBackingInfo` | SRC-140–143 | `System.Read` | ✅ |
| Host storage devices | `host.config.storageDevice.scsiLun[]` — LUN UUID and vendor | SRC-110 shared RDM, SRC-134 array identification | `Host.Config.Storage` | ➕ ❓ — may be readable under Read-only; **verify** |
| Encryption state | `config.keyId`, disk `backing.keyId` | SRC-159, SRC-015 | `Cryptographer.ReadKeyServersInfo` | ➕ ❓ — reading the *flag* may need nothing extra; the docs cite `Cryptographer.Access`/`.Decrypt` for **migration**, which vAssessment does not do. **Verify the read-only case and do not request Decrypt.** |
| Storage policy | PBM `QueryAssociatedProfile`, vVol detection | SRC-115 | `StorageProfile.View` | ➕ |
| Performance counters | CPU Ready % | SRC-025 critical-workload candidate | `System.Read` | ❓ — perf counter retrieval is generally covered, but CPU Ready specifically needs confirmation. If it proves to need more, **drop the counter rather than widen the role** and evaluate SRC-025 on vCPU + memory alone |
| Session handling | login, session keep-alive | all | `Sessions.ValidateSession` | ❓ — typically implicit; verify |

## Explicitly excluded

These must never appear. Listed so a reviewer can confirm their absence at a glance:

| Privilege | Why excluded |
|---|---|
| `VirtualMachine.Interact.*` | Power operations — vAssessment never changes VM state |
| `VirtualMachine.Config.*` | Reconfiguration |
| `VirtualMachine.State.*` | Snapshot create/revert/remove — vAssessment reads the snapshot tree, never modifies it |
| `VirtualMachine.Provisioning.*` | Clone, deploy, disk access |
| `VirtualMachine.GuestOperations.*` | **Not needed.** Guest data comes from `guest.*` properties reported by Tools to vCenter, not from guest operations. Phase 2's `guestinfo` read also uses VM properties, not this privilege |
| `Cryptographer.Decrypt` / `.Access` | vAssessment reads whether a VM is encrypted, never its contents |
| `Datastore.FileManagement` / `.Browse` | vAssessment reads datastore summary properties, not files |
| `Host.Config.*` (beyond storage read) | No host configuration is read or written |
| `Global.*` | No global settings access |

## Role definition format

`vassessment validate role` diffs a granted role against a required one, both in this shape:

```json
{
  "name": "vassessment-readonly",
  "privileges": ["System.Anonymous", "System.View", "System.Read", "..."]
}
```

Two files ship with the product:

| File | Purpose |
|---|---|
| `vassessment-role-minimum.json` | The smallest role that runs Basic + Intermediate. Some checks will report `insufficient_privilege` |
| `vassessment-role-complete.json` | Minimum plus the ➕ extras, so no check is skipped for privilege reasons |

A customer applies whichever their security posture allows. Both are strictly read-only; the difference is coverage, never safety.

## Verification required before this list ships

The ❓ rows cannot be settled from documentation alone — vSphere privilege requirements for property reads are not exhaustively documented per-property, and behaviour varies by vCenter version. Before finalising:

1. Create a role with **only** `System.Anonymous`, `System.View`, `System.Read` on a test vCenter
2. Run a full scan and record every property that returns `NoPermission`
3. Add the smallest privilege that unblocks each, one at a time
4. Repeat against vCenter 6.7, 7.0 and 8.0 — the supported set per SRC-178

Whatever that produces is authoritative and supersedes the ❓ rows here. **The empirical result wins over this document**, which is a derivation from the catalog rather than a measurement.

## Connect-time pre-flight

`POST /sources/vcenter` reports three distinct outcomes, never collapsed into one "connection failed" (FR-001):

| Outcome | Meaning | User action |
|---|---|---|
| `unreachable` | Host does not resolve, or no network path | Fix DNS or firewall |
| `auth_failed` | Credentials rejected | Fix username or password |
| `insufficient_privilege` | Authenticated, but the role lacks privileges — **the response names which, and which checks will be skipped** | Extend the role, or proceed with reduced coverage |

`insufficient_privilege` is a warning, not a hard failure: the user may knowingly proceed with reduced coverage, and the scan records which checks were skipped so the report states its own limits honestly.
