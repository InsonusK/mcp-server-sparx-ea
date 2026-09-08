package mcpserver

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type fakeConn struct {
	rs       *eapx.ResultSet
	queryErr error
	closed   bool
}

func (f *fakeConn) Query(string) (*eapx.ResultSet, error) { return f.rs, f.queryErr }
func (f *fakeConn) Close() error                          { f.closed = true; return nil }

func call(t *testing.T, opener Opener, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "ea_query"
	req.Params.Arguments = args
	res, err := handleEAQuery(opener)(context.Background(), req)
	if err != nil {
		t.Fatalf("handler returned transport error: %v", err)
	}
	return res
}

func text(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestHandlerReturnsRowsAsJSON(t *testing.T) {
	fc := &fakeConn{rs: &eapx.ResultSet{
		Columns:  []string{"Object_ID", "Name"},
		Rows:     [][]string{{"1", "Pkg"}, {"2", "Cls"}},
		RowCount: 2,
	}}
	res := call(t, func(string) (Querier, error) { return fc, nil },
		map[string]any{"file": "x.eapx", "sql": "select Object_ID, Name from t_object"})

	if res.IsError {
		t.Fatalf("unexpected tool error: %s", text(res))
	}
	got := text(res)
	for _, want := range []string{`"columns":["Object_ID","Name"]`, `"rowCount":2`, `"Cls"`} {
		if !strings.Contains(got, want) {
			t.Errorf("result %s missing %q", got, want)
		}
	}
	if !fc.closed {
		t.Error("connector was not closed")
	}
}

func TestHandlerEmptyRowsSerializeAsArray(t *testing.T) {
	fc := &fakeConn{rs: &eapx.ResultSet{Columns: []string{"Object_ID"}, Rows: nil, RowCount: 0}}
	res := call(t, func(string) (Querier, error) { return fc, nil },
		map[string]any{"file": "x.eapx", "sql": "select Object_ID from t_object"})
	if got := text(res); !strings.Contains(got, `"rows":[]`) {
		t.Errorf("nil rows should serialize as [], got %s", got)
	}
}

func TestHandlerOpenErrorBecomesToolError(t *testing.T) {
	res := call(t, func(string) (Querier, error) { return nil, errors.New("eapx open: file not found") },
		map[string]any{"file": "missing.eapx", "sql": "select 1"})
	if !res.IsError || !strings.Contains(text(res), "file not found") {
		t.Fatalf("expected tool error with cause, got IsError=%v %s", res.IsError, text(res))
	}
}

func TestHandlerQueryErrorBecomesToolError(t *testing.T) {
	fc := &fakeConn{queryErr: errors.New("eapx query: Got no result")}
	res := call(t, func(string) (Querier, error) { return fc, nil },
		map[string]any{"file": "x.eapx", "sql": "select bad"})
	if !res.IsError || !strings.Contains(text(res), "Got no result") {
		t.Fatalf("expected tool error, got IsError=%v %s", res.IsError, text(res))
	}
}

func TestHandlerMissingArguments(t *testing.T) {
	opener := func(string) (Querier, error) { return &fakeConn{}, nil }
	for _, args := range []map[string]any{
		{"sql": "select 1"},
		{"file": "x.eapx"},
	} {
		res := call(t, opener, args)
		if !res.IsError {
			t.Errorf("args %v: expected a tool error", args)
		}
	}
}

// TestNewServesEaQueryEndToEnd drives New(nil) — the production wiring — through
// a real in-process MCP client against the sample project, proving the default
// opener is installed and the tool round-trips.
func TestNewServesEaQueryEndToEnd(t *testing.T) {
	fixture, err := filepath.Abs(filepath.Join("..", "..", "example", "TestProject.eapx"))
	if err != nil {
		t.Fatal(err)
	}

	c, err := client.NewInProcessClient(New(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		t.Fatal(err)
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		t.Fatal(err)
	}

	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "ea_query" {
		t.Fatalf("expected exactly the ea_query tool, got %+v", tools.Tools)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = "ea_query"
	req.Params.Arguments = map[string]any{"file": fixture, "sql": "select count(*) from t_object"}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool call failed: %s", text(res))
	}
	if got := text(res); !strings.Contains(got, `"rows":[["16"]]`) {
		t.Fatalf("unexpected result payload: %s", got)
	}
}
