---
name: go-mutation-testing-tool
status: accepted
date: 2026-09-08
extends: ".claude/skills/solution-conformance-testing/adr/mutation-tool-per-stack.md"
---

# ADR 0001 — Mutation-testing tool for the Go stack

## Context

[solution-conformance-testing](../../.claude/skills/solution-conformance-testing/SKILL.md)
mandates mutation testing on top of coverage and picks the tool **per stack**, but
its ADR only names tools for C#/.NET, Angular/TypeScript and Python. This project
is Go + cgo, so the Go choice has to be recorded here.

## Decision

Use **[gremlins](https://github.com/go-gremlins/gremlins)** (`go-gremlins/gremlins`),
pinned to `v0.6.0`, invoked only through `make mutation-test`.

- One `gremlins unleash` run per production package (`./eapx`,
  `./internal/mcpserver`) — gremlins accepts a single path per invocation.
- Run with `--integration --workers 1`: the Cucumber suite lives in the separate
  `features` package, so a per-package coverage run sees the connector as
  uncovered. `--integration` runs the whole test binary per mutant, which
  restores that coverage; `--workers 1` plus a large `--timeout-coefficient`
  works around gremlins mislabelling healthy mutants `TIMED OUT` when the
  baseline suite finishes in well under a second.
- `tools/testkit mutation` merges the per-package gremlins JSON into the
  normalised `tmp/result/mutation-test.json`; `make mutation-test` then exits
  with gremlins' own status.

## Alternatives considered

### go-mutesting / avito-tech/go-mutesting

AST-based, older, less active. No first-class JSON report and weaker handling of
build-tag / cgo packages. Rejected.

### Skip mutation testing for Go

Violates the parent skill's Core Principle ("Mutation testing verifies testing
quality"). Rejected.

## Consequences

- `make mutation-test` is slow (~40 s here) because every mutant reruns the full
  suite serially. Acceptable for a small codebase and for CI's `ONLY_DELTA` mode
  (`--diff`).
- gremlins does not mutate the C in `eapx/cgo_mdb.go` (it is inside a cgo
  comment). The Go glue around it is mutated and covered by the in-package
  `eapx` tests. One defensively-coded fallback branch
  (`cgo_mdb.go` "empty engine message") is left as `NOT COVERED` — see
  [test-trace-matrix.md](../test-trace-matrix.md) row 17.
