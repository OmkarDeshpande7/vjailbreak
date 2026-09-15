# Contract: vAssessment Pre-Check Catalog

**Date**: 2026-09-15
**Status**: Draft — candidate catalog for review
**Sources**: the product doc's baseline catalog, plus a systematic audit of vJailbreak's own error paths, guard clauses and known-limitation workarounds, plus the published docs

## Why this document exists

The product doc supplies a baseline of ~41 checks. That baseline was written from field experience; it does not know what the migration code actually fails on.

This catalog adds what the codebase knows. Every error return, guard clause and OS-specific workaround in `v2v-helper/` and `k8s/migration/` is evidence of a real migration that broke. **Roughly 60 of the checks below were derived that way and are not in the baseline.** Each one is a migration that would otherwise fail at a customer site, often after the entire disk copy has already run.

That is the point of the exercise: *measure everything we can from credentials alone, so nothing we already know about surprises a customer.*

## ID policy

**Rule IDs are a public contract.** They appear in reports customers keep and in support conversations. They are never reused, never renumbered, and never change meaning.

| Range | Meaning |
|---|---|
| `SRC-001`–`SRC-027` | Baseline source-side checks from the product doc |
| `SRC-100`+ | **New** — derived from the codebase audit |
| `GST-xxx` | Advanced level — requires the in-guest script (Phase 2) |
| `MAN-xxx` | Manual — human review |
| `DST-xxx` | Destination — requires PCD credentials (Phase 3) |

## Levels

| Level | Data source | Phase |
|---|---|---|
| **Basic** | vCenter configuration properties, or RVTools columns | 1 |
| **Intermediate** | vCenter guest properties — needs VMware Tools running, but no script | 1 |
| **Advanced** | In-guest script results via `guestinfo` | 2 |
| **Destination** | PCD / OpenStack API | 3 |

**Severity**: ❌ incompatible · ⚠ remediation · ℹ informational

---

## ⚠ Prerequisite: the `VM` entity has no disk fields

`vassessment/catalog/vm.go` as it stands today carries no disk, controller, or backing information at all. **Roughly 30 of the checks below cannot be written until `data-model.md`'s `Disk` entity exists.** This is the single largest blocker to catalog completeness and is why `tasks.md` sequences the VM/disk mapping (T018) before the catalog work.

---

## Basic level — disk, controller and backing

The richest seam. Almost none of this is checked anywhere today, and several failures are silent.

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-100 | **Independent disk mode** — VDDK ≥7 cannot open `independent_*` disks | `…(VirtualDisk).backing.diskMode` | ❌ | `v2v-helper/pkg/utils/nbd_errors.go:64-71` |
| SRC-101 | Independent disks are excluded from snapshots, so the device key is absent at copy time | `snapshot.config.hardware.device[].key` | ❌ | `v2v-helper/vm/vmops.go:450` |
| SRC-102 | **SeSparse backing** (VMFS-6 snapshots, linked clones) is handled nowhere | `…backing.(VirtualDiskSeSparseBackingInfo)` | ❌ | no non-vendor handler; falls to `credutils.go:979` "unsupported disk backing type" |
| SRC-103 | **`VirtualDiskRawDiskVer2BackingInfo` crashes the worker** — unchecked type assertion to a file backing | `…backing` concrete type | ❌ | `v2v-helper/vm/vmops.go:408`, `:498` |
| SRC-104 | **Shared-bus SCSI (MSCS/WSFC)** — a comment claims this is checked; the code does not check it | `…(VirtualSCSIController).sharedBus` | ❌ | `credutils.go:950` comment vs `:951-984` body |
| SRC-105 | **Multi-writer disks** — a peer VM keeps writing during the copy, producing a torn image | `…backing.sharing == sharingMultiWriter` | ❌ | no handler anywhere |
| SRC-106 | **Hot-Add disk-index mismatch** — the transfer list skips non-FlatVer2 disks but is zipped by index against a list that does not | FlatVer2 count vs total non-RDM disks | ❌ *silent* | `migrate/hotadd_copy.go:119-122` vs `:578-582` |
| SRC-107 | Mixed RDM and non-RDM disks mis-index change IDs on incremental sync | RDM presence + disk count | ❌ *silent* | `vm/vmops.go:509-515` |
| SRC-108 | **An RDM-only VM disappears from inventory** with only a log line — no CR, no event, no UI row | RDM backing with zero flat disks | ❌ | `credutils.go:1921-1923` |
| SRC-109 | RDM **physical vs virtual** compatibility mode is never read; physical mode is unsupported | `…(RawDiskMappingVer1BackingInfo).compatibilityMode` | ❌ | `rdmdisk_types.go:22-31`; docs `network-storage-mapping.md:41` |
| SRC-110 | Shared RDM — every VM owning the same LUN must be in one plan | `…backing.lunUuid` × `host.…scsiLun[].uuid` | ❌ | `credutils.go:968-975` |
| SRC-111 | RDM requires a `VJB_RDM:` annotation, and the disk label match is case-sensitive | `config.annotation`, `…deviceInfo.label` | ⚠ | docs `migrating_rdm_disk_..._cli.md:63-71` |
| SRC-112 | Any RDM implies PCD ≥2025.7, cold migration only, retry disabled | any RDM backing | ❌ | docs `migrating_rdm_disk_..._cli.md:11,189` |
| SRC-113 | **Linked clone / delta chain** — Hot-Add copies one file from the chain, not the chain | `layoutEx.disk[].chain[]`, `…backing.parent` | ⚠ | `hotadd_copy.go:123-127` |
| SRC-114 | `disk.EnableUUID=false` leaves no `backing.uuid`, so Hot-Add cannot match block devices | `…backing.uuid`, `config.extraConfig` | ⚠ | `hotadd_copy.go:313`, `:293` |
| SRC-115 | Non-default **storage policy or vVol** backing must be reset to Datastore Default | PBM profile; `Datastore.summary.type` | ❌ | docs `vtpm_migration.md:46-52`, `network-storage-mapping.md:39` |
| SRC-116 | **Source datastore free space** is never checked before snapshotting (docs want ≥20%) | `Datastore.summary.freeSpace/capacity` | ⚠ | `vmops.go:259` fetches only `{"name"}` |
| SRC-117 | Pre-existing **user snapshots** block the copy; cleanup only removes vJailbreak's own | `snapshot.rootSnapshotList[]` | ⚠ | `vmops.go:1126` |
| SRC-118 | Stale `vjailbreak-*` snapshots from a prior attempt | `snapshot.rootSnapshotList[].name` | ⚠ | docs `vjailbreak-accelerated-copy.md:439` |
| SRC-119 | UEFI **multi-disk** — ESP on a different disk than root affects boot index | `config.firmware`, disk count | ℹ | `conversion.go:162-163`, `439-441` |
| SRC-120 | Disk capacity truncation — discovery uses `capacityInKB/1024/1024` while the copy path uses `capacityInBytes` with `math.Ceil` | both properties | ℹ | `credutils.go:1901` vs `openstack/clients.go:196` |
| SRC-121 | The **largest single disk**, not total size, drives the copy window | per-disk `capacityInKB` | ℹ | docs `prerequisites.md:112,120-122` |

## Basic level — copy-method eligibility

These determine which copy methods a VM may use, and therefore feed method eligibility directly (FR-022).

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-130 | Accelerated Copy requires matching **VMFS version and block size** between source and proxy datastore | `…backing.datastore` → `info.vmfs.version`, `blockSizeMb` | ❌ | docs `vjailbreak-accelerated-copy.md:50` |
| SRC-131 | **≤60 disks per proxy VM**, shared across concurrent migrations — the constant exists but is never enforced | Σ planned disks per proxy | ⚠ | `constants.go:856-857` (zero usages); fails at `hotadd_copy.go:230` |
| SRC-132 | Proxy OVA is vmx-21, requiring **ESXi/vCenter 8.0 U2+** | `HostSystem.config.product.version` | ❌ | docs `vjailbreak-accelerated-copy.md:93` |
| SRC-133 | Proxy VM must be Linux, same vCenter, PVSCSI on controller 0, Tools running | proxy hardware, `guest.toolsRunningStatus`, `about.instanceUuid` | ❌ | `proxyvm_controller.go:135` |
| SRC-134 | **XCOPY is VMFS-only** — NFS and vSAN unsupported; datastore must be a single NAA extent on a certified array | `Datastore.summary.type`, `info.vmfs.extent[]` → `scsiLun[].vendor` | ❌ | docs `storage-accelerated-copy.md:416-420` |
| SRC-135 | XCOPY assumes **512-byte sectors**; 4Kn LUNs are mis-sized | `capacityInBytes`, datastore block size | ❌ | `migrate/vaai_copy.go:184-187` |
| SRC-136 | XCOPY needs array credentials mapped for every source datastore — today this fails *after the VM is powered off* | `…backing.datastore` name | ❌ | `vaai_copy.go:404` |
| SRC-137 | **XCOPY silently powers off a running production VM** rather than refusing | `runtime.powerState` | ⚠ | `vaai_copy.go:98-106` |

## Basic level — network

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-140 | **Orphaned portgroup** — the NIC's backing no longer exists, so it cannot be mapped | `…(VirtualEthernetCard).backing == nil` | ❌ | `credutils.go:823`, `:1947` |
| SRC-141 | **SR-IOV and vRDMA NICs are silently dropped** — the type switch covers only E1000/E1000e/VMXNET/PCNet32, and the count drift later aborts the migration | NIC concrete type | ❌ *silent* | `credutils.go:774-784`; trips `migrate/migrate.go:241` |
| SRC-142 | **NSX-T opaque networks** are unsupported | `…backing.(OpaqueNetworkBackingInfo)` | ❌ | `credutils.go:850-868`; docs `network-storage-mapping.md:18` |
| SRC-143 | Cross-vCenter DVS — the backing carries a portgroup *key*, not a live reference | `…backing.port.portgroupKey` | ⚠ | `credutils.go:840-848` |
| SRC-144 | Network persistence is unsupported on Windows Server 2008 / 2012 | `config.guestId` | ❌ | docs `network-persistence.md:65` |

## Basic level — guest OS, firmware, security

Everything here reads `config.guestId`, which is available **even when the VM is powered off** — unlike the Tools-reported OS the pipeline uses today.

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-150 | **`config.guestId` is never used** — OS detection relies on Tools-only `guest.guestFamily`, so a powered-off VM hard-aborts inside the migration pod | `config.guestId` | ❌ | `credutils.go:1974`, `2011-2014`; `vm/vmops.go:294` |
| SRC-151 | **The supported-distro allowlist runs after the full disk copy** — 17 literals; Gentoo, Arch, Alpine, Amazon Linux and Photon are absent | `config.guestId` | ❌ | `migrate/conversion.go:174-188` |
| SRC-152 | Solaris, FreeBSD, macOS and `otherGuest` are handled nowhere | `config.guestId` | ❌ | no handler; contradicts `compatibility.md:37` |
| SRC-153 | **Windows Server 2012 needs a pinned virtio ISO**; a failed version probe silently selects the wrong one, producing an unbootable guest | `config.guestId` | ❌ | `virtv2vops.go:419-429`, `:1213-1216` |
| SRC-154 | SUSE with legacy GRUB 0.97 + BIOS + multiple disks → GRUB Error 21 | `config.guestId`, `config.firmware`, disk count | ❌ | `virtv2vops.go:1255-1263`; docs `known-limitations.md:123` |
| SRC-155 | SLES 11 `mkinitrd` cannot resolve LVM paths, and the fix is non-fatal on failure | `config.guestId` | ⚠ | `virtv2vops.go:1332`; `conversion.go:395` |
| SRC-156 | Multi-disk Windows — non-`C:` disks come up **Offline** due to SAN policy | `guestFamily`, disk count | ⚠ | docs `windows-offline-disks.md:24` |
| SRC-157 | **Secure Boot is read nowhere** — unsigned virtio drivers will not load on a Secure Boot guest | `config.bootOptions.efiSecureBootEnabled` | ⚠ | zero hits repo-wide; only `firmware == "efi"` is read |
| SRC-158 | **VBS enabled** must be disabled before migration — distinct from the vTPM device check | `config.flags.vbsEnabled` | ⚠ | docs `vtpm_migration.md:37-42`; zero hits |
| SRC-159 | An encrypted VM additionally requires `Cryptographer.Access` / `.Decrypt` on the role — today this surfaces as a confusing transport error | `config.keyId` + `AuthorizationManager` | ❌ | docs `prerequisites.md:64-81` |
| SRC-160 | The legacy-chipset guest list currently in the catalog is an explicit placeholder | `config.guestId` | ⚠ | `vassessment/catalog/rules_src.go:17-26` |

## Basic level — VM state, host and platform

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-170 | **A powered-off VM has a nil `config` and is skipped from discovery entirely** | `runtime.powerState`, `config` | ❌ | `credutils.go:1818-1821` |
| SRC-171 | **Duplicate VM names across datacenters — the first match silently wins**, so the wrong VM can be migrated | `config.name` across all datacenters | ❌ | `v2v-helper/vcenter/vcenterops.go:232-245` |
| SRC-172 | A VM already migrated and powered back on produces no new Migration object | name suffix / existing target | ⚠ | docs `archives/release_notes.md:36` |
| SRC-173 | **ESXi reachability for every host in the cluster** (DNS + ports 902/443) — the VM may vMotion mid-copy | `runtime.host`, cluster `host[]`, `sslThumbprint` | ❌ | `nbd_errors.go:39-51` |
| SRC-174 | **ESXi host fan-out** — concurrent copies against one host exhaust hostd NFC memory | `runtime.host` grouped across the plan | ⚠ | `nbd/nbdops.go:69-76` |
| SRC-175 | Fault Tolerance enabled, suspended VMs, or a mixed-CPU cluster without EVC | `runtime.faultToleranceState`, `powerState`, `summary.currentEVCModeKey` | ❌ | docs `cluster-conversion.md:36-45` |
| SRC-176 | **Exotic devices are read nowhere** — connected CD-ROM/ISO, floppy, serial, parallel, USB, NVDIMM, and NVMe/SATA/IDE controllers | `config.hardware.device[]` | ⚠ | only `maintenance_mode.go:380-400`, which is Cluster Conversion, not VM migration |
| SRC-177 | **Template VMs are flagged in the catalog but not filtered in the live pipeline** | `config.template` | ❌ | `rules_src.go:47`; no read in either runtime module |
| SRC-178 | vCenter version outside 6.7 / 7.0 / 8.0 | `about.version`, `apiVersion` | ℹ | docs `compatibility.md:10-12` |
| SRC-179 | PCI slot exhaustion — roughly 26 volumes with virtio-blk on the agent | disk count × concurrency | ⚠ | docs `known-limitations.md:186` |

## Intermediate level — guest properties (VMware Tools)

| ID | Check | Reads | Sev | Evidence |
|---|---|---|---|---|
| SRC-190 | **`guest.disk[].freeSpace` is never collected**, yet low free space is the single most common conversion abort (`/` 100 MB, `/boot` 50 MB, `C:` 100 MB) | `guest.disk[]` | ❌ | `virtv2v_errors.go:39,41,85`; docs `known-limitations.md:209-222` |
| SRC-191 | **`guest.toolsRunningStatus` is never read** — it predicts both the 5-minute shutdown timeout and that quiesced snapshots will silently be unquiesced | `guest.toolsRunningStatus` | ⚠ | `vmops.go:883`; `:626` |
| SRC-192 | **IPv6 addresses are silently discarded throughout**, so an IPv6-only VM has no address preserved | `guest.net[].ipConfig.ipAddress[]` | ❌ *silent* | `v2v-helper/vm/netinfo.go:33,50`; `virtv2vops.go:674` |
| SRC-193 | When `guest.net` is non-empty the hardware-NIC fallback is never consulted, so a MAC absent from Tools gets no IPs | `guest.net[].macAddress` vs device MACs | ❌ *silent* | `netinfo.go:25-41` |
| SRC-194 | More than one IP per NIC is unsupported | `guest.net[].ipConfig.ipAddress[]` | ❌ | `migrate/network.go:189`; docs `known-limitations.md:103` |
| SRC-195 | Two NICs on the same subnet cause asymmetric routing that port security drops | per-NIC IP + prefix | ⚠ | docs `network-persistence.md:81-84` |

## Advanced level — in-guest script (Phase 2)

Not detectable from vCenter. Listed to justify the optional script and to scope Phase 2.

| ID | Check | Sev | Evidence |
|---|---|---|---|
| GST-001…009 | Baseline: bootloader, multi-boot, LDM, BtrFS, LUKS/BitLocker, GPO, AD DC role, databases, containers | mixed | product doc |
| GST-100 | **Immutable `/etc/resolv.conf`** (`lsattr +i`) — and note it is not among the 20 recognised patterns, so it falls to a generic catch-all error | ❌ | docs `troubleshooting.md:96`; `virtv2v_errors.go` |
| GST-101 | LDM **system** volume — virt-v2v has no guard, so conversion produces an unbootable disk or hangs | ❌ | `virtv2vops.go:878-881` |
| GST-102 | BtrFS with Snapper snapshots is misdetected as multi-boot | ❌ | `guestfish.go:7-13` |
| GST-103 | Windows hibernation / Fast Startup leaves a dirty NTFS volume | ❌ | `virtv2v_errors.go:69` |
| GST-104 | Xen-only kernels | ❌ | `virtv2v_errors.go:60` |
| GST-105 | Missing `dracut` / `mkinitrd` | ❌ | `virtv2v_errors.go:64` |
| GST-106 | Inode exhaustion (`df -i`) — free bytes can look fine while inodes are exhausted | ❌ | `virtv2v_errors.go:41` |
| GST-107 | XFS/ext filesystem corruption | ❌ | `virtv2v_errors.go:56`, `:95` |
| GST-108 | RHEL 7 missing `/boot/grub/grub.cfg` symlink | ⚠ | virt-v2v behaviour |
| GST-109 | **AD Domain Controller** — VM-GenerationID is not carried on disk; 2008 R2 risks USN rollback | ❌ | docs `known-limitations.md:35-42` |

## Destination level — PCD (Phase 3)

| ID | Check | Sev | Evidence |
|---|---|---|---|
| DST-001…006 | Baseline: VLAN, subnet, cluster, volume type, tenant, capacity vs quota | ❌/⚠ | product doc |
| DST-100 | **There is no OpenStack quota pre-check anywhere** — instances, vCPU, RAM, volumes, GB and ports all fail at creation time instead | ❌ | zero `quota` hits across all Go modules; fails at `migrate.go:355`, `:419` |
| DST-101 | A preserved IP that falls outside every subnet CIDR on the mapped network | ❌ | `openstack/clients.go:633` |

---

## Shift-left: checks enforced too late today

These already exist as enforcement, but only *after* a migration has started — several after the entire disk copy has completed. Every one is detectable at assessment time, and moving it earlier is the clearest measurable win this product offers.

Ranked by wasted work:

1. **SRC-151 / SRC-150** — the distro allowlist and OS-family detection fire *after the full disk copy*, when `config.guestId` answers both before the first byte moves
2. **SRC-100 / SRC-101** — independent disk mode, discovered mid-copy by VDDK
3. **CBT and hardware version < 7** — checked at `nbd_copy.go:54`, and `config.version` is never even persisted to `VMInfo`
4. **Disk-count ↔ volume-type and NIC-count ↔ network-count mismatches** (`migrate.go:238`, `:241`) — both operands are known at plan time
5. **Mapping completeness** (`migrationplan_controller.go:2196-2265`) — and because `TriggerMigration` returns on the first error (`:2365`), one bad VM starves an entire batch
6. **VDDK presence**, checked per VM after the Migration CR already exists (`:2374-2379`)
7. **SRC-137** — XCOPY silently powering off a running production VM

## Contradictions to resolve

Two conflicts between the docs and the code. Both need an owner's ruling before the affected checks can be written, because the check's correct behaviour depends on which side is right:

1. **virtio-win version for Windows Server 2012** — `windows-ldm-migration.md:54` says 0.1.185, `faq.md:37` says 0.1.189, both keyed on the same `guestId` (affects SRC-153)
2. **FreeBSD 14** is listed as "Verified" in `compatibility.md:37`, but would be rejected by the distro allowlist at `conversion.go:174-188` (affects SRC-152)

## Coverage summary

| Level | Baseline | New | Total | Phase |
|---|---|---|---|---|
| Basic | 17 | ~45 | ~62 | 1 |
| Intermediate | 6 | 6 | ~12 | 1 |
| Advanced | 9 | 10 | ~19 | 2 |
| Manual | 3 | — | 3 | — |
| Destination | 6 | 2 | 8 | 3 |

**Phase 1 target is roughly 74 checks, against a baseline of 23.** The increase is almost entirely the long tail of real failure modes already encoded in vJailbreak's own error handling — which is exactly the safety margin this product is meant to provide.

Each check needs, before implementation: a stable ID, an owner-approved severity, remediation text, and a table-driven unit test covering pass, fire, and not-evaluated (`tasks.md` T024).
