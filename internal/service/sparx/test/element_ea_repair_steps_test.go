package sparxtest

// Finding and repairing elements whose EA base type doesn't match their
// ArchiMate type (element_ea_repair.feature) — the class of defect fixed in
// eaForElement (archimate.go): a model exported before that fix still has
// elements stored with the wrong uml:Class/uml:Interface/… base, so EA
// renders them as anonymous classes instead of their real ArchiMate shape.

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func registerElementEARepairSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I validate the model$`, func(ctx context.Context) error {
		w.Issues = w.Active().ValidateModel()
		common.Logf(ctx, "validate found %d issue(s)", len(w.Issues))
		return nil
	})

	sc.Step(`^I fix the elements:$`, func(ctx context.Context, table *godog.Table) error {
		var refs []string
		for _, row := range table.Rows {
			if len(row.Cells) == 1 {
				refs = append(refs, row.Cells[0].Value)
			}
		}
		issues, err := w.Mut.FixElements(refs)
		w.Issues, w.Err = issues, err
		if err != nil {
			common.Logf(ctx, "fix elements → error: %v", err)
			return nil
		}
		common.Logf(ctx, "fixed %d element(s)", len(issues))
		return w.Save(ctx)
	})

	sc.Step(`^I validate and fix the model$`, func(ctx context.Context) error {
		issues, err := w.Mut.ValidateAndFixModel()
		w.Issues, w.Err = issues, err
		if err != nil {
			common.Logf(ctx, "validate and fix → error: %v", err)
			return nil
		}
		common.Logf(ctx, "fixed %d element(s)", len(issues))
		return w.Save(ctx)
	})

	sc.Step(`^the (?:validation|fix) reports (\d+) issues?$`, func(ctx context.Context, n int) error {
		if len(w.Issues) != n {
			return fmt.Errorf("reported %d issue(s), want %d: %+v", len(w.Issues), n, w.Issues)
		}
		return nil
	})

	sc.Step(`^the validation reports exactly these issues:$`, func(ctx context.Context, table *godog.Table) error {
		rows, err := w.IssueRows()
		if err != nil {
			return err
		}
		return common.MatchTable(rows, table, true)
	})
}
