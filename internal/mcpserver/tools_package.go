package mcpserver

import (
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerPackageTools adds the package tools: read one package, and
// create / rename / move / deep-copy / delete.
func registerPackageTools(s *server.MCPServer, h sparxTools) {
	s.AddTool(mcp.NewTool("ea_package",
		mcp.WithDescription("One package by id, GUID or slash path: its parent and immediate children."),
		fileArg(), refArg("the package"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Package(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_create_package",
		mcp.WithDescription("Add a package inside another package."),
		fileArg(), outputArg(),
		mcp.WithString("parent", mcp.Required()),
		mcp.WithString("name", mcp.Required()),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CreatePackage(r.GetString("parent", ""), r.GetString("name", ""))
	}))

	s.AddTool(mcp.NewTool("ea_update_package",
		mcp.WithDescription("Rename a package and/or move it to another parent."),
		fileArg(), outputArg(), refArg("the package"),
		mcp.WithString("name", mcp.Description("new name")),
		mcp.WithString("parent", mcp.Description("move it into this package")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		ref := r.GetString("ref", "")
		var info *sparx.PackageInfo
		var err error
		if v := r.GetString("name", ""); v != "" {
			if info, err = m.RenamePackage(ref, v); err != nil {
				return nil, err
			}
			ref = info.Path
		}
		if v := r.GetString("parent", ""); v != "" {
			if info, err = m.MovePackage(ref, v); err != nil {
				return nil, err
			}
		}
		if info == nil {
			return m.Package(ref)
		}
		return info, nil
	}))

	s.AddTool(mcp.NewTool("ea_copy_package",
		mcp.WithDescription("Deep-copy a package (with a fresh identity) into another package."),
		fileArg(), outputArg(), refArg("the package to copy"),
		mcp.WithString("dest", mcp.Required(), mcp.Description("the package to copy it into")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CopyPackage(r.GetString("ref", ""), r.GetString("dest", ""))
	}))

	s.AddTool(mcp.NewTool("ea_delete_package",
		mcp.WithDescription("Remove a package. Without cascade it refuses a non-empty package."),
		fileArg(), outputArg(), refArg("the package"),
		mcp.WithBoolean("cascade", mcp.Description("also delete every sub-package, element and diagram inside it")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		n, err := m.DeletePackage(r.GetString("ref", ""), r.GetBool("cascade", false))
		if err != nil {
			return nil, err
		}
		return map[string]any{"removed": n}, nil
	}))
}
