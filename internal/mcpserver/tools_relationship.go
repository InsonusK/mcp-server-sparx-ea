package mcpserver

import (
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerRelationshipTools adds the relationship tools: create one (with
// ArchiMate validation) and delete by id or by endpoints.
func registerRelationshipTools(s *server.MCPServer, h sparxTools) {
	s.AddTool(mcp.NewTool("ea_create_relationship",
		mcp.WithDescription("Add an ArchiMate relationship between two elements. Fails if ArchiMate does not permit it or an identical one already exists."),
		fileArg(), outputArg(),
		mcp.WithString("source", mcp.Required(), mcp.Description("source element (id, GUID or path)")),
		mcp.WithString("target", mcp.Required(), mcp.Description("target element")),
		mcp.WithString("type", mcp.Required(), mcp.Description("ArchiMate relationship type, e.g. 'ArchiMate.Realization'")),
		mcp.WithString("name", mcp.Description("connector name")),
		mcp.WithString("note", mcp.Description("documentation")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CreateRelationship(r.GetString("source", ""), r.GetString("target", ""),
			r.GetString("type", ""), r.GetString("name", ""), r.GetString("note", ""))
	}))

	s.AddTool(mcp.NewTool("ea_delete_relationship",
		mcp.WithDescription("Remove a relationship by id, or every relationship between two elements."),
		fileArg(), outputArg(),
		mcp.WithString("id", mcp.Description("the relationship id or GUID")),
		mcp.WithString("source", mcp.Description("with 'target': delete every relationship from source to target")),
		mcp.WithString("target", mcp.Description("with 'source'")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		if id := r.GetString("id", ""); id != "" {
			return deleted(m.DeleteRelationship(id))
		}
		src, tgt := r.GetString("source", ""), r.GetString("target", "")
		if src == "" || tgt == "" {
			return nil, errors.New("pass either 'id' or both 'source' and 'target'")
		}
		n, err := m.DeleteRelationshipsBetween(src, tgt)
		if err != nil {
			return nil, err
		}
		return map[string]any{"deleted": n}, nil
	}))
}
