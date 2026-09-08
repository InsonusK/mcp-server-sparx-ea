package eapxtest

// Steps that assert on the outcome of a query. These are generic comparators:
// the expected data comes entirely from the .feature (a data table or a quoted
// string), the code here only compares.

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) theQuerySucceeds(ctx context.Context) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	logf(ctx, "query succeeded: %d row(s), columns %v", len(w.rs.Rows), w.rs.Columns)
	return nil
}

func (w *world) theQueryFailsWith(ctx context.Context, want string) error {
	if w.queryErr == nil {
		return fmt.Errorf("expected the query to fail with %q, but it returned %d row(s)", want, len(w.rs.Rows))
	}
	if !strings.Contains(w.queryErr.Error(), want) {
		return fmt.Errorf("query error %q does not contain %q", w.queryErr.Error(), want)
	}
	logf(ctx, "query failed as expected: %v", w.queryErr)
	return nil
}

func (w *world) noResultIsReturned(ctx context.Context) error {
	if w.rs != nil {
		return fmt.Errorf("expected no result set, got one with %d row(s)", len(w.rs.Rows))
	}
	logf(ctx, "no partial result was returned")
	return nil
}

// theResultIsExactly: the data table is the whole expected result. Header row =
// expected columns (exact order); the rest = expected rows, compared
// order-independently (the backend has no ORDER BY).
func (w *world) theResultIsExactly(ctx context.Context, table *godog.Table) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	wantCols, wantRows := tableData(table)

	if !slices.Equal(w.rs.Columns, wantCols) {
		return fmt.Errorf("columns: got %v, want %v", w.rs.Columns, wantCols)
	}
	if !maps.Equal(rowMultiset(w.rs.Rows), rowMultiset(wantRows)) {
		return fmt.Errorf("rows differ (order ignored):\n  got:  %s\n  want: %s",
			joinRows(w.rs.Rows), joinRows(wantRows))
	}
	logf(ctx, "result matched exactly: %d column(s), %d row(s)", len(wantCols), len(wantRows))
	return nil
}

// theResultContainsTheRows: every row in the table must appear in the result
// (order-independent, multiplicity-aware). The result may have more rows.
func (w *world) theResultContainsTheRows(ctx context.Context, table *godog.Table) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	wantCols, wantRows := tableData(table)
	if !slices.Equal(w.rs.Columns, wantCols) {
		return fmt.Errorf("columns: got %v, want %v", w.rs.Columns, wantCols)
	}
	got := rowMultiset(w.rs.Rows)
	for row, n := range rowMultiset(wantRows) {
		if got[row] < n {
			return fmt.Errorf("expected row [%s] at least %d time(s), found %d\n  actual rows: %s",
				strings.ReplaceAll(row, "\x1f", " | "), n, got[row], joinRows(w.rs.Rows))
		}
	}
	logf(ctx, "result contains all %d expected row(s)", len(wantRows))
	return nil
}

func (w *world) theResultHasNRows(ctx context.Context, n int) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	if len(w.rs.Rows) != n || w.rs.RowCount != n {
		return fmt.Errorf("expected %d row(s), got len(Rows)=%d RowCount=%d", n, len(w.rs.Rows), w.rs.RowCount)
	}
	logf(ctx, "result has %d row(s)", n)
	return nil
}

func (w *world) theResultIsEmpty(ctx context.Context) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	if len(w.rs.Rows) != 0 || w.rs.RowCount != 0 {
		return fmt.Errorf("expected an empty result, got len(Rows)=%d RowCount=%d: %s",
			len(w.rs.Rows), w.rs.RowCount, joinRows(w.rs.Rows))
	}
	logf(ctx, "result is empty (0 rows), columns still %v", w.rs.Columns)
	return nil
}

func (w *world) theResultColumnsAre(ctx context.Context, cols string) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	want := splitList(cols)
	if !slices.Equal(w.rs.Columns, want) {
		return fmt.Errorf("columns: got %v, want %v", w.rs.Columns, want)
	}
	logf(ctx, "result columns are %v", want)
	return nil
}

func registerResultSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the query succeeds$`, w.theQuerySucceeds)
	sc.Step(`^the query fails with "([^"]*)"$`, w.theQueryFailsWith)
	sc.Step(`^no result is returned$`, w.noResultIsReturned)
	sc.Step(`^the result is exactly:$`, w.theResultIsExactly)
	sc.Step(`^the result contains the rows:$`, w.theResultContainsTheRows)
	sc.Step(`^the result has (\d+) rows?$`, w.theResultHasNRows)
	sc.Step(`^the result is empty$`, w.theResultIsEmpty)
	sc.Step(`^the result columns are "([^"]*)"$`, w.theResultColumnsAre)
}
