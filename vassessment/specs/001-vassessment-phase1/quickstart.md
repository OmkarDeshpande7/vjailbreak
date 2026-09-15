# Quickstart: vAssessment

**Date**: 2026-09-15
**Status**: Phase 1 target state. Commands marked **(planned)** do not exist yet — this file documents the intended experience so the CLI surface is designed rather than accreted.

## Build

```bash
cd vassessment
make build          # -> bin/vassessment
make test
make run            # go run ./cmd/vassessment
```

From the repo root:

```bash
make vassessment        # build
make test-vassessment   # test
```

## Available today (#2386)

```bash
vassessment version

# Offline validation — no vCenter contact
vassessment validate connection --host=vcenter.example.com \
                                --username=assessment@vsphere.local \
                                --password=***
vassessment validate role --granted=granted-role.json --required=required-role.json
```

`validate connection` checks input shape only. `validate role` diffs an exported vCenter role against a required-privileges file. Neither opens a network connection — they are the pre-flight that runs before a credential is ever used.

## Phase 1 target experience

### Serve the UI **(planned)**

```bash
vassessment serve --port 8080
```

Opens the assessment UI on `http://localhost:8080`. Everything below is doable from that UI; the CLI equivalents exist for scripted and CI use.

### Headless scan **(planned)**

```bash
vassessment scan --host vcenter.example.com \
                 --username assessment@vsphere.local \
                 --password-stdin \
                 --out ./estate.assessment

vassessment scan --rvtools ./rvtools-export.xlsx --out ./estate.assessment
```

Credentials are accepted via `--password-stdin` or environment variable, never as a plain argument — process arguments are visible to any user on the host, and these are customer production credentials.

### Report **(planned)**

```bash
vassessment report --in ./estate.assessment --format pdf --out ./migration-scope.pdf
vassessment report --in ./estate.assessment --format csv --out ./migration-scope.csv
```

## Read-only guarantee

vAssessment issues no mutating vSphere call under any code path. This is a product property, asserted by test against a simulated vCenter, not a convention (FR-008, SC-002). The shipped role definition contains the minimum privileges needed and nothing more.

## What you need from the customer

1. A vCenter hostname reachable from where the binary runs
2. A read-only account — apply the shipped role definition:
   ```bash
   vassessment validate role --granted=<exported-role>.json --required=<shipped>.json
   ```
3. Nothing else. No appliance, no Kubernetes, no agent in any guest, no change to any VM.

If credentials are not yet issued, start from an RVTools export instead and add the live connection later — both sources for the same vCenter merge (FR-004).

## Development

```bash
cd vassessment
go build ./... && go vet ./... && go test ./...
gofmt -l .        # must print nothing
```

Tests must not contact a real vCenter. Use `govmomi/simulator` (`vcsim`) — constitution principle IV.

Working on shared vCenter or OpenStack connectivity in `pkg/common`? Run **both** module test suites; `pkg/common` is shared with the appliance and must stay backward-compatible (`research.md` R2.4):

```bash
cd pkg/common     && go test ./...
cd ../vassessment && go test ./...
```
