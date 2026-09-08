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

- One `gremlins unleash` run per production package (`./client/eapx`,
  `./internal/mcpserver`) — gremlins accepts a single path per invocation.
- Run with `--integration --workers 1 --timeout-coefficient 60`: each package's
  only tests are godog scenarios in an external `_test` package, so a plain
  per-package `go test` sees the code as uncovered. `--integration` runs the
  whole test binary per mutant, which restores the attribution. `--workers 1`
  plus the large timeout coefficient work around gremlins mislabelling healthy
  mutants `TIMED OUT` when the baseline suite finishes in well under a second.
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
- gremlins does not mutate the C in `client/eapx/cgo_mdb.go` (it is inside a cgo
  comment). The Go glue around it is mutated and fully killed (27/27 mutants,
  100% score) by the Cucumber suite alone.
