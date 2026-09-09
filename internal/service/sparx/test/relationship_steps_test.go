package sparxtest

// Actions on relationships: create / delete (method 5).

import (
	"context"

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
		w.LastRelID = r.ID
		common.Logf(ctx, "created %s %q -> %q (id %s)", rel, src, tgt, r.ID)
		return w.Save(ctx)
	}

	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)"$`, func(ctx context.Context, s, t, r string) error {
		return relate(ctx, s, t, r, "")
	})
	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)" named "([^"]*)"$`, relate)

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
