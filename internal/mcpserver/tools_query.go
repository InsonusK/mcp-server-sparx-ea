package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Querier is the subset of *eapx.Connector the ea_query handler needs.
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

// QueryResult is the JSON shape returned in ea_query's text content.
type QueryResult struct {
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
	RowCount int        `json:"rowCount"`
}

func registerQueryTool(s *server.MCPServer, open Opener) {
	tool := mcp.NewTool("ea_query",
		mcp.WithDescription("Run a read-only SQL SELECT query against a Sparx Enterprise Architect project file (.eapx / MS Access JET database) and return the rows as JSON."),
		mcp.WithString("file", mcp.Required(), mcp.Description("Path to the .eapx file on the server's filesystem.")),
		mcp.WithString("sql", mcp.Required(), mcp.Description("A single read-only SQL SELECT statement.")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, err := req.RequireString("file")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: file"), nil
		}
		sql, err := req.RequireString("sql")
		if err != nil {
			return mcp.NewToolResultError("missing required argument: sql"), nil
		}

		conn, err := open(file)
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
		out, err := json.Marshal(QueryResult{Columns: rs.Columns, Rows: rows, RowCount: rs.RowCount})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(out)), nil
	})
}
