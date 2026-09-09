// Package mcpserver exposes the eapx SQL connector and the ArchiMate service
// over an MCP server (stdio).
//
//	ea_query                       read-only SQL against a .eapx file
//	ea_model_tree / ea_element /    read an exported .xml model
//	  ea_package / ea_diagram
//	ea_archimate_types             the type vocabulary
//	ea_create_element / …          edit a copy of the model and save it
//
// Tool handlers depend on interfaces (Querier, Model); New(nil) wires the real
// implementations, tests pass fakes through Options.
package mcpserver

import (
	"github.com/mark3labs/mcp-go/server"
)

// Version is reported to MCP clients during initialization.
const Version = "0.3.0"

// Options overrides the tool handlers' dependencies. A nil *Options (or nil
// fields) uses the real eapx connector and XMI service.
type Options struct {
	EAPX  Opener      // path -> Querier (SQL)
	Sparx SparxOpener // path -> Model (ArchiMate)
}

func (o *Options) eapx() Opener {
	if o != nil && o.EAPX != nil {
		return o.EAPX
	}
	return defaultOpener
}

func (o *Options) sparx() SparxOpener {
	if o != nil && o.Sparx != nil {
		return o.Sparx
	}
	return defaultSparxOpener
}

// New builds the MCP server with every tool registered. Pass nil for the real
// implementations.
func New(o *Options) *server.MCPServer {
	s := server.NewMCPServer(
		"mcp-server-sparx-ea",
		Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	registerQueryTool(s, o.eapx())
	registerSparxTools(s, o.sparx())

	return s
}
