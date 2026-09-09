package sparxtest

// Steps for features/model_tree.feature + the shared "given the model file" step.

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func (w *world) givenTheModelFile(ctx context.Context, name string) error {
	svc, err := sparx.Open(fixturePath(name))
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	w.svc = svc
	logf(ctx, "loaded model %q", name)
	return nil
}

func (w *world) iReadTheModelTree(ctx context.Context) error {
	if w.svc == nil {
		return fmt.Errorf("no model loaded")
	}
	w.tree = w.svc.Tree()
	logf(ctx, "read tree, root %q with %d top-level package(s)", w.tree.Name, len(w.tree.Children))
	return nil
}

func (w *world) findNode(path string) *sparx.Node {
	var rec func(n *sparx.Node) *sparx.Node
	rec = func(n *sparx.Node) *sparx.Node {
		if n.Path == path {
			return n
		}
		for _, c := range n.Children {
			if got := rec(c); got != nil {
				return got
			}
		}
		return nil
	}
	return rec(w.tree)
}

func (w *world) nodeHasChildren(ctx context.Context, path string, table *godog.Table) error {
	n := w.findNode(path)
	if n == nil {
		return fmt.Errorf("no node at path %q", path)
	}
	rows, err := toRows(n.Children)
	if err != nil {
		return err
	}
	logf(ctx, "node %q has %d child(ren)", path, len(rows))
	return matchTable(rows, table, true)
}

func (w *world) nodeHasField(ctx context.Context, path, field, want string) error {
	n := w.findNode(path)
	if n == nil {
		return fmt.Errorf("no node at path %q", path)
	}
	rows, err := toRows([]*sparx.Node{n})
	if err != nil {
		return err
	}
	if got := rows[0][field]; got != want {
		return fmt.Errorf("node %q %s = %q, want %q", path, field, got, want)
	}
	logf(ctx, "node %q %s == %q", path, field, want)
	return nil
}

func registerModelSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the model file "([^"]*)"$`, w.givenTheModelFile)
	sc.Step(`^I read the model tree$`, w.iReadTheModelTree)
	sc.Step(`^the node at path "([^"]*)" has children:$`, w.nodeHasChildren)
	sc.Step(`^the node at path "([^"]*)" has id "([^"]*)"$`, func(ctx context.Context, p, v string) error {
		return w.nodeHasField(ctx, p, "id", v)
	})
	sc.Step(`^the node at path "([^"]*)" has guid "([^"]*)"$`, func(ctx context.Context, p, v string) error {
		return w.nodeHasField(ctx, p, "guid", v)
	})
}
