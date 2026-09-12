package mcpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/mark3labs/mcp-go/mcp"
)

func registerSparxSteps(sc *godog.ScenarioContext, w *toolWorld) {
	sc.Step(`^a running MCP server with a fake model$`, w.aRunningMCPServerWithAFakeModel)

	sc.Step(`^the model returns the error "([^"]*)"$`, func(msg string) error {
		w.fake.err = errors.New(msg)
		return nil
	})
	sc.Step(`^saving the model fails with "([^"]*)"$`, func(msg string) error {
		w.fake.saveErr = errors.New(msg)
		return nil
	})

	sc.Step(`^I call "([^"]*)" with:$`, func(tool string, table *godog.Table) error {
		args := map[string]any{}
		for _, row := range table.Rows {
			if len(row.Cells) != 2 {
				continue
			}
			args[row.Cells[0].Value] = typed(row.Cells[1].Value)
		}
		req := mcp.CallToolRequest{}
		req.Params.Name = tool
		req.Params.Arguments = args
		w.toolRes, w.toolErr = w.client.CallTool(context.Background(), req)
		return nil
	})

	sc.Step(`^the model was opened with "([^"]*)"$`, func(path string) error {
		if w.fake.openedWith != path {
			return fmt.Errorf("model opened with %q, want %q", w.fake.openedWith, path)
		}
		return nil
	})

	sc.Step(`^the model method "([^"]*)" was called$`, func(sub string) error {
		if !w.fake.called(sub) {
			return fmt.Errorf("no recorded call matches %q; calls: %v", sub, w.fake.calls)
		}
		return nil
	})

	sc.Step(`^no model method was called$`, func() error {
		if len(w.fake.calls) != 0 {
			return fmt.Errorf("expected no calls, got %v", w.fake.calls)
		}
		return nil
	})

	sc.Step(`^the model method "([^"]*)" was not called$`, func(sub string) error {
		if w.fake.called(sub) {
			return fmt.Errorf("did not expect a call matching %q; calls: %v", sub, w.fake.calls)
		}
		return nil
	})

	sc.Step(`^the model was saved to "([^"]*)"$`, func(path string) error {
		if w.fake.savedTo != path {
			return fmt.Errorf("model saved to %q, want %q", w.fake.savedTo, path)
		}
		return nil
	})

	sc.Step(`^the model was not saved$`, func() error {
		if w.fake.savedTo != "" {
			return fmt.Errorf("model was saved to %q", w.fake.savedTo)
		}
		return nil
	})
}

// typed turns a table cell into an int, a bool, a []any (a JSON array, for a
// refs-style list arg) or a string, so tool args reach the handler with the
// right JSON type.
func typed(s string) any {
	if strings.HasPrefix(s, "[") {
		var v []any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			return v
		}
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	if b, err := strconv.ParseBool(s); err == nil && (s == "true" || s == "false") {
		return b
	}
	return s
}
