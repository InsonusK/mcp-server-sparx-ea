package eapxtest

// Steps for features/concurrent_access.feature.
//
// The test is designed to FAIL if Connector's mutex is removed: goroutines read
// different elements at the same time, and an unsynchronised mdbtools handle
// lets one goroutine observe another's row.

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/cucumber/godog"
)

func (w *world) givenElementsToReadConcurrently(ctx context.Context, table *godog.Table) error {
	header, rows := tableData(table)
	if !slices.Equal(header, []string{"Object_ID", "Name", "Object_Type"}) {
		return fmt.Errorf("expected columns Object_ID, Name, Object_Type; got %v", header)
	}
	w.concExpect = nil
	for _, r := range rows {
		w.concExpect = append(w.concExpect, concTarget{id: r[0], wantRow: slices.Clone(r)})
	}
	logf(ctx, "%d distinct element(s) to read concurrently", len(w.concExpect))
	return nil
}

func (w *world) nGoroutinesRead(ctx context.Context, n int) error {
	if w.conn == nil {
		return fmt.Errorf("no open connection")
	}
	if len(w.concExpect) == 0 {
		return fmt.Errorf("no elements were set with the 'elements to read concurrently' step")
	}
	w.concGot = make([]concResult, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			target := w.concExpect[i%len(w.concExpect)]
			rs, err := w.conn.Query(
				"select Object_ID, Name, Object_Type from t_object where Object_ID = " + target.id)
			w.concGot[i] = concResult{target: target, rs: rs, err: err}
		}(i)
	}
	wg.Wait()
	logf(ctx, "%d goroutine(s) finished (round-robin over %d element(s))", n, len(w.concExpect))
	return nil
}

func (w *world) noGoroutineReturnedError(ctx context.Context) error {
	for i, g := range w.concGot {
		if g.err != nil {
			return fmt.Errorf("goroutine %d (Object_ID=%s) errored: %v", i, g.target.id, g.err)
		}
	}
	logf(ctx, "all %d goroutine(s) returned without error", len(w.concGot))
	return nil
}

func (w *world) everyGoroutineGotItsRow(ctx context.Context) error {
	for i, g := range w.concGot {
		if g.rs == nil {
			return fmt.Errorf("goroutine %d (Object_ID=%s) has no result", i, g.target.id)
		}
		if len(g.rs.Rows) != 1 {
			return fmt.Errorf("goroutine %d (Object_ID=%s): expected 1 row, got %d — %s",
				i, g.target.id, len(g.rs.Rows), joinRows(g.rs.Rows))
		}
		if !slices.Equal(g.rs.Rows[0], g.target.wantRow) {
			return fmt.Errorf("goroutine %d saw the wrong row: got %v, want %v (mutex removed?)",
				i, g.rs.Rows[0], g.target.wantRow)
		}
	}
	logf(ctx, "every goroutine got exactly its assigned row")
	return nil
}

func registerConcurrencySteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the elements to read concurrently:$`, w.givenElementsToReadConcurrently)
	sc.Step(`^(\d+) goroutines read those elements concurrently, round-robin$`, w.nGoroutinesRead)
	sc.Step(`^no goroutine returned an error$`, w.noGoroutineReturnedError)
	sc.Step(`^every goroutine got exactly its assigned row$`, w.everyGoroutineGotItsRow)
}
