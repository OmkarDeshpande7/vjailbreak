# Research: vAssessment Phase 1

**Date**: 2026-09-15
**Feeds**: `spec.md`, `plan.md`

This document records what was found in the existing vJailbreak codebase and infrastructure, and the decisions taken as a result. Every claim below is anchored to a `file:line` so it can be re-verified when the code moves.

---

## R1 — Code reuse: what vAssessment can take from vJailbreak today

The explicit reason vAssessment lives inside the vjailbreak repo rather than its own is **code reuse** — above all for vCenter and PCD connectivity. The controlling constraint is that vAssessment is a **standalone binary with no Kubernetes cluster**, so anything that requires a `client.Client`, reads a CR, or reads a Secret/ConfigMap is not reusable as-is regardless of how well it is written.

### R1.1 Directly reusable today (no Kubernetes dependency)

| Component | Location | Why it is reusable |
|---|---|---|
| vCenter session cache | `pkg/common/vmware/sessioncache.go:31,68,73,89,102,111` — `CredentialFingerprint`, `ClientCache`, `NewClientCache`, `Get`, `Store`, `Delete` | Pure govmomi. Caches an authenticated `*vim25.Client` keyed by credential fingerprint. Exactly what a rescan-heavy tool needs. |
| Proxy-aware HTTP client + vCenter URL normalization | `pkg/common/utils/net.go:19,251` — `VjbNet`, `NormalizeVCenterURL` | Pure stdlib/net. Already used by both existing validators. Customers frequently sit behind a proxy, so this is not optional. |
| OpenStack flavor/AZ matching | `pkg/common/openstack/flavor.go:23,29,35` | Pure gophercloud, operates on `[]flavors.Flavor`. |
| OpenStack network + security group helpers | `pkg/common/openstack/network.go:13`, `securitygroups.go:18` | Take a `*gophercloud.ServiceClient`; no k8s. |

**Decision**: import these from `pkg/common` unchanged. This is the baseline reuse and requires no refactor.

### R1.2 Reusable core trapped inside a Kubernetes wrapper

These contain exactly the logic vAssessment needs, wrapped in credential retrieval it cannot use:

| Component | Location | The problem |
|---|---|---|
| vCenter login + retry | `pkg/common/validation/vmware/validate.go:123-195` | The govmomi login and retry loop is k8s-free, but `Validate()` reaches a Secret for credentials and a ConfigMap for the retry limit. `FetchResourcesPostValidation` (`:207`) writes CRs and is not reusable at all. |
| OpenStack authentication | `pkg/common/validation/openstack/validate.go:77-150` | `gophercloud.AuthOptions` + `openstack.Authenticate` against a plain `OpenStackCredsInfo` struct — fully reusable once credentials arrive from somewhere other than a Secret. `verifyCredentialsMatchCurrentEnvironment` (`:238`) reads host DMI/metadata to assert "these creds belong to the cluster I am running in", which is meaningless for a CLI pointed at an arbitrary cloud. |
| Credential structs | `pkg/common/vmware/credentials.go:19,29`, `pkg/common/openstack/credentials.go:20,29` | Every exported function takes `client.Client` and reads a `corev1.Secret`. The *shape* (`VMwareCredsInfo{Host, Username, Password, Datacenter, Insecure}`) is plain data and is fine; the retrieval is not. |

**Decision**: split **credential acquisition** from **credential use**. The transport/auth core moves behind an interface that takes a plain credentials struct; the k8s Secret reader and a new CLI/env reader both become implementations that produce that struct. See R1.5.

### R1.3 Not reusable

`pkg/common/k8s/client.go:16` (`GetInclusterClient`) and `pkg/common/config/settings.go:52` (`GetVjailbreakSettings`, reads a ConfigMap) are inherently cluster-bound. vAssessment needs its own configuration mechanism — flags, env, and a config file.

### R1.4 Duplication already present in the repo

This matters because vAssessment would otherwise become the *fourth* copy of the same code:

| Capability | Implementations |
|---|---|
| govmomi client creation + login/retry | `pkg/common/validation/vmware/validate.go:123-195` **and** `v2v-helper/vcenter/vcenterops.go:48,104` |
| Session-alive check | `v2v-helper/vcenter/vcenterops.go:114` (`EnsureSessionActive`) **and** `pkg/common/vmware/sessioncache.go:126` (`defaultCheckSession`) |
| Datacenter / cluster / host / datastore enumeration | `k8s/migration/pkg/utils/vmwareutils.go:39,401` **and** `v2v-helper/vcenter/vcenterops.go:180,349` **and** `pkg/vpwned/sdk/targets/vcenter/vcenter.go:286` — **three** independent implementations |
| OpenStack client creation | `pkg/common/validation/openstack/validate.go:77-150` **and** `k8s/migration/pkg/utils/credutils.go:364,400` |

**Decision**: vAssessment must not add a fourth inventory-enumeration implementation. Extracting the shared core into `pkg/common` is therefore not gold-plating — it is the only way to satisfy the reuse mandate without making the duplication worse.

### R1.5 Proposed `pkg/common` refactor

Each item is independently shippable and behaviour-preserving. **None of this is in scope for the current PR** — it is the specification for the work that follows.

| # | Change | Rationale | Risk |
|---|---|---|---|
| **C1** | Extract the govmomi connect/login/retry core from `validation/vmware/validate.go:123-195` into `pkg/common/vmware/connect.go`, taking a plain credentials struct and returning an authenticated client. Existing `Validate()` calls it after fetching the Secret. | Removes the k8s dependency from the part vAssessment needs; single implementation for all modules | Low — pure extraction, callers unchanged |
| **C2** | Define a `CredentialProvider` interface in `pkg/common/vmware` with two implementations: the existing Secret-backed reader, and a new static/env/flag-backed one. | This is the actual blocker. Credential *acquisition* is the only genuinely k8s-bound part | Low |
| **C3** | Same extraction for OpenStack: auth core out of `validation/openstack/validate.go:77-150` into `pkg/common/openstack/connect.go`; keep `verifyCredentialsMatchCurrentEnvironment` out of it, as it is appliance-specific | Symmetry with C1; needed for Phase 3 destination checks | Low |
| **C4** | Add `pkg/common/vmware/inventory.go`: a single `ContainerView` + `property.Collector` bulk read of `[]mo.VirtualMachine` over a caller-supplied property list | The performance requirement (NFR-001) depends on this, and it consolidates the three duplicate enumerations. Only one real batch-retrieve exists today (`k8s/migration/pkg/utils/maintenance_mode.go:157`); the production inventory path (`credutils.go:600` `GetAndCreateAllVMs`) is a `finder.VirtualMachineList` plus per-VM goroutines, which will not meet the target | Medium — the win is real but migrating existing callers should be a separate, later change |
| **C5** | Break the `pkg/common/constants` → `vjailbreakv1alpha1` import. The constants themselves are plain values, but the package imports the CRD types, so importing one constant transitively pulls in the entire CRD module | A standalone CLI should not depend on the CRD module to read a string constant | Low, but touches many files |

**Sequencing decision**: Phase 1 depends on **C1 and C2 only**. C4 is a performance optimisation that vAssessment can implement inside its own module first and promote to `pkg/common` once proven. C3 is not needed until destination checks. C5 is hygiene and can be done at any time.

**Anti-decision**: do *not* attempt to move `v2v-helper`'s or `pkg/vpwned`'s vCenter code into `pkg/common` as part of this work. Those modules are on the migration hot path; consolidating them is a separate change with its own risk profile, and coupling it to vAssessment's delivery would block both.

### R1.6 Capabilities that do not exist anywhere and must be built

Searched across all non-vendor code:

| Needed by | Status |
|---|---|
| vTPM detection (`VirtualTPM` device), USB devices (`VirtualUSB`), VM encryption (`config.keyId`) | **No implementation anywhere.** Must be built. |
| Guest properties — `guest.guestFullName`, `guest.disk` (free space), `guest.net` (IPs) | **No implementation anywhere.** Every Intermediate-level check depends on these. |
| `guestinfo.*` read/write via govmomi | **Does not exist.** The only `ExtraConfig` usage is `disk.enableUUID` toggling (`k8s/migration/internal/controller/proxyvm_controller.go:226,371,389`) and OVF properties (`pkg/vpwned/server/ova_deploy.go:255`). The Phase 2 guest-script delivery mechanism has no precedent in this repo and carries the risk noted in `plan.md`. |

This is a significant finding: a large share of Phase 1's Intermediate level is genuinely new code, not reuse.

---

## R2 — How the appliance is built and published today

Read from `.github/workflows/packer.yml`, `image_builder/`, and the root `Makefile`.

### R2.1 What exists

- **Five container images** — `vjailbreak-ui`, `vjailbreak-v2v-helper`, `vjailbreak-controller`, `vjailbreak-vpwned`, `vjailbreak-ai` — pushed to `quay.io/platform9/*` (`packer.yml:43-48`, login at `:517-521`).
- **One QCOW2 appliance image**, built by Packer/QEMU (`image_builder/vjailbreak-image.pkr.hcl:40-49`), published twice: as an OCI artifact via ORAS to `quay.io/platform9/vjailbreak:<TAG>` (`packer.yml:741-746`), and to S3 (`packer.yml:776-813`). Bucket names are GitHub repo variables (`S3_BUCKET_PROD`, `S3_BUCKET_DEV`, `S3_REGION` — `packer.yml:52-54`), not literals.
- **Tag scheme** (`packer.yml:174-186`): release event → the release tag; nightly → `<last git tag>-<short sha>`; otherwise → `<run_number>-<short sha>`.
- **Gating**: container images build on ordinary pushes and PRs, but the entire QCOW2 half is gated on `release_found == 'true' || is_nightly == 'true'` (`packer.yml:682` onward).

### R2.2 What `:latest` actually means here

**There is no `:latest` container tag in this repository.** Nothing publishes one.

`latest` exists only as an **S3 path prefix** (`packer.yml:783-802`):

- release event → `releases/<tag>` **and** `releases/latest`
- nightly → `nightly/<tag>` **and** `nightly/latest`
- anything else (including ordinary merges to `main`) → `dev/<tag>`, **and no `latest` is written**

So the assumption that "whatever was last built in the repo is what a user gets from `:latest`" does **not** hold today. A merge to `main` publishes container images but moves no `latest` pointer and produces no QCOW2 at all. Only cutting a GitHub release moves `releases/latest`; only the nightly cron (19:30 UTC, and only if `main` received a commit in the preceding 24h — `nightly-build-trigger.yml:5-6,37-41`) moves `nightly/latest`.

**Decision**: if vAssessment wants `:latest` to mean "the newest released build", that must be implemented deliberately — it is not inherited. The recommendation is to mirror the existing semantics exactly (`latest` = newest *release*, plus a separate `nightly/latest`) rather than invent a third meaning, because support engineers already reason about vJailbreak artifacts this way.

### R2.3 Can vAssessment piggyback on `packer.yml`?

No. `packer.yml` is a monolith: every image is a hardcoded job with its own env var, its own artifact download and its own push line. There is no matrix and no parameterisation; adding a sixth image means editing roughly six places, and it would couple vAssessment's release to vJailbreak's — the precise outcome the product owner ruled out.

**Decision — independent delivery**:

1. **New workflow** `.github/workflows/vassessment-release.yml`, entirely separate from `packer.yml`. The existing `vassessment.yml` (build + test on every PR) stays as the CI gate and remains publish-free.
2. **New quay repository** — `quay.io/platform9/vassessment` — used both for the container image and, via ORAS, for the QCOW2 artifact, mirroring how `quay.io/platform9/vjailbreak` carries both today. **This must be created in Quay, with the existing `QUAY_ROBOT_*` credentials granted push access, before the first release.**
3. **No new S3 bucket.** Reuse `S3_BUCKET_PROD` / `S3_BUCKET_DEV` under a `vassessment/` prefix (`vassessment/releases/<tag>`, `vassessment/releases/latest`, `vassessment/nightly/<tag>`, `vassessment/dev/<tag>`). Provisioning a separate bucket buys nothing and doubles the credential surface.
4. **`cleanup-old-builds.yaml` prunes by prefix** and will need a matching entry, or vAssessment's dev builds will accumulate indefinitely.
5. **Own version file** — vAssessment's equivalent of `image_builder/configs/version-config.yaml`, so its reported version is independent of the appliance's.
6. **Copy, do not import**, the tag-determination and packer/ORAS/S3 step sequence. Sharing them via a composite action is tempting but would re-couple the two release trains.

### R2.4 Independent release, restated

A vAssessment release must be cuttable without building or publishing any vJailbreak appliance artifact, and a vJailbreak release must not force a vAssessment rebuild. The two share a git repository and `pkg/common`; they share nothing else. Note the consequence: a breaking change to `pkg/common` affects both, so `pkg/common` changes must stay backward-compatible or be coordinated — this is the standing cost of the monorepo arrangement, and it is accepted deliberately in exchange for the reuse.

---

## R3 — How to expose the vAssessment UI ("the Grafana pattern")

The brief was to copy how Grafana is exposed. What the code actually does differs from the common description of it, so it is worth stating precisely.

### R3.1 What Grafana actually does

It is **not** a NodePort, and **not** a port assigned per click. It is a **path-prefix Ingress on the appliance's single shared nginx**, and the user never leaves port 443.

- `k8s/kube-prometheus/manifests/grafana-service.yaml:9-20` — Service `grafana` in namespace `monitoring`, port `3000`, type `LoadBalancer` (k3s klipper-lb). This is *not* the path the UI uses.
- `ui/deploy/ui.yaml:110-128` — the real exposure: an Ingress named `grafana-api-ingress`, `ingressClassName: nginx`, `path: /grafana`, `pathType: Prefix`, backend `grafana:3000`, with `nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"`. Traefik is disabled and ingress-nginx is installed by `image_builder/scripts/install.sh:419,438`.
- `k8s/kube-prometheus/manifests/grafana-config.yaml:12-18` — the part that makes a sub-path work at all: `root_url = %(protocol)s://%(domain)s:%(http_port)s/grafana/` and `serve_from_sub_path = true`.
- `ui/src/config/navigation.tsx:92-98` — nav entry `{ id: 'monitoring', label: 'Monitoring', path: '/grafana', external: true }`.
- `ui/src/components/layout/Sidenav/Sidenav.tsx:171-176` — the link is computed as `https://${window.location.host}${path}` and opened in a new tab. Nothing about the host or port is hardcoded.

### R3.2 Two flaws in the pattern not to inherit

1. **The Grafana Ingress is defined in the UI's own manifest** (`ui/deploy/ui.yaml`), not in Grafana's. vAssessment must ship its own Ingress in its own manifest, or it recreates this coupling.
2. **Authentication does not carry over.** The UI container's nginx does basic auth against a hostPath `/etc/htpasswd` (`ui.yaml:34-41`, seeded at `install.sh:347-350`), but the `/grafana` Ingress does not traverse the UI's nginx and carries no auth annotation — so Grafana is protected only by its own login, and no admin credential is configured anywhere in `k8s/kube-prometheus/` (`env: []`), leaving stock defaults. **vAssessment must decide its own authentication explicitly and must not assume the appliance's basic auth covers it.**

### R3.3 Decision

For the embedded/appliance deployment, copy the mechanism and not the mistakes: a Deployment, a ClusterIP Service, and one `pathType: Prefix` Ingress at `/vassessment` on the shared nginx class, shipped in vAssessment's own manifest, with the binary configured to serve from that sub-path, and a nav entry with `external: true`.

**However** — Phase 1 does not deploy into the appliance at all. Phase 1 is a standalone binary a user runs on a laptop or jump host, which serves its UI on a local port directly. The Ingress pattern above is the Phase 3 embedded story. Recording it now because it constrains a Phase 1 decision: **the binary must be able to serve its UI under a configurable base path**, since retrofitting sub-path support into a frontend after it is written is considerably more painful than building it in.

---

## R4 — Open risks carried into the plan

| Risk | Impact | Mitigation |
|---|---|---|
| Phase 2's `guestinfo` delivery has no precedent in the repo, and the read path must work under the minimal read-only role | Advanced level has no delivery mechanism at all if it fails; the redesign removed the file-upload fallback | Prototype the read path against a real vCenter before committing Phase 2 scope. Out of Phase 1 scope but must not be discovered late. |
| BIOS UUID is the design's stated merge key but is not guaranteed unique — it is settable and survives cloning | Silent VM merging or double-counting across sources | Detect collisions and surface them; never merge silently (FR-004) |
| A read-only role missing a privilege makes checks silently unevaluable | Assessment looks complete but is not | Report `not_evaluated` with the missing privilege named (FR-016); validate privileges at connect time (FR-001) |
| The estate figures in the design do not reconcile — its report header and inventory header show different totals and baselines for the same scan | Loss of confidence in the headline number | One canonical estimator owns every figure; reconciliation asserted by test (SC-005, FR-030) |
| `pkg/common` is shared by two independently-released products | A breaking change to it breaks the other product's build | Additive changes only; both modules' CI runs on any `pkg/common` change |
