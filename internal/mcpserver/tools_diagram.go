package mcpserver

import (
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerDiagramTools adds the diagram tools: read one diagram, create a
// diagram, and place / move / remove an element on it.
func registerDiagramTools(s *server.MCPServer, h sparxTools) {
	s.AddTool(mcp.NewTool("ea_diagram",
		mcp.WithDescription("One diagram by id, GUID or slash path: its type and the elements placed on it with coordinates."),
		fileArg(), refArg("the diagram"),
	), h.read(func(m Model, r mcp.CallToolRequest) (any, error) { return m.Diagram(r.GetString("ref", "")) }))

	s.AddTool(mcp.NewTool("ea_create_diagram",
		mcp.WithDescription("Add an empty diagram to a package, then fill it with ea_place_on_diagram. "+
			"'layer' is an optional ArchiMate viewpoint (Motivation, Strategy, Business, Application, "+
			"Technology, Physical, Implementation_Migration) that picks EA's toolbox; omit it for a plain diagram."),
		fileArg(), outputArg(),
		mcp.WithString("parent", mcp.Required(), mcp.Description("the package to add the diagram to (id, GUID or path)")),
		mcp.WithString("name", mcp.Required()),
		mcp.WithString("layer", mcp.Description("ArchiMate viewpoint / EA toolbox (optional)")),
	), h.write(func(m Model, r mcp.CallToolRequest) (any, error) {
		return m.CreateDiagram(r.GetString("parent", ""), r.GetString("name", ""), r.GetString("layer", ""))
	}))

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
