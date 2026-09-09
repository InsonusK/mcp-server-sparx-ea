package mcpserver

import (
	"context"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerModelTools adds the whole-model tools: the navigator tree, the type
// vocabulary, and the model-lifecycle tools (create from scratch, add / rename
// the root package).
func registerModelTools(s *server.MCPServer, h sparxTools) {
	// ---------- read ----------

	s.AddTool(mcp.NewTool("ea_model_tree",
		mcp.WithDescription("The whole model as a tree of packages, diagrams and elements with ids — like the Sparx EA project browser."),
		fileArg(),
	), h.read(func(m Model, _ mcp.CallToolRequest) (any, error) { return m.Tree(), nil }))

	s.AddTool(mcp.NewTool("ea_archimate_types",
		mcp.WithDescription("The ArchiMate element and relationship type names this server accepts."),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return jsonResult(map[string]any{
			"elementTypes":      sparx.ElementTypes(),
			"relationshipTypes": sparx.RelationshipTypes(),
		})
	})

	// ---------- lifecycle ----------

	s.AddTool(mcp.NewTool("ea_new_model",
		mcp.WithDescription("Build a model from scratch (not a copy of an export): one ArchiMate-ready root package named 'root', saved to 'output' for import into EA."),
		mcp.WithString("root", mcp.Required(), mcp.Description("the name of the single EA root package")),
		mcp.WithString("output", mcp.Required(), mcp.Description("path to write the new model to; import it into EA")),
	), h.create(func(m Model, _ mcp.CallToolRequest) (any, error) { return m.Tree(), nil }))

	s.AddTool(mcp.NewTool("ea_create_root_package",
		mcp.WithDescription("Add a new top-level package alongside the model root."),
		fileArg(), outputArg(),
		mcp.WithString("name", mcp.Required()),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CreateRootPackage(r.GetString("name", ""))
	}))

	s.AddTool(mcp.NewTool("ea_set_root_name",
		mcp.WithDescription("Rename the single EA root package. With fresh_identity it also regenerates every GUID, so the saved copy imports into EA as an independent package that can sit next to the original."),
		fileArg(), outputArg(),
		mcp.WithString("name", mcp.Required(), mcp.Description("the new root package name")),
		mcp.WithBoolean("fresh_identity", mcp.Description("also regenerate every GUID in the model")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		name, err := m.SetRootName(r.GetString("name", ""), r.GetBool("fresh_identity", false))
		if err != nil {
			return nil, err
		}
		return map[string]any{"rootName": name}, nil
	}))
}
