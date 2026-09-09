package sparxtest

// Steps for features/diagram_contents.feature.

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func (w *world) iReadTheDiagram(ctx context.Context, ref string) error {
	if w.svc == nil {
		return fmt.Errorf("no model loaded")
	}
	w.diagram, w.lastErr = w.svc.Diagram(ref)
	if w.lastErr != nil {
		logf(ctx, "read diagram %q → error: %v", ref, w.lastErr)
		return nil
	}
	logf(ctx, "read diagram %q → %s (%s), %d object(s), %d link(s)",
		ref, w.diagram.Name, w.diagram.DiagramType, len(w.diagram.Objects), len(w.diagram.Links))
	return nil
}

func (w *world) diagramFieldIs(ctx context.Context, field, want string) error {
	if w.diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.lastErr)
	}
	rows, _ := toRows([]*sparx.DiagramInfo{w.diagram})
	if got := rows[0][field]; got != want {
		return fmt.Errorf("diagram %s = %q, want %q", field, got, want)
	}
	logf(ctx, "diagram %s == %q", field, want)
	return nil
}

func (w *world) diagramHasNPlacedElements(ctx context.Context, n int) error {
	if w.diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.lastErr)
	}
	if got := len(w.diagram.Objects); got != n {
		return fmt.Errorf("diagram has %d placed elements, want %d", got, n)
	}
	logf(ctx, "diagram has %d placed element(s)", n)
	return nil
}

func (w *world) diagramHasNLinks(ctx context.Context, n int) error {
	if w.diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.lastErr)
	}
	if got := len(w.diagram.Links); got != n {
		return fmt.Errorf("diagram has %d links, want %d", got, n)
	}
	logf(ctx, "diagram has %d link(s)", n)
	return nil
}

func (w *world) placedElementsInclude(ctx context.Context, table *godog.Table) error {
	if w.diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.lastErr)
	}
	rows, err := toRows(w.diagram.Objects)
	if err != nil {
		return err
	}
	return matchTable(rows, table, false)
}

func registerDiagramSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I read the diagram "([^"]*)"$`, w.iReadTheDiagram)
	sc.Step(`^the diagram field "([^"]*)" is "([^"]*)"$`, w.diagramFieldIs)
	sc.Step(`^the diagram has (\d+) placed elements$`, w.diagramHasNPlacedElements)
	sc.Step(`^the diagram has (\d+) links$`, w.diagramHasNLinks)
	sc.Step(`^the placed elements include:$`, w.placedElementsInclude)
}
