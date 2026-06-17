# ADR-0003: Ship real JSON Schemas with an embedded validator (no new dependency)

Status: accepted (M1)

## Context

M1 requires "Go types + JSON Schema + validators" for every short-form artifact, with
positive and negative ("accept/reject") schema tests. It also requires (§11) that a
clean checkout + `make verify` reproduces all green gates **offline** — "no network
egress required to pass tests" (§12).

The repo currently vendors no JSON-Schema library. Adding one (e.g.
`santhosh-tekuri/jsonschema`) would require a module download on a clean offline
checkout, jeopardizing the takeover test, and would add a dependency surface for a
milestone whose ethos is "fewer, correct, tested components."

## Decision

- Author real **JSON Schema (Draft 2020-12 subset)** documents, committed under
  `docs/schemas/*.schema.json` and embedded into the binary via `go:embed`.
- Implement a small, dependency-free schema interpreter in `internal/shortform/schema`
  that validates a decoded instance against a parsed schema. Supported keywords (the
  exact subset the artifact schemas use): `type` (object/array/string/number/integer/
  boolean/null), `required`, `properties`, `additionalProperties` (bool), `enum`,
  `const`, `items` (single subschema), `minItems`, `maxItems`, `minimum`, `maximum`,
  `minLength`, `pattern`, plus ignored annotations (`$id`, `$schema`, `title`,
  `description`). Unsupported keywords are rejected at schema-load time so a schema can
  never silently pass instances it does not actually constrain.
- The schema is therefore **load-bearing**: the accept/reject tests validate instances
  through the interpreter against the committed schema, not against a hand-rolled Go
  check that happens to agree with it.

## Consequences

- Offline, reproducible, zero new go.mod dependencies.
- The interpreter is intentionally a subset; it is unit-tested directly (keyword by
  keyword) and indirectly (per-artifact accept/reject). If a future schema needs a
  keyword outside the subset, it must be added to the interpreter (fail-closed at load).
- Typed Go structs remain the authoring surface; schemas validate the serialized form;
  additional semantic checks (cross-field invariants) live in Go validators on top.
