package mcpserver_test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/bddsupport"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/mcpserver"
)

type toolWorld struct {
	client  *client.Client
	tools   []mcp.Tool
	toolRes *mcp.CallToolResult
	toolErr error
}

func (w *toolWorld) reset() {
	if w.client != nil {
		_ = w.client.Close()
	}
	*w = toolWorld{}
}

func (w *toolWorld) aRunningMCPServer() error {
	c, err := client.NewInProcessClient(mcpserver.New(nil))
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		return err
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "godog", Version: "1.0.0"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		return err
	}
	w.client = c
	return nil
}

func (w *toolWorld) iListTheMCPTools() error {
	res, err := w.client.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return err
	}
	w.tools = res.Tools
	return nil
}

func (w *toolWorld) theToolIsAvailable(name string) error {
	for _, t := range w.tools {
		if t.Name == name {
			return nil
		}
	}
	return fmt.Errorf("tool %q not found in %d advertised tools", name, len(w.tools))
}

func (w *toolWorld) iCallToolWith(name, file, sql string) error {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = map[string]any{
		"file": bddsupport.ResolvePath(file),
		"sql":  sql,
	}
	w.toolRes, w.toolErr = w.client.CallTool(context.Background(), req)
	return nil
}

func (w *toolWorld) theToolCallIsNotAnError() error {
	if w.toolErr != nil {
		return fmt.Errorf("transport error: %w", w.toolErr)
	}
	if w.toolRes == nil {
		return errors.New("no tool result")
	}
	if w.toolRes.IsError {
		return fmt.Errorf("tool reported an error: %s", toolText(w.toolRes))
	}
	return nil
}

func (w *toolWorld) theToolCallIsAnErrorContaining(msg string) error {
	if w.toolErr != nil {
		return fmt.Errorf("transport error instead of tool error: %w", w.toolErr)
	}
	if w.toolRes == nil {
		return errors.New("no tool result")
	}
	if !w.toolRes.IsError {
		return fmt.Errorf("expected a tool error, got success: %s", toolText(w.toolRes))
	}
	if got := toolText(w.toolRes); !strings.Contains(got, msg) {
		return fmt.Errorf("tool error %q does not contain %q", got, msg)
	}
	return nil
}

func (w *toolWorld) theToolJSONFieldEquals(field, value string) error {
	if err := w.theToolCallIsNotAnError(); err != nil {
		return err
	}
	text := toolText(w.toolRes)
	needleNum := fmt.Sprintf(`"%s":%s`, field, value)
	needleStr := fmt.Sprintf(`"%s":"%s"`, field, value)
	if !strings.Contains(text, needleNum) && !strings.Contains(text, needleStr) {
		return fmt.Errorf("JSON %s does not contain field %q = %q", text, field, value)
	}
	return nil
}

func (w *toolWorld) theToolJSONContains(substr string) error {
	if err := w.theToolCallIsNotAnError(); err != nil {
		return err
	}
	if got := toolText(w.toolRes); !strings.Contains(got, substr) {
		return fmt.Errorf("tool JSON %q does not contain %q", got, substr)
	}
	return nil
}

func toolText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &toolWorld{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	sc.Step(`^a running MCP server$`, w.aRunningMCPServer)
	sc.Step(`^I list the MCP tools$`, w.iListTheMCPTools)
	sc.Step(`^the tool "([^"]*)" is available$`, w.theToolIsAvailable)
	sc.Step(`^I call "([^"]*)" with file "([^"]*)" and sql "([^"]*)"$`, w.iCallToolWith)
	sc.Step(`^the tool call is not an error$`, w.theToolCallIsNotAnError)
	sc.Step(`^the tool call is an error containing "([^"]*)"$`, w.theToolCallIsAnErrorContaining)
	sc.Step(`^the tool JSON field "([^"]*)" equals "([^"]*)"$`, w.theToolJSONFieldEquals)
	sc.Step(`^the tool JSON contains "(.*)"$`, w.theToolJSONContains)
}
