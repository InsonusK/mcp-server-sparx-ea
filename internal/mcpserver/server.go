// Package mcpserver exposes the ArchiMate service over an MCP server (stdio).
//
//	ea_model_tree / ea_element /    read an exported .xml model
//	  ea_package / ea_diagram
//	ea_archimate_types             the type vocabulary
//	ea_new_model / …               build a model from scratch, add / rename the root
//	ea_create_element / …          edit a copy of the model and save it
//
// Tool handlers depend on the Model interface; New(nil) wires the real XMI
// service, tests pass a fake through Options.
package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
)

// Version is reported to MCP clients during initialization.
const Version = "0.8.2"

// Options overrides the tool handlers' dependencies. A nil *Options (or a nil
// field) uses the real XMI service.
type Options struct {
	Sparx    SparxOpener  // path -> Model (open an existing export)
	NewModel ModelFactory // root name -> Model (build from scratch)
}

func (o *Options) sparx() SparxOpener {
	if o != nil && o.Sparx != nil {
		return o.Sparx
	}
	return defaultSparxOpener
}

func (o *Options) newModel() ModelFactory {
	if o != nil && o.NewModel != nil {
		return o.NewModel
	}
	return defaultModelFactory
}

// New builds the MCP server with every tool registered. Pass nil for the real
// implementation.
func New(o *Options) *server.MCPServer {
	s := server.NewMCPServer(
		"mcp-server-sparx-ea",
		Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerSparxTools(s, deps{open: o.sparx(), newModel: o.newModel()})

	return s
}
