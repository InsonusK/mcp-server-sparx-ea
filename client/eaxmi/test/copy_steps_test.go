package eaxmitest

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

func registerCopySteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^a second model file "([^"]*)"$`, func(ctx context.Context, name string) error {
		d, err := eaxmi.Open(fixturePath(name))
		if err != nil {
			return fmt.Errorf("open second model %s: %w", name, err)
		}
		w.other = d
		return nil
	})

	sc.Step(`^I copy the package "([^"]*)" into "([^"]*)"$`, func(ctx context.Context, srcRef, destRef string) error {
		return copyPackage(ctx, w, w.doc, srcRef, destRef)
	})

	sc.Step(`^I copy the package "([^"]*)" from the second model into "([^"]*)"$`, func(ctx context.Context, srcRef, destRef string) error {
		return copyPackage(ctx, w, w.other, srcRef, destRef)
	})

	sc.Step(`^copying the package "([^"]*)" into "([^"]*)" fails with "([^"]*)"$`, func(ctx context.Context, srcRef, destRef, msg string) error {
		srcID, destID := "EAPK_00000000_0000_0000_0000_000000000000", "EAPK_00000000_0000_0000_0000_000000000000"
		if p := w.doc.ResolvePackage(srcRef); p != nil {
			srcID = p.XMIID
		}
		if p := w.doc.ResolvePackage(destRef); p != nil {
			destID = p.XMIID
		}
		_, err := w.doc.CopyPackage(w.doc, srcID, destID)
		return wantErr(err, msg)
	})
}

func copyPackage(ctx context.Context, w *world, from *eaxmi.Document, srcRef, destRef string) error {
	if from == nil {
		return fmt.Errorf("source model not loaded")
	}
	src := from.ResolvePackage(srcRef)
	dest := w.doc.ResolvePackage(destRef)
	if src == nil {
		return fmt.Errorf("no source package %q", srcRef)
	}
	if dest == nil {
		return fmt.Errorf("no destination package %q", destRef)
	}
	p, err := w.doc.CopyPackage(from, src.XMIID, dest.XMIID)
	w.err = err
	if err != nil {
		logf(ctx, "copy %q into %q → error: %v", srcRef, destRef, err)
		return err
	}
	w.pkg = p
	logf(ctx, "copied %q into %q → %s", srcRef, destRef, w.doc.PathOfPackage(p))
	return nil
}
