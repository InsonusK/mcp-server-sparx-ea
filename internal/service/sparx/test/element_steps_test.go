package sparxtest

// Actions on elements: read (method 2), create / rename / set-note / delete
// (method 4). Assertions on the result live in package common.

import (
	"context"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func parentPath(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[:i]
	}
	return ""
}

func registerElementSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I read the element "([^"]*)"$`, func(ctx context.Context, ref string) error {
		w.Element, w.Err = w.Active().Element(w.Rel(ref))
		if w.Err != nil {
			common.Logf(ctx, "read element %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "read element %q → %s (%s), %d relation(s)", ref, w.Element.Name, w.Element.Type, len(w.Element.Relations))
		return nil
	})

	sc.Step(`^I create a "([^"]*)" named "([^"]*)" in "([^"]*)" with note "([^"]*)"$`,
		func(ctx context.Context, typ, name, pkg, note string) error {
			w.LastElemPath = pkg + "/" + name
			el, err := w.Mut.CreateElement(w.Rel(pkg), typ, name, note)
			w.Err = err
			if err != nil {
				common.Logf(ctx, "create %s %q in %q → error: %v", typ, name, pkg, err)
				return nil
			}
			common.Logf(ctx, "created %s %q (id %s)", typ, name, el.ID)
			return w.Save(ctx)
		})

	sc.Step(`^I rename the element "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, name string) error {
		w.LastElemPath = parentPath(ref) + "/" + name
		_, w.Err = w.Mut.RenameElement(w.Rel(ref), name)
		if w.Err != nil {
			common.Logf(ctx, "rename %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "renamed %q → %q", ref, name)
		return w.Save(ctx)
	})

	sc.Step(`^I set the note of "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, note string) error {
		w.LastElemPath = ref
		_, w.Err = w.Mut.SetElementDocumentation(w.Rel(ref), note)
		if w.Err != nil {
			return nil
		}
		common.Logf(ctx, "set note of %q", ref)
		return w.Save(ctx)
	})

	sc.Step(`^I delete the element "([^"]*)"$`, func(ctx context.Context, ref string) error {
		w.Err = w.Mut.DeleteElement(w.Rel(ref))
		if w.Err != nil {
			common.Logf(ctx, "delete %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "deleted element %q", ref)
		return w.Save(ctx)
	})
}
