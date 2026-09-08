// Package eapxtest is the black-box Cucumber suite for the eapx connector.
// It imports client/eapx and exercises only its public API.
//
// Conventions: docs/skills/cucumber-go-testing.md
//   - every step takes ctx as its first arg and logs what it did via logf
//   - assertion steps are generic comparators; the expected data lives in the
//     .feature (usually a data table), never hard-coded here
//   - fixtures live in testdata/, not in the repo-wide example/ directory
package eapxtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
)

// world is the mutable state of one scenario. It is created once per scenario in
// suite_test.go and cleared by reset in the Before/After hooks.
type world struct {
	conn     *eapx.Connector
	openErr  error
	rs       *eapx.ResultSet
	queryErr error

	// architecture: on-disk state / TMPDIR isolation
	recordedPath string
	recordedStat os.FileInfo
	tmpDir       string
	prevTMPDIR   string
	hadTMPDIR    bool
	tmpDirSet    bool

	// concurrency
	concExpect []concTarget // id -> expected row, indexed round-robin
	concGot    []concResult
}

type concTarget struct {
	id      string
	wantRow []string
}

type concResult struct {
	target concTarget
	rs     *eapx.ResultSet
	err    error
}

func newWorld() *world { return &world{} }

func (w *world) reset() {
	if w.conn != nil {
		_ = w.conn.Close()
	}
	if w.tmpDirSet {
		if w.hadTMPDIR {
			_ = os.Setenv("TMPDIR", w.prevTMPDIR)
		} else {
			_ = os.Unsetenv("TMPDIR")
		}
	}
	*w = world{}
}

// ---- shared helpers ----

// logf records what a step did. It writes straight to stdout (not t.Log) so the
// line interleaves with godog's pretty tree and so VS Code's test terminal
// attributes it to the running scenario rather than to godog's internals. Only
// shown by `go test -v` (or on failure) — the repo's .vscode config passes -v.
// ctx is kept in the signature for future use.
func logf(_ context.Context, format string, a ...any) {
	fmt.Printf("      · "+format+"\n", a...)
}

// fixturePath maps a name written in a .feature to a real path. "" stays "" (so
// the "empty file path" case reaches the connector unchanged); an absolute path
// stays; anything else is resolved under testdata/.
func fixturePath(name string) string {
	if name == "" || filepath.IsAbs(name) {
		return name
	}
	return filepath.Join("testdata", name)
}

// ensureQuerySucceeded is the internal check the comparator steps share. The
// user-facing "the query succeeds" step wraps this and logs.
func (w *world) ensureQuerySucceeded() error {
	if w.queryErr != nil {
		return fmt.Errorf("expected the query to succeed, got error: %v", w.queryErr)
	}
	if w.rs == nil {
		return fmt.Errorf("expected a result set, got nil")
	}
	return nil
}

// tableData splits a godog data table into its header row and the remaining
// data rows.
func tableData(t *godog.Table) (header []string, rows [][]string) {
	for i, r := range t.Rows {
		cells := make([]string, len(r.Cells))
		for j, c := range r.Cells {
			cells[j] = c.Value
		}
		if i == 0 {
			header = cells
			continue
		}
		rows = append(rows, cells)
	}
	return header, rows
}

// rowMultiset counts identical rows, so comparisons are order-independent — the
// mdbtools backend has no ORDER BY and returns rows in storage order.
func rowMultiset(rows [][]string) map[string]int {
	m := make(map[string]int, len(rows))
	for _, r := range rows {
		m[strings.Join(r, "\x1f")]++
	}
	return m
}

func joinRows(rows [][]string) string {
	parts := make([]string, len(rows))
	for i, r := range rows {
		parts[i] = "[" + strings.Join(r, " | ") + "]"
	}
	return strings.Join(parts, " ")
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
