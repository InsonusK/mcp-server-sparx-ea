package sparxtest

// Actions on packages: read, create, rename, move, delete (with / without
// cascade). Assertions on the result live in package common.

import (
	"context"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func registerPackageSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I read the package "([^"]*)"$`, func(ctx context.Context, ref string) error {
		w.Pkg, w.Err = w.Active().Package(ref)
		if w.Err != nil {
			common.Logf(ctx, "read package %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "read package %q → %s (%d sub, %d elem, %d diag)",
			ref, w.Pkg.Name, len(w.Pkg.Packages), len(w.Pkg.Elements), len(w.Pkg.Diagrams))
		return nil
	})

	sc.Step(`^I create a package "([^"]*)" in "([^"]*)"$`, func(ctx context.Context, name, parent string) error {
		w.LastPkgPath = parent + "/" + name
		p, err := w.Mut.CreatePackage(parent, name)
		w.Err = err
		if err != nil {
			common.Logf(ctx, "create package %q in %q → error: %v", name, parent, err)
			return nil
		}
		common.Logf(ctx, "created package %q (id %s)", name, p.ID)
		return w.Save(ctx)
	})

	sc.Step(`^I rename the package "([^"]*)" to "([^"]*)"$`, func(ctx context.Context, ref, name string) error {
		w.LastPkgPath = parentPath(ref) + "/" + name
		_, w.Err = w.Mut.RenamePackage(ref, name)
		if w.Err != nil {
			common.Logf(ctx, "rename package %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "renamed package %q → %q", ref, name)
		return w.Save(ctx)
	})

	sc.Step(`^I move the package "([^"]*)" into "([^"]*)"$`, func(ctx context.Context, ref, parent string) error {
		p, err := w.Mut.MovePackage(ref, parent)
		w.Err = err
		if err != nil {
			common.Logf(ctx, "move package %q into %q → error: %v", ref, parent, err)
			return nil
		}
		w.LastPkgPath = p.Path
		common.Logf(ctx, "moved package %q into %q → %s", ref, parent, p.Path)
		return w.Save(ctx)
	})

	sc.Step(`^I delete the package "([^"]*)"$`, func(ctx context.Context, ref string) error {
		_, w.Err = w.Mut.DeletePackage(ref, false)
		if w.Err != nil {
			common.Logf(ctx, "delete package %q → error: %v", ref, w.Err)
			return nil
		}
		common.Logf(ctx, "deleted package %q", ref)
		return w.Save(ctx)
	})

	sc.Step(`^I delete the package "([^"]*)" with its contents$`, func(ctx context.Context, ref string) error {
		n, err := w.Mut.DeletePackage(ref, true)
		w.Err = err
		if err != nil {
			common.Logf(ctx, "cascade-delete package %q → error: %v", ref, err)
			return nil
		}
		common.Logf(ctx, "cascade-deleted package %q (%d objects removed)", ref, n)
		return w.Save(ctx)
	})
}
