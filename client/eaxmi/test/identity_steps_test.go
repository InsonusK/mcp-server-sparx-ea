package eaxmitest

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

func registerIdentitySteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I record every element GUID by path$`, func(ctx context.Context) error {
		w.guidBefore = map[string]string{}
		for _, e := range w.doc.AllElements() {
			w.guidBefore[w.doc.PathOfElement(e)] = e.GUID
		}
		for _, p := range w.doc.AllPackages() {
			w.guidBefore["pkg:"+w.doc.PathOfPackage(p)] = p.GUID
		}
		logf(ctx, "recorded %d GUIDs", len(w.guidBefore))
		return nil
	})

	sc.Step(`^I remap the model identity$`, func(ctx context.Context) error {
		return w.doc.RemapIdentity()
	})

	sc.Step(`^every recorded GUID has changed$`, func(ctx context.Context) error {
		if len(w.guidBefore) == 0 {
			return fmt.Errorf("no GUIDs were recorded")
		}
		for path, old := range w.guidBefore {
			var now string
			if p, ok := trimPkg(path); ok {
				pk := w.doc.ResolvePackage(p)
				if pk == nil {
					return fmt.Errorf("package %q vanished", p)
				}
				now = pk.GUID
			} else {
				e := w.doc.ResolveElement(path)
				if e == nil {
					return fmt.Errorf("element %q vanished", path)
				}
				now = e.GUID
			}
			if now == old {
				return fmt.Errorf("%s kept its GUID %s", path, old)
			}
		}
		return nil
	})

	sc.Step(`^the connector from "([^"]*)" to "([^"]*)" still links those two elements$`,
		func(ctx context.Context, srcRef, tgtRef string) error {
			src := w.doc.ResolveElement(srcRef)
			tgt := w.doc.ResolveElement(tgtRef)
			if src == nil || tgt == nil {
				return fmt.Errorf("endpoint not found after remap")
			}
			for _, c := range w.doc.Connectors() {
				if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID {
					return nil
				}
			}
			return fmt.Errorf("connector %s -> %s lost after remap", src.Name, tgt.Name)
		})
}

func trimPkg(s string) (string, bool) {
	const p = "pkg:"
	if len(s) > len(p) && s[:len(p)] == p {
		return s[len(p):], true
	}
	return "", false
}
