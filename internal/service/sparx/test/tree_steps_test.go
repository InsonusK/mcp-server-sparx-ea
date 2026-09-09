package sparxtest

// Actions on the model tree (method 1) and the root package name.

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func findNode(n *sparx.Node, path string) *sparx.Node {
	if n.Path == path {
		return n
	}
	for _, c := range n.Children {
		if got := findNode(c, path); got != nil {
			return got
		}
	}
	return nil
}

func registerTreeSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I read the model tree$`, func(ctx context.Context) error {
		w.Tree = w.Active().Tree()
		common.Logf(ctx, "read tree, root %q with %d top-level package(s)", w.Tree.Name, len(w.Tree.Children))
		return nil
	})

	sc.Step(`^the node at path "([^"]*)" has children:$`, func(ctx context.Context, path string, table *godog.Table) error {
		n := findNode(w.Tree, path)
		if n == nil {
			return fmt.Errorf("no node at path %q", path)
		}
		rows, err := common.ToRows(nodeSlice(n.Children))
		if err != nil {
			return err
		}
		common.Logf(ctx, "node %q has %d child(ren)", path, len(rows))
		return common.MatchTable(rows, table, true)
	})

	sc.Step(`^the node at path "([^"]*)" has id "([^"]*)"$`, func(ctx context.Context, path, want string) error {
		return nodeField(w, path, "id", want)
	})
	sc.Step(`^the node at path "([^"]*)" has guid "([^"]*)"$`, func(ctx context.Context, path, want string) error {
		return nodeField(w, path, "guid", want)
	})

	// the working-model root is renamed in the "Given the working model" step;
	// this asserts the read-only model's roots.
	sc.Step(`^the root packages are "([^"]*)"$`, func(ctx context.Context, want string) error {
		got := w.Active().RootPackages()
		if len(got) != 1 || got[0] != want {
			return fmt.Errorf("root packages = %v, want [%s]", got, want)
		}
		return nil
	})
}

func nodeField(w *common.World, path, field, want string) error {
	n := findNode(w.Tree, path)
	if n == nil {
		return fmt.Errorf("no node at path %q", path)
	}
	rows, _ := common.ToRows(nodeSlice([]*sparx.Node{n}))
	if rows[0][field] != want {
		return fmt.Errorf("node %q %s = %q, want %q", path, field, rows[0][field], want)
	}
	return nil
}

func nodeSlice(ns []*sparx.Node) []any {
	out := make([]any, len(ns))
	for i := range ns {
		out[i] = ns[i]
	}
	return out
}
