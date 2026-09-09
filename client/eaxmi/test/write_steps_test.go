package eaxmitest

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

func registerWriteSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I add a package "([^"]*)" under "([^"]*)"$`, func(ctx context.Context, name, parentRef string) error {
		parent := w.doc.ResolvePackage(parentRef)
		if parent == nil {
			return fmt.Errorf("no package for %q", parentRef)
		}
		p, err := w.doc.AddPackage(parent.XMIID, name)
		w.err = err
		if err == nil {
			w.pkg = p
		}
		return err
	})

	sc.Step(`^I add an element "([^"]*)" of type "([^"]*)" stereotype "([^"]*)" under "([^"]*)"$`,
		func(ctx context.Context, name, umlType, stereo, parentRef string) error {
			parent := w.doc.ResolvePackage(parentRef)
			if parent == nil {
				return fmt.Errorf("no package for %q", parentRef)
			}
			e, err := w.doc.AddElement(parent.XMIID, eaxmi.ElementSpec{
				Name: name, UMLType: umlType, Stereotype: stereo,
			})
			w.err = err
			if err == nil {
				w.elem = e
			}
			return err
		})

	sc.Step(`^I set the name of the element "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, name string) error {
		e := w.doc.ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("no element for %q", ref)
		}
		return w.doc.SetElementName(e.XMIID, name)
	})

	sc.Step(`^I set the documentation of the element "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, doc string) error {
		e := w.doc.ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("no element for %q", ref)
		}
		return w.doc.SetElementDocumentation(e.XMIID, doc)
	})

	sc.Step(`^I move the element "([^"]*)" under "([^"]*)"$`, func(ctx context.Context, ref, parentRef string) error {
		e := w.doc.ResolveElement(ref)
		parent := w.doc.ResolvePackage(parentRef)
		if e == nil || parent == nil {
			return fmt.Errorf("element or package not found")
		}
		return w.doc.MoveElement(e.XMIID, parent.XMIID)
	})

	sc.Step(`^I remove the element "([^"]*)"$`, func(ctx context.Context, ref string) error {
		e := w.doc.ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("no element for %q", ref)
		}
		return w.doc.RemoveElement(e.XMIID)
	})

	sc.Step(`^I remove the package "([^"]*)"$`, func(ctx context.Context, ref string) error {
		p := w.doc.ResolvePackage(ref)
		if p == nil {
			return fmt.Errorf("no package for %q", ref)
		}
		n, err := w.doc.RemovePackage(p.XMIID)
		logf(ctx, "removed %d model object(s)", n)
		return err
	})

	sc.Step(`^I add a connector from "([^"]*)" to "([^"]*)" ea-type "([^"]*)" repr "([^"]*)" stereotype "([^"]*)"$`,
		func(ctx context.Context, srcRef, tgtRef, eaType, repr, stereo string) error {
			src := w.doc.ResolveElement(srcRef)
			tgt := w.doc.ResolveElement(tgtRef)
			if src == nil || tgt == nil {
				return fmt.Errorf("endpoint not found")
			}
			c, err := w.doc.AddConnector(eaxmi.ConnectorSpec{
				SourceID: src.XMIID, TargetID: tgt.XMIID,
				EAType: eaType, ModelRepr: repr, Stereotype: stereo,
			})
			w.err = err
			if err == nil {
				w.conn = c
			}
			return err
		})

	sc.Step(`^I remove the connector from "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, srcRef, tgtRef string) error {
		src := w.doc.ResolveElement(srcRef)
		tgt := w.doc.ResolveElement(tgtRef)
		if src == nil || tgt == nil {
			return fmt.Errorf("endpoint not found")
		}
		for _, c := range w.doc.Connectors() {
			if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID {
				return w.doc.RemoveConnector(c.XMIID)
			}
		}
		return fmt.Errorf("no connector from %s to %s", src.Name, tgt.Name)
	})

	sc.Step(`^I place the element "([^"]*)" on the diagram "([^"]*)" at (\d+),(\d+),(\d+),(\d+)$`,
		func(ctx context.Context, elemRef, diagRef string, l, t, r, b int) error {
			e := w.doc.ResolveElement(elemRef)
			g := w.doc.ResolveDiagram(diagRef)
			if e == nil || g == nil {
				return fmt.Errorf("element or diagram not found")
			}
			return w.doc.AddDiagramObject(g.XMIID, e.XMIID, l, t, r, b)
		})

	sc.Step(`^I take the element "([^"]*)" off the diagram "([^"]*)"$`, func(ctx context.Context, elemRef, diagRef string) error {
		e := w.doc.ResolveElement(elemRef)
		g := w.doc.ResolveDiagram(diagRef)
		if e == nil || g == nil {
			return fmt.Errorf("element or diagram not found")
		}
		return w.doc.RemoveDiagramObject(g.XMIID, e.XMIID)
	})

	sc.Step(`^I write and reopen as "([^"]*)"$`, func(ctx context.Context, name string) error {
		return w.saveAndReopen(ctx, name)
	})

	sc.Step(`^the operation fails with "([^"]*)"$`, func(ctx context.Context, msg string) error {
		return wantErr(w.err, msg)
	})

	// assertions that run against the reopened file and read raw XML
	sc.Step(`^the reopened file has no dangling id references$`, func(ctx context.Context) error {
		return noDanglingRefs(ctx, w)
	})

	sc.Step(`^the written file contains "([^"]*)"$`, func(ctx context.Context, want string) error {
		b, err := os.ReadFile(w.savedPath)
		if err != nil {
			return err
		}
		if !strings.Contains(string(b), want) {
			return fmt.Errorf("written file %s does not contain %q", w.savedPath, want)
		}
		return nil
	})
}
