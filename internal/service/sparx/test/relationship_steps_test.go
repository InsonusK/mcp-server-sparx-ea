package sparxtest

// Actions on relationships: create / delete (method 5).

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func registerRelationshipSteps(sc *godog.ScenarioContext, w *common.World) {
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
}

var errNoRel = errString("no relationship was created in this scenario")

type errString string

func (e errString) Error() string { return string(e) }
