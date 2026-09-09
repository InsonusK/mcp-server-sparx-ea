// Package mcpserver exposes the ArchiMate service over an MCP server (stdio).
//
//	ea_model_tree / ea_element /    read an exported .xml model
//	  ea_package / ea_diagram
//	ea_archimate_types             the type vocabulary
//	ea_create_element / …          edit a copy of the model and save it
//
// Tool handlers depend on the Model interface; New(nil) wires the real XMI
// service, tests pass a fake through Options.
package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
)

// Version is reported to MCP clients during initialization.
const Version = "0.4.0"

// Options overrides the tool handlers' dependencies. A nil *Options (or a nil
// field) uses the real XMI service.
type Options struct {
	Sparx SparxOpener // path -> Model (ArchiMate)
}

func (o *Options) sparx() SparxOpener {
	if o != nil && o.Sparx != nil {
		return o.Sparx
	}
	return defaultSparxOpener
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

	registerSparxTools(s, o.sparx())

	return s
}
