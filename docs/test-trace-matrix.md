# Test Trace Matrix

Answers "what is covered by tests, and what is not" at a glance. Update it in the
same PR that adds or changes tests (see
[no-test-theater](../.claude/skills/no-test-theater/SKILL.md)).

- **Behaviour** is specified as Cucumber scenarios under `features/` and run by
  `go test ./features/` (each scenario is a Go subtest).
- **Plain Go tests** (`eapx/*_test.go`, `internal/mcpserver/*_test.go`) pin the
  cgo binding layer and pure helpers directly, and give mutation testing runnable
  targets inside each package.
- `make unit-test` runs both in one invocation, with `-race` and coverage.

Legend — Type: `happy` | `boundary` | `negative` | `error` | `concurrency` | `architecture`.
Status: ✅ covered | ⚠️ weak | ❌ not covered | 🔧 planned.

---

## Module: `eapx` — SQL connector (`Connector.Open` / `.Query` / `.Close`)

| # | Scenario (Given–When–Then) | Type | Test(s) | Assert checks | Status |
|---|----------------------------|------|---------|---------------|--------|
| 1 | Given a populated `.eapx`, When opening it, Then a connection is returned | happy | `open_file.feature` row `example/TestProject.eapx` | `openErr == nil` && connection non-nil | ✅ |
| 2 | Given an empty-project `.eapx`, When opening it, Then a connection is returned | happy | `open_file.feature` row `example/EmptyProject.eapx` | same | ✅ |
| 3 | Given a path to a missing file, When opening, Then it fails "file not found" and unwraps to `os.ErrNotExist` | error | `open_file.feature` rows `does-not-exist` / nonexistent dir; `TestOpenRejectsMissingAndEmpty` | error text contains `file not found`; `errors.Is(err, os.ErrNotExist)` | ✅ |
| 4 | Given an empty path, When opening, Then it fails "empty file path" | negative | `open_file.feature` blank row; `TestOpenRejectsMissingAndEmpty` | error text contains `empty file path` | ✅ |
| 5 | Given a non-JET file, When opening, Then it fails "Unable to locate database" | error | `open_file.feature` row `go.mod` | error text contains `Unable to locate database` | ✅ |
| 6 | Given a connection, When running a `SELECT`, Then columns, rows and `RowCount` match | happy | `run_query.feature` outline; `TestConnectorAgainstFixture` | exact `Columns`, `len(Rows)`, `RowCount`, row values | ✅ |
| 7 | Given a connection, When a `SELECT` matches nothing, Then `RowCount == 0` and `Rows` empty | boundary | `run_query.feature` row `999999`; `TestConnectorAgainstFixture` | `RowCount == 0 && len(Rows) == 0` | ✅ |
| 8 | Given an aggregate query, When run, Then the single computed value is returned | happy | `run_query.feature` "Aggregate query" | one row / one col == `16` | ✅ |
| 9 | Given an empty project, When selecting objects, Then zero rows | boundary | `run_query.feature` "empty project" | `RowCount == 0` | ✅ |
| 10 | Given a query for a missing table / bad column / syntax error, When run, Then it fails "Got no result" | error | `query_errors.feature` outline; `TestConnectorAgainstFixture` "engine error" | error text contains `Got no result`; error is `*eapx.Error{Op:"query"}`; message echoes the SQL | ✅ |
| 11 | Given a write/DDL statement (`insert/update/delete/drop/...`), When run, Then it is rejected before the engine | negative | `query_errors.feature` rows; `TestWriteVerbsRejectedBeforeEngine` | `*eapx.Error.msg == "only read-only SELECT statements are supported"`; engine untouched (nil-engine connector) | ✅ |
| 12 | Given empty SQL, When run, Then it fails "empty SQL statement" | negative | `query_errors.feature` blank row | error text contains `empty SQL statement` | ✅ |
| 13 | Given a closed connection, When querying, Then it fails "connector is closed" | error | `query_errors.feature` "closed connection"; `TestConnectorAgainstFixture` "after close" | error text contains `connector is closed` | ✅ |
| 14 | Given one connection, When 40 goroutines query it concurrently, Then all succeed with the same value and `-race` is clean | concurrency | `concurrent_access.feature` | 40/40 `err == nil`; every result value == `16`; suite run under `-race` | ✅ |
| 15 | Given the low-level engine wrapper with no handle, When query/close is called (twice), Then it errors / no-ops without panic | negative | `TestMdbSQLZeroValueIsSafe` | `query` returns error; `close` twice does not panic | ✅ |
| 16 | Given `firstWord`, When given padded / parenthesised / `;`-led / empty SQL, Then the leading verb is lower-cased or "" | boundary | `TestFirstWord` | exact string for 9 inputs | ✅ |
| 17 | Given a JET file whose open fails with no engine message, Then a fallback message is used | error | — | — | ❌ mdbtools always sets a message on open failure; the `msg == ""` fallback (`cgo_mdb.go:75`) is unreachable defensively-coded and left as the single surviving mutant (97.1% score). |

## Module: `eapx` — text decoding (architectural, base-task DoD)

| # | Scenario | Type | Test(s) | Assert checks | Status |
|---|----------|------|---------|---------------|--------|
| 18 | Given a query over text/memo columns, When run, Then every value is valid UTF-8 | architecture | `text_encoding.feature` "valid UTF-8" | `utf8.ValidString` for every cell | ✅ |
| 19 | Given a `.eapx` containing Cyrillic, When selecting it, Then it round-trips without mojibake | architecture | `text_encoding.feature` `@needs-cyrillic-fixture` | single value == `Кириллица объекта` | ❌ **skipped** — authoring a Cyrillic `.eapx` needs Sparx EA (mdbtools is read-only, cannot `INSERT`). Scenario self-skips via `godog.ErrSkip` until `example/CyrillicProject.eapx` is added. JET4 UCS-2LE → UTF-8 path is exercised by #18; the CP1251/JET3 path is not. |

## Module: `eapx` — in-process execution (architectural, base-task DoD)

| # | Scenario | Type | Test(s) | Assert checks | Status |
|---|----------|------|---------|---------------|--------|
| 20 | No project package imports `os/exec` | architecture | `in_process_execution.feature` | `go list -deps` over `github.com/InsonusK/mcp-server-sparx-ea/...` yields no `os/exec` importer | ✅ |
| 21 | Given `TMPDIR` set to an empty dir, When 25 queries run, Then the dir stays empty | architecture | `in_process_execution.feature` | `os.ReadDir(tmpdir)` empty afterwards | ✅ |
| 22 | The compiled binary is linked directly against `libmdb` | architecture | `in_process_execution.feature` | `ldd <binary>` output contains `libmdb` | ✅ |

## Module: `internal/mcpserver` — the `ea_query` MCP tool

| # | Scenario (Given–When–Then) | Type | Test(s) | Assert checks | Status |
|---|----------------------------|------|---------|---------------|--------|
| 23 | Given a running server, When listing tools, Then `ea_query` is advertised (and it is the only tool) | happy | `mcp_tool.feature`; `TestNewServesEaQueryEndToEnd` | tool name `ea_query` present; exactly one tool | ✅ |
| 24 | Given valid `file`+`sql`, When calling `ea_query`, Then a non-error result with `{columns,rows,rowCount}` JSON | happy | `mcp_tool.feature`; `TestHandlerReturnsRowsAsJSON`; `TestNewServesEaQueryEndToEnd` | `IsError == false`; JSON contains `"columns":[...]`, `"rowCount":16`/`2`, a known cell value | ✅ |
| 25 | Given a query with no rows, When calling `ea_query`, Then `rows` serialises as `[]` not `null` | boundary | `TestHandlerEmptyRowsSerializeAsArray` | text contains `"rows":[]` | ✅ |
| 26 | Given a bad `file`, When calling `ea_query`, Then a tool error carrying the open failure | error | `mcp_tool.feature`; `TestHandlerOpenErrorBecomesToolError` | `IsError == true`; text contains `file not found` | ✅ |
| 27 | Given a write statement, When calling `ea_query`, Then a tool error "only read-only SELECT statements are supported" | error | `mcp_tool.feature`; `TestHandlerQueryErrorBecomesToolError` | `IsError == true`; text contains the message | ✅ |
| 28 | Given a missing required argument (`file` or `sql`), When calling `ea_query`, Then a tool error | negative | `TestHandlerMissingArguments` | `IsError == true` for both cases | ✅ |
| 29 | Given a successful call, When it returns, Then the connector was closed | happy | `TestHandlerReturnsRowsAsJSON` | `fakeConn.closed == true` | ✅ |

---

## Module summary

| Type | Total | ✅ | ⚠️ | ❌ |
|------|-------|----|----|----|
| happy | 9 | 9 | 0 | 0 |
| boundary | 5 | 5 | 0 | 0 |
| negative | 5 | 5 | 0 | 0 |
| error | 7 | 6 | 0 | 1 (#17 — unreachable defensive fallback) |
| concurrency | 1 | 1 | 0 | 0 |
| architecture | 5 | 4 | 0 | 1 (#19 — Cyrillic fixture unavailable, scenario self-skips) |

Both `❌` rows carry an explicit justification above and are the two mutants gremlins reports as `NOT COVERED` / the one skipped scenario. No happy/negative row is uncovered.

---

## Function → Test (reverse breakdown)

| Public method / branch | Covering test(s) |
|------------------------|------------------|
| `eapx.Open` — empty path | `open_file.feature` blank row, `TestOpenRejectsMissingAndEmpty` |
| `eapx.Open` — `os.Stat` failure | `open_file.feature` missing rows, `TestOpenRejectsMissingAndEmpty` |
| `eapx.Open` — `mdbOpen` failure | `open_file.feature` `go.mod` row |
| `eapx.Open` — success | every `run_query.feature` scenario |
| `Connector.Query` — empty SQL | `query_errors.feature` blank row |
| `Connector.Query` — write-verb rejection | `query_errors.feature` write rows, `TestWriteVerbsRejectedBeforeEngine` |
| `Connector.Query` — closed connector | `query_errors.feature` closed, `TestConnectorAgainstFixture` |
| `Connector.Query` — engine error | `query_errors.feature` outline, `TestConnectorAgainstFixture` |
| `Connector.Query` — success + result mapping | `run_query.feature`, `TestConnectorAgainstFixture` |
| `Connector.Close` — first call / idempotent second | `TestConnectorAgainstFixture`, `TestMdbSQLZeroValueIsSafe` |
| `mdbSQL.query` / `mdbSQL.close` — nil handle | `TestMdbSQLZeroValueIsSafe` |
| `firstWord` — all branches | `TestFirstWord` |
| `Error.Error` — with/without SQL | `TestConnectorAgainstFixture` ("engine error"), every error-path scenario |
| `Error.Unwrap` | `TestOpenRejectsMissingAndEmpty` (`errors.Is`) |
| `mcpserver.New` — nil opener default | `mcp_tool.feature`, `TestNewServesEaQueryEndToEnd` |
| `mcpserver.defaultOpener` | `TestNewServesEaQueryEndToEnd` |
| `handleEAQuery` — missing args / open err / query err / success / nil rows / close | `server_test.go` tests + `mcp_tool.feature` |

## Weakened tests

None. No existing test's assertions, timeouts or skip state were reduced in this change.
