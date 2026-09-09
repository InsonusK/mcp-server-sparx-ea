package sparxtest

// Steps for features/element_info.feature.

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func (w *world) iReadTheElement(ctx context.Context, ref string) error {
	if w.svc == nil {
		return fmt.Errorf("no model loaded")
	}
	w.element, w.lastErr = w.svc.Element(ref)
	if w.lastErr != nil {
		logf(ctx, "read element %q → error: %v", ref, w.lastErr)
		return nil
	}
	logf(ctx, "read element %q → %s (%s), %d relation(s)", ref, w.element.Name, w.element.Type, len(w.element.Relations))
	return nil
}

func (w *world) readingFailsWith(ctx context.Context, want string) error {
	if w.lastErr == nil {
		return fmt.Errorf("expected an error containing %q, but the read succeeded", want)
	}
	if !strings.Contains(w.lastErr.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.lastErr.Error(), want)
	}
	logf(ctx, "failed as expected: %v", w.lastErr)
	return nil
}

func (w *world) elementIs(ctx context.Context, table *godog.Table) error {
	if w.element == nil {
		return fmt.Errorf("no element read (err: %v)", w.lastErr)
	}
	rows, err := toRows([]*sparx.ElementInfo{w.element})
	if err != nil {
		return err
	}
	actual := rows[0]
	for _, r := range table.Rows[1:] { // skip "| field | value |" header
		field, want := r.Cells[0].Value, r.Cells[1].Value
		if actual[field] != want {
			return fmt.Errorf("element %s = %q, want %q", field, actual[field], want)
		}
	}
	logf(ctx, "element matches all %d field(s)", len(table.Rows)-1)
	return nil
}

func (w *world) elementFieldIs(ctx context.Context, field, want string) error {
	if w.element == nil {
		return fmt.Errorf("no element read (err: %v)", w.lastErr)
	}
	rows, _ := toRows([]*sparx.ElementInfo{w.element})
	if got := rows[0][field]; got != want {
		return fmt.Errorf("element %s = %q, want %q", field, got, want)
	}
	logf(ctx, "element %s == %q", field, want)
	return nil
}

func (w *world) elementRelations(ctx context.Context, exact bool, table *godog.Table) error {
	if w.element == nil {
		return fmt.Errorf("no element read (err: %v)", w.lastErr)
	}
	rows, err := toRows(w.element.Relations)
	if err != nil {
		return err
	}
	logf(ctx, "element has %d relation(s)", len(rows))
	return matchTable(rows, table, exact)
}

func registerElementSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I read the element "([^"]*)"$`, w.iReadTheElement)
	sc.Step(`^reading fails with "([^"]*)"$`, w.readingFailsWith)
	sc.Step(`^the element is:$`, w.elementIs)
	sc.Step(`^the element field "([^"]*)" is "([^"]*)"$`, w.elementFieldIs)
	sc.Step(`^the element relations are exactly:$`, func(ctx context.Context, t *godog.Table) error {
		return w.elementRelations(ctx, true, t)
	})
	sc.Step(`^the element relations include:$`, func(ctx context.Context, t *godog.Table) error {
		return w.elementRelations(ctx, false, t)
	})
}
