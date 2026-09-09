package sparxtest

// Steps for methods 4-6 (create_element / create_relationship / diagram_placement).
//
// Each mutating scenario loads a fresh in-memory model from testdata/, applies
// its edits (the testdata file is never written), and Saves the result to
// tmp/<scenario>.xml — a gitignored directory kept after the run so it can be
// imported into Sparx EA for manual review. "after reload" steps reopen that
// saved file, so assertions are made against a real round-trip.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

const tmpDir = "tmp"

func (w *world) givenAWorkingCopy(ctx context.Context, name string) error {
	svc, err := sparx.Open(fixturePath(name))
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	w.mutSvc = svc
	w.mutFixture = name
	w.savedPath = ""
	w.lastRelID = ""
	logf(ctx, "working copy of %q loaded (edits go to a copy, not the fixture)", name)
	return nil
}

// save persists the current working copy under tmp/<scenario>.xml and returns
// its path.
func (w *world) save(ctx context.Context) error {
	if w.mutSvc == nil {
		return fmt.Errorf("no working copy")
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	w.savedPath = filepath.Join(tmpDir, slug(w.scenarioName)+".xml")
	if err := w.mutSvc.Save(w.savedPath); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	logf(ctx, "saved working copy → %s (open in Sparx to review / import)", w.savedPath)
	return nil
}

func (w *world) reloaded(ctx context.Context) (*sparx.Service, error) {
	if w.savedPath == "" {
		if err := w.save(ctx); err != nil {
			return nil, err
		}
	}
	return sparx.Open(w.savedPath)
}

// ---------- method 4: elements ----------

func (w *world) iCreateElement(ctx context.Context, typ, name, pkg, note string) error {
	el, err := w.mutSvc.CreateElement(pkg, typ, name, note)
	w.lastErr = err
	if err != nil {
		logf(ctx, "create %s %q in %q → error: %v", typ, name, pkg, err)
		return nil
	}
	logf(ctx, "created %s %q (id %s)", typ, name, el.ID)
	return w.save(ctx)
}

func (w *world) iRename(ctx context.Context, ref, newName string) error {
	_, w.lastErr = w.mutSvc.RenameElement(ref, newName)
	if w.lastErr == nil {
		logf(ctx, "renamed %q → %q", ref, newName)
		return w.save(ctx)
	}
	logf(ctx, "rename %q → error: %v", ref, w.lastErr)
	return nil
}

func (w *world) iSetNote(ctx context.Context, ref, note string) error {
	_, w.lastErr = w.mutSvc.SetElementDocumentation(ref, note)
	if w.lastErr == nil {
		logf(ctx, "set note of %q", ref)
		return w.save(ctx)
	}
	return nil
}

func (w *world) iDeleteElement(ctx context.Context, ref string) error {
	w.lastErr = w.mutSvc.DeleteElement(ref)
	if w.lastErr == nil {
		logf(ctx, "deleted element %q", ref)
		return w.save(ctx)
	}
	logf(ctx, "delete %q → error: %v", ref, w.lastErr)
	return nil
}

// ---------- method 5: relationships ----------

func (w *world) iRelate(ctx context.Context, src, tgt, rel, name string) error {
	r, err := w.mutSvc.CreateRelationship(src, tgt, rel, name, "")
	w.lastErr = err
	if err != nil {
		logf(ctx, "relate %q -> %q as %s → error: %v", src, tgt, rel, err)
		return nil
	}
	w.lastRelID = r.ID
	logf(ctx, "created %s %q -> %q (id %s)", rel, src, tgt, r.ID)
	return w.save(ctx)
}

func (w *world) iDeleteLastRelationship(ctx context.Context) error {
	if w.lastRelID == "" {
		return fmt.Errorf("no relationship was created in this scenario")
	}
	w.lastErr = w.mutSvc.DeleteRelationship(w.lastRelID)
	if w.lastErr == nil {
		logf(ctx, "deleted relationship %s", w.lastRelID)
		return w.save(ctx)
	}
	return nil
}

// ---------- method 6: diagram placement ----------

func (w *world) iAddToDiagram(ctx context.Context, elemRef, diagRef string, rect string) error {
	r, err := parseRect(rect)
	if err != nil {
		return err
	}
	w.lastErr = w.mutSvc.AddToDiagram(diagRef, elemRef, r)
	if w.lastErr == nil {
		logf(ctx, "added %q to diagram %q at %s", elemRef, diagRef, rect)
		return w.save(ctx)
	}
	logf(ctx, "add %q to %q → error: %v", elemRef, diagRef, w.lastErr)
	return nil
}

func (w *world) iMoveOnDiagram(ctx context.Context, elemRef, diagRef, rect string) error {
	r, err := parseRect(rect)
	if err != nil {
		return err
	}
	w.lastErr = w.mutSvc.MoveOnDiagram(diagRef, elemRef, r)
	if w.lastErr == nil {
		logf(ctx, "moved %q on %q to %s", elemRef, diagRef, rect)
		return w.save(ctx)
	}
	return nil
}

func (w *world) iRemoveFromDiagram(ctx context.Context, elemRef, diagRef string) error {
	w.lastErr = w.mutSvc.RemoveFromDiagram(diagRef, elemRef)
	if w.lastErr == nil {
		logf(ctx, "removed %q from diagram %q", elemRef, diagRef)
		return w.save(ctx)
	}
	return nil
}

// ---------- outcome assertions ----------

func (w *world) opSucceeds(ctx context.Context) error {
	if w.lastErr != nil {
		return fmt.Errorf("expected success, got: %v", w.lastErr)
	}
	logf(ctx, "operation succeeded")
	return nil
}

func (w *world) opFailsWith(ctx context.Context, want string) error {
	if w.lastErr == nil {
		return fmt.Errorf("expected failure containing %q, but it succeeded", want)
	}
	if !strings.Contains(w.lastErr.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.lastErr.Error(), want)
	}
	logf(ctx, "failed as expected: %v", w.lastErr)
	return nil
}

func (w *world) opOutcome(ctx context.Context, outcome string) error {
	if outcome == "succeeds" {
		return w.opSucceeds(ctx)
	}
	if strings.HasPrefix(outcome, "fails with ") {
		return w.opFailsWith(ctx, strings.Trim(strings.TrimPrefix(outcome, "fails with "), `"`))
	}
	return fmt.Errorf("unknown outcome %q", outcome)
}

// ---------- "after reload ..." assertions ----------

func (w *world) reloadElementIs(ctx context.Context, ref string, table *godog.Table) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(ref)
	if err != nil {
		return err
	}
	rows, _ := toRows([]*sparx.ElementInfo{el})
	for _, r := range table.Rows[1:] {
		f, v := r.Cells[0].Value, r.Cells[1].Value
		if rows[0][f] != v {
			return fmt.Errorf("after reload %s = %q, want %q", f, rows[0][f], v)
		}
	}
	logf(ctx, "reloaded element %q matches", ref)
	return nil
}

func (w *world) reloadElementFieldIs(ctx context.Context, field, want string) error {
	// operates on the element created/renamed last — the feature always pairs
	// this with a preceding create/rename, so use w.savedPath's newest name.
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.lastElemPath)
	if err != nil {
		return err
	}
	rows, _ := toRows([]*sparx.ElementInfo{el})
	if rows[0][field] != want {
		return fmt.Errorf("after reload %s = %q, want %q", field, rows[0][field], want)
	}
	return nil
}

func (w *world) reloadElementNotFound(ctx context.Context, ref string) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Element(ref); err == nil {
		return fmt.Errorf("element %q still resolves after reload", ref)
	}
	logf(ctx, "element %q is gone after reload", ref)
	return nil
}

func (w *world) reloadElementNoRelationTo(ctx context.Context, ref, otherName string) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(ref)
	if err != nil {
		return err
	}
	for _, r := range el.Relations {
		if r.OtherName == otherName {
			return fmt.Errorf("%s still has a %s relation to %s after reload", el.Name, r.Type, otherName)
		}
	}
	logf(ctx, "%q has no relation to %q after reload", ref, otherName)
	return nil
}

func (w *world) reloadElementRelationsInclude(ctx context.Context, ref string, table *godog.Table) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(ref)
	if err != nil {
		return err
	}
	rows, err := toRows(el.Relations)
	if err != nil {
		return err
	}
	return matchTable(rows, table, false)
}

func (w *world) reloadDiagramHasNPlaced(ctx context.Context, ref string, n int) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	d, err := svc.Diagram(ref)
	if err != nil {
		return err
	}
	if len(d.Objects) != n {
		return fmt.Errorf("after reload diagram has %d placed elements, want %d", len(d.Objects), n)
	}
	logf(ctx, "reloaded diagram %q has %d placed element(s)", ref, n)
	return nil
}

func (w *world) reloadDiagramPlacedInclude(ctx context.Context, table *godog.Table) error {
	svc, err := w.reloaded(ctx)
	if err != nil {
		return err
	}
	d, err := svc.Diagram(w.lastDiagramRef)
	if err != nil {
		return err
	}
	rows, err := toRows(d.Objects)
	if err != nil {
		return err
	}
	return matchTable(rows, table, false)
}

// ---------- helpers ----------

var rectRe = regexp.MustCompile(`^\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*$`)

func parseRect(s string) (sparx.Rect, error) {
	m := rectRe.FindStringSubmatch(s)
	if m == nil {
		return sparx.Rect{}, fmt.Errorf("bad rectangle %q, want left,top,right,bottom", s)
	}
	n := func(i int) int { v, _ := strconv.Atoi(m[i]); return v }
	return sparx.Rect{Left: n(1), Top: n(2), Right: n(3), Bottom: n(4)}, nil
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func slug(s string) string {
	return strings.Trim(strings.ToLower(slugRe.ReplaceAllString(s, "_")), "_")
}

func registerMutateSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^a working copy of "([^"]*)"$`, w.givenAWorkingCopy)

	sc.Step(`^I create a "([^"]*)" named "([^"]*)" in "([^"]*)" with note "([^"]*)"$`,
		func(ctx context.Context, typ, name, pkg, note string) error {
			w.lastElemPath = pkg + "/" + name
			return w.iCreateElement(ctx, typ, name, pkg, note)
		})
	sc.Step(`^I rename "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, n string) error {
		w.lastElemPath = parentPath(ref) + "/" + n
		return w.iRename(ctx, ref, n)
	})
	sc.Step(`^I set the note of "([^"]*)" to "([^"]*)"$`, w.iSetNote)
	sc.Step(`^I delete the element "([^"]*)"$`, w.iDeleteElement)

	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)"$`, func(ctx context.Context, s, t, r string) error {
		return w.iRelate(ctx, s, t, r, "")
	})
	sc.Step(`^I relate "([^"]*)" to "([^"]*)" as "([^"]*)" named "([^"]*)"$`, w.iRelate)
	sc.Step(`^I delete the last created relationship$`, w.iDeleteLastRelationship)

	sc.Step(`^I add "([^"]*)" to the diagram "([^"]*)" at (.+)$`, func(ctx context.Context, e, d, r string) error {
		w.lastDiagramRef = d
		return w.iAddToDiagram(ctx, e, d, r)
	})
	sc.Step(`^I move "([^"]*)" on the diagram "([^"]*)" to (.+)$`, func(ctx context.Context, e, d, r string) error {
		w.lastDiagramRef = d
		return w.iMoveOnDiagram(ctx, e, d, r)
	})
	sc.Step(`^I remove "([^"]*)" from the diagram "([^"]*)"$`, func(ctx context.Context, e, d string) error {
		w.lastDiagramRef = d
		return w.iRemoveFromDiagram(ctx, e, d)
	})

	sc.Step(`^the (?:create|delete|placement) succeeds$`, w.opSucceeds)
	sc.Step(`^the (?:create|placement) fails with "([^"]*)"$`, w.opFailsWith)
	sc.Step(`^the relate (succeeds|fails with .+)$`, w.opOutcome)

	sc.Step(`^after reload the element "([^"]*)" is:$`, w.reloadElementIs)
	sc.Step(`^after reload the element field "([^"]*)" is "([^"]*)"$`, w.reloadElementFieldIs)
	sc.Step(`^after reload the element "([^"]*)" cannot be found$`, w.reloadElementNotFound)
	sc.Step(`^after reload the element "([^"]*)" has no relation to "([^"]*)"$`, w.reloadElementNoRelationTo)
	sc.Step(`^after reload the element "([^"]*)" relations include:$`, w.reloadElementRelationsInclude)
	sc.Step(`^after reload the diagram "([^"]*)" has (\d+) placed elements$`, w.reloadDiagramHasNPlaced)
	sc.Step(`^after reload the diagram placed elements include:$`, w.reloadDiagramPlacedInclude)
}

func parentPath(ref string) string {
	i := strings.LastIndex(ref, "/")
	if i < 0 {
		return ""
	}
	return ref[:i]
}
