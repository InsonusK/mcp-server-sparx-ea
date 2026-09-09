package mcpserver

import (
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerElementTools adds the element tools: read one element, and
// create / rename / re-note / move / delete.
func registerElementTools(s *server.MCPServer, h sparxTools) {
	s.AddTool(mcp.NewTool("ea_element",
		mcp.WithDescription("One element by id, GUID or slash path: its ArchiMate type, note and every relationship."),
		fileArg(), refArg("the element (id, GUID or 'Model/Package/Element')"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Element(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_create_element",
		mcp.WithDescription("Add an ArchiMate element (e.g. 'ArchiMate.Goal') to a package."),
		fileArg(), outputArg(),
		mcp.WithString("parent", mcp.Required(), mcp.Description("the package to add it to (id, GUID or path)")),
		mcp.WithString("type", mcp.Required(), mcp.Description("ArchiMate element type, e.g. 'ArchiMate.Requirement'")),
		mcp.WithString("name", mcp.Required()),
		mcp.WithString("note", mcp.Description("documentation")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CreateElement(r.GetString("parent", ""), r.GetString("type", ""), r.GetString("name", ""), r.GetString("note", ""))
	}))

	s.AddTool(mcp.NewTool("ea_update_element",
		mcp.WithDescription("Rename an element, change its note and/or move it to another package. Only the fields you pass are changed."),
		fileArg(), outputArg(), refArg("the element"),
		mcp.WithString("name", mcp.Description("new name")),
		mcp.WithString("note", mcp.Description("new documentation")),
		mcp.WithString("parent", mcp.Description("move it into this package")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		ref := r.GetString("ref", "")
		var info *sparx.ElementInfo
		var err error
		if v := r.GetString("name", ""); v != "" {
			if info, err = m.RenameElement(ref, v); err != nil {
				return nil, err
			}
			ref = info.Path
		}
		if _, ok := r.GetArguments()["note"]; ok {
			if info, err = m.SetElementDocumentation(ref, r.GetString("note", "")); err != nil {
				return nil, err
			}
		}
		if v := r.GetString("parent", ""); v != "" {
			if info, err = m.MoveElement(ref, v); err != nil {
				return nil, err
			}
		}
		if info == nil {
			return m.Element(ref)
		}
		return info, nil
	}))

	s.AddTool(mcp.NewTool("ea_delete_element",
		mcp.WithDescription("Remove an element and every relationship attached to it."),
		fileArg(), outputArg(), refArg("the element"),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return deleted(m.DeleteElement(r.GetString("ref", "")))
	}))
}
