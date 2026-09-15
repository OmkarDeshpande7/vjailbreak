# vAssessment Specifications

Specs for the `vassessment` module, following the same Spec-Driven Development layout as the repository-root `specs/` directory and the templates in `.specify/templates/`.

They live here rather than at the repository root because `vassessment` is an independently released product that happens to share this repository. Its specs version alongside its code.

## Layout

```text
vassessment/specs/
└── 001-vassessment-phase1/
    ├── spec.md                      # WHAT and WHY — user stories, requirements, success criteria
    ├── plan.md                      # HOW — architecture, decisions, constitution check, phasing
    ├── research.md                  # Findings behind the decisions, with file:line evidence
    ├── data-model.md                # Entities, invariants, storage sketch
    ├── quickstart.md                # Build and run
    ├── contracts/
    │   ├── rest-api.md              # HTTP contract, per UI page
    │   └── pre-check-catalog.md     # The check catalog — rule IDs are a public contract
    ├── checklists/
    │   └── requirements.md          # Pre-implementation quality gate
    └── tasks.md                     # Ordered, dependency-aware work breakdown
```

## Reading order

1. **`spec.md`** — what is being built and why
2. **`plan.md`** — how, plus **two unresolved constitution questions that block implementation**
3. **`research.md`** — the evidence: what can be reused, how the appliance is built and published, how Grafana is actually exposed
4. **`data-model.md`** / **`contracts/`** — the detail an implementer needs
5. **`checklists/requirements.md`** — what must be answered first
6. **`tasks.md`** — the work breakdown

## External sources of truth

| Artifact | Role |
|---|---|
| [Product doc (Confluence)](https://platform9.atlassian.net/wiki/spaces/vJailbreak/pages/6262587411) | Product intent, phasing, original check catalog |
| [API doc (Google Docs)](https://docs.google.com/document/d/1aF91tcAo3Qe6zQ3nb5ttIIckNGc9MoLQdfW9cgy94ZE/edit) | Shared with the UI team; `contracts/rest-api.md` is the in-repo copy of record and wins on conflict |
| vAssessment redesign canvas | The 4-page UI and the generated scope report |
| [vJailbreak docs](https://platform9.github.io/vjailbreak/introduction/getting_started/) | Prerequisites and known limitations feeding the check catalog |

## Status

| Spec | Status | Tracking |
|---|---|---|
| 001 — Phase 1 | Draft; blocked on two constitution rulings (`plan.md`) | [#2385](https://github.com/platform9/vjailbreak/issues/2385) epic, [#2386](https://github.com/platform9/vjailbreak/issues/2386), [#2387](https://github.com/platform9/vjailbreak/issues/2387), [#2388](https://github.com/platform9/vjailbreak/issues/2388) |

Phases 2 (guest script / Advanced level), 3 (destination checks and embedded mode), and 4 (bucket seeding) get their own spec directories when they are scheduled.
