# Test Trace Matrix

Answers "what is covered by tests, and what is not" at a glance. Update it in the
same PR that adds or changes tests (see
[no-test-theater](../.claude/skills/no-test-theater/SKILL.md)).

**Every test is a Cucumber scenario** (per
[solution-conformance-testing](../.claude/skills/solution-conformance-testing/SKILL.md)).
Each package owns its spec — the `.feature` files and their godog step
definitions sit next to the code they exercise:

| Package | Features | Step defs (`_test` package) |
|---------|----------|------------------------------|
| `client/eapx` | `client/eapx/features/` | `client/eapx/steps_test.go` |
| `internal/mcpserver` | `internal/mcpserver/features/` | `internal/mcpserver/steps_test.go` |
| `features` (project-wide) | `features/project_architecture.feature` | `features/steps_test.go` |

`make unit-test` runs all three in one `go test ./...`, with `-race` and
coverage. Each scenario is a Go subtest. There are **no plain Go unit tests** —
code that the public API could not reach through a scenario (nil-handle guards
on the cgo wrapper, an `mdb_sql_init` OOM check) was removed rather than
unit-tested.

Legend — Type: `happy` | `boundary` | `negative` | `error` | `concurrency` | `architecture`.
Status: ✅ covered | ⚠️ weak | ❌ not covered | ⏳ excluded from the run (`@todo`).

---

## Module: `client/eapx` — SQL connector (`Open` / `Query` / `Close`)

| # | Scenario (Given–When–Then) | Type | Feature | Assert checks | Status |
|---|----------------------------|------|---------|---------------|--------|
| 1 | Given a populated / empty-project `.eapx`, When opening it, Then a connection is returned | happy | `open_file.feature` rows `TestProject` / `EmptyProject` | `openErr == nil` && connection non-nil | ✅ |
| 2 | Given a missing file, When opening, Then it fails "file not found" | error | `open_file.feature` missing rows | error text contains `file not found` | ✅ |
| 3 | Given an empty path, When opening, Then it fails "empty file path" | negative | `open_file.feature` blank row | error text contains `empty file path` | ✅ |
| 4 | Given a non-JET file, When opening, Then it fails "Unable to locate database" | error | `open_file.feature` row `go.mod` | error text contains `Unable to locate database` (also kills the `msg == ""` fallback mutant in `mdbOpen`) | ✅ |
| 5 | Given a connection, When running a `SELECT`, Then columns, rows and `RowCount` match | happy | `run_query.feature` outline | exact `Columns`, `len(Rows)`, `RowCount` | ✅ |
| 6 | Given a `SELECT` that matches nothing, Then `RowCount == 0` and `Rows` empty | boundary | `run_query.feature` row `999999` | `RowCount == 0 && len(Rows) == 0` | ✅ |
| 7 | Given an aggregate query, Then the single computed value is returned | happy | `run_query.feature` "Aggregate query" | one row / one col == `16` | ✅ |
| 8 | Given a specific-id query, Then the exact row is returned | happy | `run_query.feature` "Reading a specific element by id" | `row 0 == [2, Stakeholder1, Class]` | ✅ |
| 9 | Given an empty project, When selecting objects, Then zero rows | boundary | `run_query.feature` "empty project" | `RowCount == 0` | ✅ |
| 10 | Given a SELECT in any letter case, Then it runs | boundary | `run_query.feature` "letter case" outline | `SELECT` / `SeLeCt` both return 16 rows | ✅ |
| 11 | Given a SELECT with leading whitespace, Then it runs | boundary | `run_query.feature` "leading whitespace" | 16 rows (covers the trim + first-token logic in `Query`) | ✅ |
| 12 | Given a query for a missing table / bad column / syntax error, Then it fails "Got no result" | error | `query_errors.feature` outline | error text contains `Got no result` | ✅ |
| 13 | Given a write/DDL statement (`delete/update/drop/…`), Then it fails "only read-only SELECT statements are supported" | negative | `query_errors.feature` outline rows | that exact message; case-insensitive verb match | ✅ |
| 14 | Given empty SQL, Then it fails "empty SQL statement" | negative | `query_errors.feature` blank row | error text contains `empty SQL statement` | ✅ |
| 15 | Given a failing query, Then the error string echoes the offending SQL | error | `query_errors.feature` "error echoes the offending SQL" | error text contains `(sql: <the sql>)` (kills the `e.SQL != ""` mutant) | ✅ |
| 16 | Given a closed connection, When querying, Then it fails "connector is closed" | error | `query_errors.feature` "closed connection" | error text contains `connector is closed` | ✅ |
| 17 | Given one connection, When 40 goroutines query it concurrently, Then all succeed with the same value; `-race` clean | concurrency | `concurrent_access.feature` | 40/40 `err == nil`; every value == `16`; suite under `-race` | ✅ |

## Module: `client/eapx` — architectural (base-task DoD)

| # | Scenario | Type | Feature | Assert checks | Status |
|---|----------|------|---------|---------------|--------|
| 18 | Given a query over text/memo columns, Then every value is valid UTF-8 | architecture | `text_encoding.feature` "valid UTF-8" | `utf8.ValidString` for every cell | ✅ |
| 19 | Given a `.eapx` containing Cyrillic, Then it round-trips without mojibake | architecture | `text_encoding.feature` `@todo` | single value == `Кириллица объекта` | ⏳ **excluded** via the `~@todo` tag filter — no fixture. Authoring a Cyrillic `.eapx` needs Sparx EA (mdbtools is read-only, cannot `INSERT`), and neither sample project has non-ASCII text. Add `example/CyrillicProject.eapx` and drop the tag. JET4 UCS-2LE→UTF-8 is covered by #18; CP1251/JET3 is not. |
| 20 | Given `TMPDIR` set to an empty dir, When 25 queries run, Then the dir stays empty | architecture | `in_process_execution.feature` | `os.ReadDir(tmpdir)` empty afterwards | ✅ |

## Module: `internal/mcpserver` — the `ea_query` MCP tool

| # | Scenario (Given–When–Then) | Type | Feature | Assert checks | Status |
|---|----------------------------|------|---------|---------------|--------|
| 21 | Given a running server, When listing tools, Then `ea_query` is the only tool | happy | `ea_query_tool.feature` "advertises exactly the ea_query tool" | tool name `ea_query` present; `len(tools) == 1` | ✅ |
| 22 | Given valid `file`+`sql`, When calling `ea_query`, Then a non-error result with `{columns,rows,rowCount}` JSON | happy | `ea_query_tool.feature` "returns query rows as JSON" | `IsError == false`; JSON has `"rowCount":16`, `columns`, a known cell | ✅ |
| 23 | Given a query with no rows, Then `rows` is `[]`, not `null` | boundary | `ea_query_tool.feature` "empty row list, not null" | JSON has `"rows":[]` and `"rowCount":0` | ✅ |
| 24 | Given a bad `file`, Then a tool error carrying the open failure | error | `ea_query_tool.feature` "surfaces an open failure" | `IsError == true`; text contains `file not found` | ✅ |
| 25 | Given a write statement, Then a tool error "only read-only SELECT statements are supported" | error | `ea_query_tool.feature` "rejects a write statement" | `IsError == true`; text contains the message | ✅ |
| 26 | Given only `file` or only `sql`, Then a tool error naming the missing argument | negative | `ea_query_tool.feature` "needs both file and sql" outline | `IsError == true`; text contains `missing required argument: sql` / `: file` | ✅ |

## Module: `features` — project-wide architecture

| # | Scenario | Type | Feature | Assert checks | Status |
|---|----------|------|---------|---------------|--------|
| 27 | No first-party package imports `os/exec` | architecture | `project_architecture.feature` | `go list -deps` over `github.com/InsonusK/mcp-server-sparx-ea/...` yields no `os/exec` importer | ✅ |
| 28 | The compiled binary is linked directly against `libmdb` | architecture | `project_architecture.feature` | `ldd <binary>` output contains `libmdb` | ✅ |

---

## Module summary

| Type | Total | ✅ | ⚠️ | ❌ / ⏳ |
|------|-------|----|----|--------|
| happy | 6 | 6 | 0 | 0 |
| boundary | 6 | 6 | 0 | 0 |
| negative | 4 | 4 | 0 | 0 |
| error | 7 | 7 | 0 | 0 |
| concurrency | 1 | 1 | 0 | 0 |
| architecture | 5 | 4 | 0 | 1 ⏳ (#19 — Cyrillic fixture unavailable) |

No happy/negative/error scenario is uncovered. The one ⏳ row is tag-excluded
from the run (never a fake green) and carries an explicit justification.

Mutation score: **100%** (27 mutants, all killed, 0 survived, 0 not covered) —
`make mutation-test`, gremlins over `./client/eapx` and `./internal/mcpserver`.

---

## Function → Scenario (reverse breakdown)

| Public method / branch | Covering scenario(s) |
|------------------------|----------------------|
| `eapx.Open` — empty path | #3 |
| `eapx.Open` — `os.Stat` failure | #2 |
| `eapx.Open` — `mdbOpen` failure + empty-message fallback | #4 |
| `eapx.Open` — success | #1, and every `run_query` scenario |
| `Connector.Query` — empty / whitespace-only SQL | #14 |
| `Connector.Query` — first-token + case-fold + write-verb map | #10, #11, #13 |
| `Connector.Query` — closed connector | #16 |
| `Connector.Query` — engine error, with SQL echoed | #12, #15 |
| `Connector.Query` — success + result mapping | #5–#9 |
| `Connector.Close` — closes the live engine | every scenario's teardown; #16 proves the effect |
| `mdbSQL.query` / `mdbSQL.close` | driven by every `Connector.Query` / `Close` (#5–#17) |
| `Error.Error` — with / without SQL | #15 (with), #2/#3/#14/#16 (without) |
| `mcpserver.New` — nil-opener default → real connector | every `ea_query_tool` scenario (server built with `New(nil)`) |
| `mcpserver.defaultOpener` | #22–#25 |
| `handleEAQuery` — missing args / open err / query err / success / empty rows | #22–#26 |

## Weakened tests

None. The Cyrillic scenario changed from a self-skipping step (which still
reported `pass` to `go test`) to a `@todo`-tagged scenario excluded at
selection time — it is now honestly absent from the run rather than a fake
green.
