package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// The ArchiMate tools, one file per model concept:
//
//	tools_model.go        the model tree, the type vocabulary, model lifecycle
//	tools_element.go      elements: read + create / rename / re-note / move / delete
//	tools_relationship.go relationships: create / delete
//	tools_package.go      packages: read + create / rename / move / copy / delete
//	tools_diagram.go      diagrams: read + place / move / remove an element
//	tools_repair.go       EA-representation validate / fix / validate-and-fix
//
// Read tools take a "file" (an exported .xml model). Write tools also take
// "output" and save the edited copy there, never overwriting "file".
// ea_new_model takes no "file" — it builds a model from scratch.

// deps is what the tool registrars need: how to open an existing model and how
// to build a new one.
type deps struct {
	open     SparxOpener
	newModel ModelFactory
}

// registerSparxTools adds every ArchiMate tool, grouped by model concept.
func registerSparxTools(s *server.MCPServer, d deps) {
	h := sparxTools{open: d.open, newModel: d.newModel}

	registerModelTools(s, h)
	registerElementTools(s, h)
	registerRelationshipTools(s, h)
	registerPackageTools(s, h)
	registerDiagramTools(s, h)
	registerRepairTools(s, h)
}

// ---------- plumbing ----------

type sparxTools struct {
	open     SparxOpener
	newModel ModelFactory
}

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

// create wires a from-scratch tool: build a new model from "root", run fn, save
// to "output". There is no "file" — nothing is opened.
func (h sparxTools) create(fn toolFn) server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		root, err := req.RequireString("root")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: root"), nil
		}
		output, err := req.RequireString("output")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: output"), nil
		}
		m, err := h.newModel(root)
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
