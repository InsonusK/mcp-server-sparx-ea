package sparxtest

// Actions on relationships: create / delete (method 5).

import (
	"context"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

// helperElements is the fixed set of fresh elements the relationship scenarios
// connect. They are created fresh in every scenario (not taken from the fixture)
// so a scenario never trips the "relationship already exists" guard on a
// relationship the fixture already carries.
var helperElements = []struct{ typ, name string }{
	{"ArchiMate.Goal", "GoalA"},
	{"ArchiMate.Goal", "GoalB"},
	{"ArchiMate.Requirement", "ReqA"},
	{"ArchiMate.BusinessProcess", "ProcA"},
	{"ArchiMate.BusinessProcess", "ProcB"},
	{"ArchiMate.BusinessObject", "ObjA"},
}

func registerRelationshipSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^the relationship helper elements in "([^"]*)"$`, func(ctx context.Context, pkg string) error {
		for _, h := range helperElements {
			if _, err := w.Mut.CreateElement(pkg, h.typ, h.name, ""); err != nil {
				w.Err = err
				return err
			}
		}
		common.Logf(ctx, "created %d helper elements in %q", len(helperElements), pkg)
		return w.Save(ctx)
	})

	relate := func(ctx context.Context, src, tgt, rel, name string) error {
		r, err := w.Mut.CreateRelationship(src, tgt, rel, name, "")
		w.Err = err
		if err != nil {
			common.Logf(ctx, "relate %q -> %q as %s → error: %v", src, tgt, rel, err)
			return nil
		}
		w.LastRelID, w.LastRel = r.ID, r
		common.Logf(ctx, "created %s %q -> %q (id %s, verdict %q)", rel, src, tgt, r.ID, r.Verdict)
		return w.Save(ctx)
	}

	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)"$`, func(ctx context.Context, s, t, r string) error {
		return relate(ctx, s, t, r, "")
	})
	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)" named "([^"]*)"$`, relate)

	// Table form: one row per relationship, all within one package.
	//   When I relate, in "<pkg>":
	//     | source | target | relation              | name |
	//     | ReqA   | GoalA  | ArchiMate.Realization |      |
	sc.Step(`^I relate, in "([^"]*)":$`, func(ctx context.Context, pkg string, table *godog.Table) error {
		if len(table.Rows) < 2 {
			return errString("the relate table needs a header and at least one row")
		}
		cols := map[string]int{}
		for i, c := range table.Rows[0].Cells {
			cols[c.Value] = i
		}
		for _, key := range []string{"source", "target", "relation"} {
			if _, ok := cols[key]; !ok {
				return errString("the relate table needs a " + key + " column")
			}
		}
		w.Err = nil
		for _, row := range table.Rows[1:] {
			cell := func(name string) string {
				if i, ok := cols[name]; ok && i < len(row.Cells) {
					return row.Cells[i].Value
				}
				return ""
			}
			src := pkg + "/" + cell("source")
			tgt := pkg + "/" + cell("target")
			if err := relate(ctx, src, tgt, cell("relation"), cell("name")); err != nil {
				return err
			}
			if w.Err != nil {
				return nil // stop at the first rejected relationship, keep w.Err for the outcome step
			}
		}
		return nil
	})

	sc.Step(`^the relation warning contains "([^"]*)"$`, func(ctx context.Context, want string) error {
		if w.LastRel == nil {
			return errNoRel
		}
		if !strings.Contains(w.LastRel.Warning, want) {
			return errString("relation warning " + w.LastRel.Warning + " does not contain " + want)
		}
		common.Logf(ctx, "relation warning: %s", w.LastRel.Warning)
		return nil
	})

	sc.Step(`^I delete the last created relationship$`, func(ctx context.Context) error {
		if w.LastRelID == "" {
			return errNoRel
		}
		w.Err = w.Mut.DeleteRelationship(w.LastRelID)
		if w.Err != nil {
			return nil
		}
		common.Logf(ctx, "deleted relationship %s", w.LastRelID)
		return w.Save(ctx)
	})

	sc.Step(`^I delete the relationship "([^"]*)"$`, func(ctx context.Context, id string) error {
		w.Err = w.Mut.DeleteRelationship(id)
		if w.Err != nil {
			common.Logf(ctx, "delete relationship %q → error: %v", id, w.Err)
			return nil
		}
		common.Logf(ctx, "deleted relationship %q", id)
		return w.Save(ctx)
	})

	sc.Step(`^I delete the relationship between "([^"]*)" and "([^"]*)"$`, func(ctx context.Context, src, tgt string) error {
		n, err := w.Mut.DeleteRelationshipsBetween(src, tgt)
		w.Err = err
		if err != nil {
			common.Logf(ctx, "delete between %q and %q → error: %v", src, tgt, err)
			return nil
		}
		common.Logf(ctx, "deleted %d relationship(s) between %q and %q", n, src, tgt)
		return w.Save(ctx)
	})
}

var errNoRel = errString("no relationship was created in this scenario")

type errString string

func (e errString) Error() string { return string(e) }
