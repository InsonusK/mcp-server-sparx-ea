package eapxtest

// Steps for running queries. Shared by run_query / query_errors / text_encoding
// / in_process_execution.

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

func (w *world) iRunTheQuery(ctx context.Context, sql string) error {
	if w.conn == nil {
		return fmt.Errorf("no open connection")
	}
	w.rs, w.queryErr = w.conn.Query(sql)
	if w.queryErr != nil {
		logf(ctx, "query %q → error: %v", sql, w.queryErr)
		return nil
	}
	logf(ctx, "query %q → %d row(s), columns %v", sql, len(w.rs.Rows), w.rs.Columns)
	return nil
}

func (w *world) iRunTheQueryNTimes(ctx context.Context, sql string, n int) error {
	if w.conn == nil {
		return fmt.Errorf("no open connection")
	}
	for i := 0; i < n; i++ {
		rs, err := w.conn.Query(sql)
		if err != nil {
			return fmt.Errorf("run %d of %q failed: %w", i+1, sql, err)
		}
		w.rs, w.queryErr = rs, nil
	}
	logf(ctx, "ran %q %d times, last run → %d row(s)", sql, n, len(w.rs.Rows))
	return nil
}

func registerQuerySteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I run the query "([^"]*)"$`, w.iRunTheQuery)
	sc.Step(`^I run the query "([^"]*)" (\d+) times$`, w.iRunTheQueryNTimes)
}
