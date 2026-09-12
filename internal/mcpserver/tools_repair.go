package mcpserver

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerRepairTools adds the EA-representation validate/fix tools: a model
// exported before eaForElement (internal/service/sparx/archimate.go) learned
// the right EA base type for an ArchiMate type still has some elements stored
// with the wrong one, which Sparx EA renders as anonymous classes instead of
// their real ArchiMate shape. These tools find and repair that.
func registerRepairTools(s *server.MCPServer, h sparxTools) {
	s.AddTool(mcp.NewTool("ea_validate_model",
		mcp.WithDescription("Scan every element for a mismatch between its ArchiMate type and EA's underlying UML base type (uml:Class/Interface/Activity/Component) — the kind of mismatch that makes Sparx EA render an element as an anonymous class instead of its real ArchiMate shape. Read-only; use ea_fix_elements or ea_validate_and_fix_model to repair what it finds."),
		fileArg(),
	), h.read(func(m Model, _ mcp.CallToolRequest) (any, error) {
		return m.ValidateModel(), nil
	}))

	s.AddTool(mcp.NewTool("ea_fix_elements",
		mcp.WithDescription("Repair the EA representation of the given elements (ids/GUIDs/paths — normally the ids ea_validate_model reported). An element that needs no fix is skipped."),
		fileArg(), outputArg(),
		mcp.WithArray("refs", mcp.Required(), mcp.Description("element ids, GUIDs or paths to repair"), mcp.WithStringItems()),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.FixElements(r.GetStringSlice("refs", nil))
	}))

	s.AddTool(mcp.NewTool("ea_validate_and_fix_model",
		mcp.WithDescription("Scan every element and repair every EA-representation mismatch ea_validate_model would report, in one step."),
		fileArg(), outputArg(),
	), h.write(func(m Model, _ mcp.CallToolRequest) (any, error) {
		return m.ValidateAndFixModel()
	}))
}
