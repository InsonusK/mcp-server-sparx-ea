// Command mcp-server-sparx-ea is an MCP server that answers SQL queries against
// Sparx Enterprise Architect project files (.eapx) over stdio.
package main

import (
	"fmt"
	"os"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/mcpserver"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if err := server.ServeStdio(mcpserver.New(nil)); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-server-sparx-ea:", err)
		os.Exit(1)
	}
}
