---
name: adr
description: >-
  Author and maintain Architecture Decision Records (ADRs) in docs/adr/ using a
  short context-decision-why format. Use when the user asks for an ADR, records
  an architecture decision, mentions docs/adr, or locks a hard-to-reverse
  trade-off (display backend, nav engine, protocol shape, licensing).
---

# Architecture Decision Records

## When to write an ADR

Create or update an ADR only when **all three** are true:

1. **Hard to reverse** — changing later has real cost
2. **Surprising without context** — a future reader will wonder why
3. **Real trade-off** — alternatives existed and one was chosen for reasons

If any is missing, skip the ADR (a README note or code comment is enough).

## Location and numbering

- Path: `docs/adr/` at the repo root (create lazily)
- Files: `NNNN-kebab-slug.md` (zero-padded, start at `0001`)
- Scan existing files; next number = highest + 1
- Optional index: `docs/adr/README.md` listing title + one-line status

If `CONTEXT-MAP.md` exists, system-wide ADRs stay in root `docs/adr/`; context-specific ADRs may live under that context’s `docs/adr/`.

## Template (keep short)

```md
---
status: accepted
date: YYYY-MM-DD
---

# {Short title}

{1–3 sentences: context, decision, why.}

## Considered options

- **Chosen** — …
- **Rejected: X** — …

## Consequences

- …
```

Omit **Considered options** / **Consequences** when they add nothing. A one-paragraph ADR is valid.

`status`: `proposed` | `accepted` | `deprecated` | `superseded by ADR-NNNN`

## Workflow

1. Confirm the three criteria (or that the user explicitly wants an ADR anyway).
2. Search `docs/adr/` and related docs (`CONTEXT.md`, `protocol/`, README) so you don’t duplicate or contradict.
3. Write the next numbered file; update `docs/adr/README.md` if present.
4. If the decision introduces domain terms, update `CONTEXT.md` (glossary only — not implementation detail).
5. Link the ADR from the relevant README/protocol section only when readers would otherwise miss it.

## Retroactive ADRs

When catching up on built-but-undocumented decisions: one ADR per distinct trade-off, past-tense “we decided”, `status: accepted`, date ≈ when it landed if known. Do not invent debate that did not happen; keep options honest.

## Anti-patterns

- Filling ceremony sections (Context/Decision/Status tables) with no content
- ADRs for every library bump or obvious default
- Encoding OsmAnd-/MapKit-specific types into cross-platform protocol ADRs — keep wire format engine-agnostic
- Superseding by editing history — add a new ADR and mark the old one superseded

## Moto HUD

This repo’s ADRs live in [`docs/adr/`](../../../docs/adr/). Domain language: [`CONTEXT.md`](../../../CONTEXT.md).
