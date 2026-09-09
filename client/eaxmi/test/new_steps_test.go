package eaxmitest

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

func registerNewModelSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^a new model with root "([^"]*)"$`, func(ctx context.Context, root string) error {
		d, err := eaxmi.NewModel(root)
		if err != nil {
			return err
		}
		w.doc = d
		logf(ctx, "new model, root %q", root)
		return nil
	})

	sc.Step(`^creating a new model with root "([^"]*)" fails with "([^"]*)"$`, func(ctx context.Context, root, msg string) error {
		_, err := eaxmi.NewModel(root)
		return wantErr(err, msg)
	})

	sc.Step(`^I add a root package "([^"]*)"$`, func(ctx context.Context, name string) error {
		p, err := w.doc.AddRootPackage(name)
		w.err = err
		if err != nil {
			logf(ctx, "add root package %q → error: %v", name, err)
			return nil
		}
		w.pkg = p
		return nil
	})

	sc.Step(`^adding a root package "([^"]*)" fails with "([^"]*)"$`, func(ctx context.Context, name, msg string) error {
		_, err := w.doc.AddRootPackage(name)
		return wantErr(err, msg)
	})
}

func wantErr(err error, msg string) error {
	if err == nil {
		return fmt.Errorf("expected an error containing %q, got nil", msg)
	}
	if !strings.Contains(err.Error(), msg) {
		return fmt.Errorf("error %q does not contain %q", err.Error(), msg)
	}
	return nil
}
