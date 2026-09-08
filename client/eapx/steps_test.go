package eapx_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/bddsupport"
)

type world struct {
	pathArg  string
	conn     *eapx.Connector
	openErr  error
	rs       *eapx.ResultSet
	queryErr error

	// temp-dir isolation
	prevTMPDIR    string
	hadTMPDIR     bool
	tmpDir        string
	tmpDirApplied bool

	// concurrency
	concN    int
	concErrs []error
	concVals []string
}

func (w *world) reset() {
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.restoreTMPDIR()
	*w = world{}
}

func (w *world) restoreTMPDIR() {
	if !w.tmpDirApplied {
		return
	}
	if w.hadTMPDIR {
		_ = os.Setenv("TMPDIR", w.prevTMPDIR)
	} else {
		_ = os.Unsetenv("TMPDIR")
	}
	w.tmpDirApplied = false
}

// ---------- open ----------

func (w *world) theFilePath(path string) error {
	w.pathArg = bddsupport.ResolvePath(path)
	return nil
}

func (w *world) iOpenTheConnection() error {
	w.conn, w.openErr = eapx.Open(w.pathArg)
	return nil
}

func (w *world) openingOutcome(outcome string) error {
	switch {
	case outcome == "succeeds":
		if w.openErr != nil {
			return fmt.Errorf("expected open to succeed, got error: %v", w.openErr)
		}
		if w.conn == nil {
			return errors.New("expected a connection, got nil")
		}
		return nil
	case strings.HasPrefix(outcome, "fails with an error containing "):
		want := strings.Trim(strings.TrimPrefix(outcome, "fails with an error containing "), `"`)
		return assertErrContains(w.openErr, want)
	default:
		return fmt.Errorf("unknown outcome %q", outcome)
	}
}

func (w *world) aConnectionTo(path string) error {
	conn, err := eapx.Open(bddsupport.ResolvePath(path))
	if err != nil {
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.conn = conn
	return nil
}

// ---------- query ----------

func (w *world) iRunTheQuery(sql string) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	w.rs, w.queryErr = w.conn.Query(sql)
	return nil
}

func (w *world) iRunTheQueryNTimes(sql string, n int) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	for i := 0; i < n; i++ {
		if _, err := w.conn.Query(sql); err != nil {
			return fmt.Errorf("query %d failed: %w", i, err)
		}
	}
	return nil
}

func (w *world) iCloseTheConnection() error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	return w.conn.Close()
}

func (w *world) theQuerySucceeds() error {
	if w.queryErr != nil {
		return fmt.Errorf("expected query to succeed, got: %v", w.queryErr)
	}
	if w.rs == nil {
		return errors.New("expected a result set, got nil")
	}
	return nil
}

func (w *world) theQueryFailsContaining(msg string) error {
	if w.rs != nil {
		return fmt.Errorf("expected query to fail, but it returned %d rows", w.rs.RowCount)
	}
	return assertErrContains(w.queryErr, msg)
}

func (w *world) theResultHasNRows(n int) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if got := len(w.rs.Rows); got != n {
		return fmt.Errorf("expected %d rows, got %d", n, got)
	}
	if w.rs.RowCount != n {
		return fmt.Errorf("RowCount=%d disagrees with len(Rows)=%d", w.rs.RowCount, n)
	}
	return nil
}

func (w *world) theResultColumnsAre(cols string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	want := splitCSV(cols)
	if !equalStrings(w.rs.Columns, want) {
		return fmt.Errorf("expected columns %v, got %v", want, w.rs.Columns)
	}
	return nil
}

func (w *world) rowEquals(idx int, csv string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if idx < 0 || idx >= len(w.rs.Rows) {
		return fmt.Errorf("row %d out of range (%d rows)", idx, len(w.rs.Rows))
	}
	want := splitCSV(csv)
	if !equalStrings(w.rs.Rows[idx], want) {
		return fmt.Errorf("row %d: expected %v, got %v", idx, want, w.rs.Rows[idx])
	}
	return nil
}

func (w *world) theSingleResultValueIs(val string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if len(w.rs.Rows) != 1 || len(w.rs.Rows[0]) != 1 {
		return fmt.Errorf("expected exactly one row with one column, got %d rows / columns %v", len(w.rs.Rows), w.rs.Columns)
	}
	if w.rs.Rows[0][0] != val {
		return fmt.Errorf("expected value %q, got %q", val, w.rs.Rows[0][0])
	}
	return nil
}

func (w *world) everyResultValueIsValidUTF8() error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	for r, row := range w.rs.Rows {
		for c, cell := range row {
			if !utf8.ValidString(cell) {
				return fmt.Errorf("row %d col %d (%s) is not valid UTF-8: %q", r, c, w.rs.Columns[c], cell)
			}
		}
	}
	return nil
}

// ---------- in-process execution: temp dir ----------

func (w *world) emptyDirAsTempDir() error {
	dir, err := os.MkdirTemp("", "eapx-tmpdir-*")
	if err != nil {
		return err
	}
	w.prevTMPDIR, w.hadTMPDIR = os.LookupEnv("TMPDIR")
	if err := os.Setenv("TMPDIR", dir); err != nil {
		return err
	}
	w.tmpDir = dir
	w.tmpDirApplied = true
	return nil
}

func (w *world) theTempDirIsStillEmpty() error {
	entries, err := os.ReadDir(w.tmpDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		return fmt.Errorf("temp dir is not empty: %v", names)
	}
	return nil
}

// ---------- concurrency ----------

func (w *world) iRunQueryFromNGoroutines(sql string, n int) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	w.concN = n
	w.concErrs = make([]error, n)
	w.concVals = make([]string, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			rs, err := w.conn.Query(sql)
			if err != nil {
				w.concErrs[i] = err
				return
			}
			if len(rs.Rows) == 1 && len(rs.Rows[0]) == 1 {
				w.concVals[i] = rs.Rows[0][0]
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func (w *world) allNQueriesSucceed(n int) error {
	if w.concN != n {
		return fmt.Errorf("ran %d goroutines, step expects %d", w.concN, n)
	}
	for i, err := range w.concErrs {
		if err != nil {
			return fmt.Errorf("goroutine %d failed: %w", i, err)
		}
	}
	return nil
}

func (w *world) everyConcurrentQueryReturned(val string) error {
	for i, got := range w.concVals {
		if got != val {
			return fmt.Errorf("goroutine %d returned %q, want %q", i, got, val)
		}
	}
	return nil
}

// ---------- helpers ----------

func assertErrContains(err error, want string) error {
	if err == nil {
		return fmt.Errorf("expected an error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", err.Error(), want)
	}
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------- registration ----------

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	sc.Step(`^the file path "([^"]*)"$`, w.theFilePath)
	sc.Step(`^I open the connection$`, w.iOpenTheConnection)
	sc.Step(`^opening (.+)$`, w.openingOutcome)
	sc.Step(`^a connection to the Sparx EA file "([^"]*)"$`, w.aConnectionTo)

	sc.Step(`^I run the query "([^"]*)"$`, w.iRunTheQuery)
	sc.Step(`^I run the query "([^"]*)" (\d+) times$`, w.iRunTheQueryNTimes)
	sc.Step(`^I close the connection$`, w.iCloseTheConnection)
	sc.Step(`^the query succeeds$`, w.theQuerySucceeds)
	sc.Step(`^the query fails with an error containing "([^"]*)"$`, w.theQueryFailsContaining)
	sc.Step(`^the result has (\d+) rows$`, w.theResultHasNRows)
	sc.Step(`^the result columns are "([^"]*)"$`, w.theResultColumnsAre)
	sc.Step(`^row (\d+) equals "([^"]*)"$`, w.rowEquals)
	sc.Step(`^the single result value is "([^"]*)"$`, w.theSingleResultValueIs)
	sc.Step(`^every result value is valid UTF-8$`, w.everyResultValueIsValidUTF8)

	sc.Step(`^an empty directory registered as the process temp dir$`, w.emptyDirAsTempDir)
	sc.Step(`^the temp dir is still empty$`, w.theTempDirIsStillEmpty)

	sc.Step(`^I run the query "([^"]*)" from (\d+) goroutines concurrently$`, w.iRunQueryFromNGoroutines)
	sc.Step(`^all (\d+) queries succeed$`, w.allNQueriesSucceed)
	sc.Step(`^every concurrent query returned the value "([^"]*)"$`, w.everyConcurrentQueryReturned)
}
