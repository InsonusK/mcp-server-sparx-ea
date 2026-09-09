package sparxtest

// Actions on diagrams: read (method 3), add / move / remove an element
// (method 6).

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

var rectRe = regexp.MustCompile(`^\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*$`)

func parseRect(s string) (sparx.Rect, error) {
	m := rectRe.FindStringSubmatch(s)
	if m == nil {
		return sparx.Rect{}, fmt.Errorf("bad rectangle %q, want left,top,right,bottom", s)
	}
	n := func(i int) int { v, _ := strconv.Atoi(m[i]); return v }
	return sparx.Rect{Left: n(1), Top: n(2), Right: n(3), Bottom: n(4)}, nil
}

func registerDiagramSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I read the diagram "([^"]*)"$`, func(ctx context.Context, ref string) error {
		w.Diagram, w.Err = w.Active().Diagram(ref)
		if w.Err != nil {
			common.Logf(ctx, "read diagram %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "read diagram %q → %s (%s), %d object(s), %d link(s)",
			ref, w.Diagram.Name, w.Diagram.DiagramType, len(w.Diagram.Objects), len(w.Diagram.Links))
		return nil
	})

	sc.Step(`^I create a diagram "([^"]*)" with layer "([^"]*)" in "([^"]*)"$`, func(ctx context.Context, name, layer, parent string) error {
		w.LastDiagramRef = parent + "/" + name
		g, err := w.Mut.CreateDiagram(parent, name, layer)
		w.Err = err
		if err != nil {
			common.Logf(ctx, "create diagram %q (layer %q) in %q → error: %v", name, layer, parent, err)
			return nil
		}
		w.Diagram = g
		common.Logf(ctx, "created diagram %q (id %s, type %s) in %q", name, g.ID, g.DiagramType, parent)
		return w.Save(ctx)
	})

	sc.Step(`^I add "([^"]*)" to the diagram "([^"]*)" at (.+)$`, func(ctx context.Context, e, d, r string) error {
		w.LastDiagramRef = d
		rect, err := parseRect(r)
		if err != nil {
			return err
		}
		w.Err = w.Mut.AddToDiagram(d, e, rect)
		if w.Err != nil {
			common.Logf(ctx, "add %q to %q → error: %v", e, d, w.Err)
			return nil
		}
		common.Logf(ctx, "added %q to diagram %q at %s", e, d, r)
		return w.Save(ctx)
	})

	sc.Step(`^I move "([^"]*)" on the diagram "([^"]*)" to (.+)$`, func(ctx context.Context, e, d, r string) error {
		w.LastDiagramRef = d
		rect, err := parseRect(r)
		if err != nil {
			return err
		}
		w.Err = w.Mut.MoveOnDiagram(d, e, rect)
		if w.Err != nil {
			return nil
		}
		common.Logf(ctx, "moved %q on %q to %s", e, d, r)
		return w.Save(ctx)
	})

	sc.Step(`^I remove "([^"]*)" from the diagram "([^"]*)"$`, func(ctx context.Context, e, d string) error {
		w.LastDiagramRef = d
		w.Err = w.Mut.RemoveFromDiagram(d, e)
		if w.Err != nil {
			return nil
		}
		common.Logf(ctx, "removed %q from diagram %q", e, d)
		return w.Save(ctx)
	})
}
