package mcpserver

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerSparxTools adds every ArchiMate tool. Read tools take a "file"
// (an exported .xml model); write tools also take "output" and save the edited
// copy there, never overwriting "file".
func registerSparxTools(s *server.MCPServer, open SparxOpener) {
	h := sparxTools{open: open}

	// ---------- read ----------

	s.AddTool(mcp.NewTool("ea_model_tree",
		mcp.WithDescription("The whole model as a tree of packages, diagrams and elements with ids — like the Sparx EA project browser."),
		fileArg(),
	), h.read(func(m Model, _ mcp.CallToolRequest) (any, error) { return m.Tree(), nil }))

	s.AddTool(mcp.NewTool("ea_element",
		mcp.WithDescription("One element by id, GUID or slash path: its ArchiMate type, note and every relationship."),
		fileArg(), refArg("the element (id, GUID or 'Model/Package/Element')"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Element(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_package",
		mcp.WithDescription("One package by id, GUID or slash path: its parent and immediate children."),
		fileArg(), refArg("the package"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Package(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_diagram",
		mcp.WithDescription("One diagram by id, GUID or slash path: its type and the elements placed on it with coordinates."),
		fileArg(), refArg("the diagram"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Diagram(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_archimate_types",
		mcp.WithDescription("The ArchiMate element and relationship type names this server accepts."),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return jsonResult(map[string]any{
			"elementTypes":      sparx.ElementTypes(),
			"relationshipTypes": sparx.RelationshipTypes(),
		})
	})

	// ---------- elements ----------

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

	// ---------- relationships ----------

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

	// ---------- packages ----------

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

	// ---------- diagram placement ----------

	place := func(op func(m Model, d, e string, at sparx.Rect) error) toolFn {
		return func(m Model, r mcp.CallToolRequest) (any, error) {
			at := sparx.Rect{Left: mustInt(r, "left"), Top: mustInt(r, "top"), Right: mustInt(r, "right"), Bottom: mustInt(r, "bottom")}
			if err := op(m, r.GetString("diagram", ""), r.GetString("element", ""), at); err != nil {
				return nil, err
			}
			return m.Diagram(r.GetString("diagram", ""))
		}
	}

	s.AddTool(diagramTool("ea_place_on_diagram",
		"Place an element on a diagram at a rectangle (EA coordinates: origin top-left).", true),
		h.write(place(func(m Model, d, e string, at sparx.Rect) error { return m.AddToDiagram(d, e, at) })))

	s.AddTool(diagramTool("ea_move_on_diagram",
		"Move an element already on a diagram to a new rectangle.", true),
		h.write(place(func(m Model, d, e string, at sparx.Rect) error { return m.MoveOnDiagram(d, e, at) })))

	s.AddTool(diagramTool("ea_remove_from_diagram",
		"Remove an element's placement from a diagram (the element itself stays in the model).", false),
		h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
			if err := m.RemoveFromDiagram(r.GetString("diagram", ""), r.GetString("element", "")); err != nil {
				return nil, err
			}
			return m.Diagram(r.GetString("diagram", ""))
		}))
}

func diagramTool(name, desc string, rect bool) mcp.Tool {
	opts := []mcp.ToolOption{
		mcp.WithDescription(desc),
		mcp.WithString("file", mcp.Required(), mcp.Description("path to the exported .xml model")),
		mcp.WithString("output", mcp.Required(), mcp.Description("path to write the edited copy to (must differ from 'file')")),
		mcp.WithString("diagram", mcp.Required(), mcp.Description("the diagram (id, GUID or path)")),
		mcp.WithString("element", mcp.Required(), mcp.Description("the element")),
	}
	if rect {
		opts = append(opts,
			mcp.WithNumber("left", mcp.Required()),
			mcp.WithNumber("top", mcp.Required()),
			mcp.WithNumber("right", mcp.Required()),
			mcp.WithNumber("bottom", mcp.Required()),
		)
	}
	return mcp.NewTool(name, opts...)
}

// ---------- plumbing ----------

type sparxTools struct{ open SparxOpener }

type toolFn func(Model, mcp.CallToolRequest) (any, error)

// read wires a read-only tool: open the file, run fn, return its JSON.
func (h sparxTools) read(fn toolFn) server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, err := req.RequireString("file")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: file"), nil
		}
		m, err := h.open(file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := fn(m, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(res)
	}
}

// write wires an editing tool: open the file, run fn, save to "output".
func (h sparxTools) write(fn toolFn) server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, err := req.RequireString("file")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: file"), nil
		}
		output, err := req.RequireString("output")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: output"), nil
		}
		if output == file {
			return mcp.NewToolResultError("output must be a different path from file"), nil
		}
		m, err := h.open(file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		res, err := fn(m, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := m.Save(output); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(map[string]any{"result": res, "saved": output})
	}
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func deleted(err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return map[string]any{"deleted": true}, nil
}

func mustInt(r mcp.CallToolRequest, key string) int {
	n, _ := r.RequireInt(key)
	return n
}

// ---------- tool-arg shorthands ----------

func fileArg() mcp.ToolOption {
	return mcp.WithString("file", mcp.Required(), mcp.Description("path to the exported .xml model on the server's filesystem"))
}
func outputArg() mcp.ToolOption {
	return mcp.WithString("output", mcp.Required(), mcp.Description("path to write the edited copy to (must differ from 'file'); import it into EA"))
}
func refArg(what string) mcp.ToolOption {
	return mcp.WithString("ref", mcp.Required(), mcp.Description(what))
}
