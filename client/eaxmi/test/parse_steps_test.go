package eaxmitest

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func registerParseSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the model file "([^"]*)"$`, w.open)

	sc.Step(`^the exporter version is "([^"]*)"$`, func(ctx context.Context, want string) error {
		if w.active().ExporterVersion != want {
			return fmt.Errorf("exporter version = %q, want %q", w.active().ExporterVersion, want)
		}
		return nil
	})

	sc.Step(`^the root package names are "([^"]*)"$`, func(ctx context.Context, want string) error {
		return sameSet(w.active().RootPackages(), splitList(want))
	})

	sc.Step(`^there (?:is|are) (\d+) elements?$`, func(ctx context.Context, n int) error {
		if got := len(w.active().AllElements()); got != n {
			return fmt.Errorf("%d elements, want %d", got, n)
		}
		return nil
	})
	sc.Step(`^there (?:is|are) (\d+) connectors?$`, func(ctx context.Context, n int) error {
		if got := len(w.active().Connectors()); got != n {
			return fmt.Errorf("%d connectors, want %d", got, n)
		}
		return nil
	})
	sc.Step(`^there (?:is|are) (\d+) packages?$`, func(ctx context.Context, n int) error {
		if got := len(w.active().AllPackages()); got != n {
			return fmt.Errorf("%d packages, want %d", got, n)
		}
		return nil
	})

	sc.Step(`^the element "([^"]*)" is:$`, func(ctx context.Context, ref string, table *godog.Table) error {
		e := w.active().ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("no element for %q", ref)
		}
		w.elem = e
		return checkFields(table, func(f string) (string, bool) {
			switch f {
			case "name":
				return e.Name, true
			case "id":
				return e.XMIID, true
			case "guid":
				return e.GUID, true
			case "umlType":
				return e.UMLType, true
			case "stereotype":
				return e.Stereotype, true
			case "documentation":
				return e.Documentation, true
			case "package":
				return w.active().PathOfPackage(e.Package), true
			}
			return "", false
		})
	})

	sc.Step(`^the connector from "([^"]*)" to "([^"]*)" is:$`, func(ctx context.Context, srcRef, tgtRef string, table *godog.Table) error {
		src := w.active().ResolveElement(srcRef)
		tgt := w.active().ResolveElement(tgtRef)
		if src == nil || tgt == nil {
			return fmt.Errorf("endpoint not found (%q, %q)", srcRef, tgtRef)
		}
		for _, c := range w.active().Connectors() {
			if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID {
				w.conn = c
				return checkFields(table, func(f string) (string, bool) {
					switch f {
					case "name":
						return c.Name, true
					case "eaType":
						return c.EAType, true
					case "stereotype":
						return c.Stereotype, true
					case "documentation":
						return c.Documentation, true
					}
					return "", false
				})
			}
		}
		return fmt.Errorf("no connector from %s to %s", src.Name, tgt.Name)
	})

	sc.Step(`^the diagram "([^"]*)" has (\d+) objects and (\d+) links$`, func(ctx context.Context, ref string, objs, links int) error {
		g := w.active().ResolveDiagram(ref)
		if g == nil {
			return fmt.Errorf("no diagram for %q", ref)
		}
		if len(g.Objects) != objs || len(g.Links) != links {
			return fmt.Errorf("diagram %s: %d objs / %d links, want %d / %d", g.Name, len(g.Objects), len(g.Links), objs, links)
		}
		return nil
	})

	sc.Step(`^the diagram "([^"]*)" places "([^"]*)" at (\d+),(\d+),(\d+),(\d+)$`,
		func(ctx context.Context, ref, elemRef string, l, t, r, b int) error {
			g := w.active().ResolveDiagram(ref)
			el := w.active().ResolveElement(elemRef)
			if g == nil || el == nil {
				return fmt.Errorf("diagram or element not found")
			}
			for _, o := range g.Objects {
				if o.SubjectID == el.XMIID {
					if o.Left != l || o.Top != t || o.Right != r || o.Bottom != b {
						return fmt.Errorf("%s at %d,%d,%d,%d, want %d,%d,%d,%d",
							el.Name, o.Left, o.Top, o.Right, o.Bottom, l, t, r, b)
					}
					return nil
				}
			}
			return fmt.Errorf("%s is not placed on %s", el.Name, g.Name)
		})

	sc.Step(`^every element name and documentation is valid UTF-8$`, func(ctx context.Context) error {
		for _, e := range w.active().AllElements() {
			if !isValidUTF8(e.Name) || !isValidUTF8(e.Documentation) {
				return fmt.Errorf("element %q has invalid UTF-8", e.Name)
			}
		}
		return nil
	})
}

func isValidUTF8(s string) bool { return strings.ToValidUTF8(s, "�") == s }
