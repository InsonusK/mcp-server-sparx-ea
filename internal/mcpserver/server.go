// Package mcpserver wires the eapx connector to an MCP server over stdio.
package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Version is reported to MCP clients during initialization.
const Version = "0.1.0"

// Querier is the subset of *eapx.Connector the server needs. Declared as an
// interface so tests can substitute a fake without touching a real file.
type Querier interface {
	Query(sql string) (*eapx.ResultSet, error)
	Close() error
}

// Opener resolves a file path to a Querier.
type Opener func(path string) (Querier, error)

func defaultOpener(path string) (Querier, error) {
	c, err := eapx.Open(path)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// New builds the MCP server exposing the single ea_query tool. Pass nil to use
// the real eapx connector.
func New(opener Opener) *server.MCPServer {
	if opener == nil {
		opener = defaultOpener
	}

	s := server.NewMCPServer(
		"mcp-server-sparx-ea",
		Version,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	tool := mcp.NewTool("ea_query",
		mcp.WithDescription("Run a read-only SQL SELECT query against a Sparx Enterprise Architect project file (.eapx / MS Access JET database) and return the rows as JSON."),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Path to the .eapx file on the server's filesystem."),
		),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("A single read-only SQL SELECT statement."),
		),
	)

	s.AddTool(tool, handleEAQuery(opener))
	return s
}

// QueryResult is the JSON shape returned in the tool's text content.
type QueryResult struct {
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
	RowCount int        `json:"rowCount"`
}

func handleEAQuery(opener Opener) server.ToolHandlerFunc {
	return func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, err := req.RequireString("file")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: file"), nil
		}
		sql, err := req.RequireString("sql")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: sql"), nil
		}

		conn, err := opener(file)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer conn.Close()

		rs, err := conn.Query(sql)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		rows := rs.Rows
		if rows == nil {
			rows = [][]string{}
		}
		out, err := json.Marshal(QueryResult{
			Columns:  rs.Columns,
			Rows:     rows,
			RowCount: rs.RowCount,
		})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	}
}
