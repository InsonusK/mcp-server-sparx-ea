package eaxmitest

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

func registerNavigateSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^"([^"]*)" (is|is not) a ref id$`, func(ctx context.Context, s, verb string) error {
		got := eaxmi.IsRefID(s)
		want := verb == "is"
		if got != want {
			return fmt.Errorf("IsRefID(%q) = %v, want %v", s, got, want)
		}
		return nil
	})

	sc.Step(`^splitting the path "([^"]*)" gives "([^"]*)"$`, func(ctx context.Context, path, want string) error {
		got := eaxmi.SplitPath(path)
		return sameSet(got, splitList(want))
	})

	sc.Step(`^resolving the element "([^"]*)" finds "([^"]*)"$`, func(ctx context.Context, ref, wantName string) error {
		e := w.active().ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("resolved %q to nil, want %q", ref, wantName)
		}
		if e.Name != wantName {
			return fmt.Errorf("resolved %q to %q, want %q", ref, e.Name, wantName)
		}
		w.elem = e
		return nil
	})

	sc.Step(`^resolving the element "([^"]*)" finds nothing$`, func(ctx context.Context, ref string) error {
		if e := w.active().ResolveElement(ref); e != nil {
			return fmt.Errorf("resolved %q to %q, want nothing", ref, e.Name)
		}
		return nil
	})

	sc.Step(`^resolving the diagram "([^"]*)" finds "([^"]*)"$`, func(ctx context.Context, ref, wantName string) error {
		g := w.active().ResolveDiagram(ref)
		if g == nil || g.Name != wantName {
			return fmt.Errorf("resolveDiagram(%q) = %v, want %q", ref, g, wantName)
		}
		return nil
	})

	sc.Step(`^resolving the package "([^"]*)" finds "([^"]*)"$`, func(ctx context.Context, ref, wantName string) error {
		p := w.active().ResolvePackage(ref)
		if p == nil || p.Name != wantName {
			return fmt.Errorf("resolvePackage(%q) = %v, want %q", ref, p, wantName)
		}
		return nil
	})

	sc.Step(`^the same element resolves by id and by GUID$`, func(ctx context.Context) error {
		if w.elem == nil {
			return fmt.Errorf("no element resolved yet")
		}
		byID := w.active().ResolveElement(w.elem.XMIID)
		byGUID := w.active().ResolveElement(w.elem.GUID)
		if byID == nil || byGUID == nil || byID.XMIID != w.elem.XMIID || byGUID.XMIID != w.elem.XMIID {
			return fmt.Errorf("id/GUID lookup disagree for %s", w.elem.Name)
		}
		return nil
	})

	sc.Step(`^the path of the element "([^"]*)" is "([^"]*)"$`, func(ctx context.Context, ref, want string) error {
		e := w.active().ResolveElement(ref)
		if e == nil {
			return fmt.Errorf("no element for %q", ref)
		}
		if got := w.active().PathOfElement(e); got != want {
			return fmt.Errorf("path = %q, want %q", got, want)
		}
		return nil
	})

	sc.Step(`^the element names are "([^"]*)"$`, func(ctx context.Context, want string) error {
		return sameSet(names(w.active().AllElements()), splitList(want))
	})

	sc.Step(`^the package paths are "([^"]*)"$`, func(ctx context.Context, want string) error {
		var got []string
		for _, p := range w.active().AllPackages() {
			got = append(got, w.active().PathOfPackage(p))
		}
		return sameSet(got, splitList(want))
	})
}
