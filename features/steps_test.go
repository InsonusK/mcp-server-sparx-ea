package features

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/cucumber/godog"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/InsonusK/mcp-server-sparx-ea/eapx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/mcpserver"
)

// repoRoot is the module root relative to this package's working directory.
const repoRoot = ".."

// resolvePath maps a path as written in a .feature file to a real path.
// Relative paths are taken relative to the repository root; "" stays "".
func resolvePath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(repoRoot, p)
}

type world struct {
	// connector
	pathArg  string
	conn     *eapx.Connector
	openErr  error
	rs       *eapx.ResultSet
	queryErr error

	// temp-dir isolation
	prevTMPDIR    string
	hadTMPDIR     bool
	tmpDir        string
	tmpDirApplied bool

	// mcp
	mcpClient *client.Client
	tools     []mcp.Tool
	toolRes   *mcp.CallToolResult
	toolErr   error

	// concurrency
	concN    int
	concErrs []error
	concVals []string
}

func (w *world) reset() {
	if w.conn != nil {
		_ = w.conn.Close()
	}
	if w.mcpClient != nil {
		_ = w.mcpClient.Close()
	}
	w.restoreTMPDIR()
	*w = world{}
}

func (w *world) restoreTMPDIR() {
	if !w.tmpDirApplied {
		return
	}
	if w.hadTMPDIR {
		_ = os.Setenv("TMPDIR", w.prevTMPDIR)
	} else {
		_ = os.Unsetenv("TMPDIR")
	}
	w.tmpDirApplied = false
}

// ---------- connector: open ----------

func (w *world) theFilePath(path string) error {
	w.pathArg = resolvePath(path)
	return nil
}

func (w *world) iOpenTheConnection() error {
	w.conn, w.openErr = eapx.Open(w.pathArg)
	return nil
}

func (w *world) openingOutcome(outcome string) error {
	switch {
	case outcome == "succeeds":
		if w.openErr != nil {
			return fmt.Errorf("expected open to succeed, got error: %v", w.openErr)
		}
		if w.conn == nil {
			return errors.New("expected a connection, got nil")
		}
		return nil
	case strings.HasPrefix(outcome, "fails with an error containing "):
		want := strings.Trim(strings.TrimPrefix(outcome, "fails with an error containing "), `"`)
		return assertErrContains(w.openErr, want)
	default:
		return fmt.Errorf("unknown outcome %q", outcome)
	}
}

func (w *world) aConnectionTo(path string) error {
	conn, err := eapx.Open(resolvePath(path))
	if err != nil {
		return fmt.Errorf("could not open %s: %w", path, err)
	}
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.conn = conn
	return nil
}

// ---------- connector: query ----------

func (w *world) iRunTheQuery(sql string) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	w.rs, w.queryErr = w.conn.Query(sql)
	return nil
}

func (w *world) iCloseTheConnection() error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	return w.conn.Close()
}

func (w *world) theQuerySucceeds() error {
	if w.queryErr != nil {
		return fmt.Errorf("expected query to succeed, got: %v", w.queryErr)
	}
	if w.rs == nil {
		return errors.New("expected a result set, got nil")
	}
	return nil
}

func (w *world) theQueryFailsContaining(msg string) error {
	if w.rs != nil {
		return fmt.Errorf("expected query to fail, but it returned %d rows", w.rs.RowCount)
	}
	return assertErrContains(w.queryErr, msg)
}

func (w *world) theResultHasNRows(n int) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if got := len(w.rs.Rows); got != n {
		return fmt.Errorf("expected %d rows, got %d", n, got)
	}
	if w.rs.RowCount != n {
		return fmt.Errorf("RowCount=%d disagrees with len(Rows)=%d", w.rs.RowCount, n)
	}
	return nil
}

func (w *world) theResultColumnsAre(cols string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	want := splitCSV(cols)
	if !equalStrings(w.rs.Columns, want) {
		return fmt.Errorf("expected columns %v, got %v", want, w.rs.Columns)
	}
	return nil
}

func (w *world) rowEquals(idx int, csv string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if idx < 0 || idx >= len(w.rs.Rows) {
		return fmt.Errorf("row %d out of range (%d rows)", idx, len(w.rs.Rows))
	}
	want := splitCSV(csv)
	if !equalStrings(w.rs.Rows[idx], want) {
		return fmt.Errorf("row %d: expected %v, got %v", idx, want, w.rs.Rows[idx])
	}
	return nil
}

func (w *world) theSingleResultValueIs(val string) error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	if len(w.rs.Rows) != 1 || len(w.rs.Rows[0]) != 1 {
		return fmt.Errorf("expected exactly one row with one column, got %d rows / columns %v", len(w.rs.Rows), w.rs.Columns)
	}
	if w.rs.Rows[0][0] != val {
		return fmt.Errorf("expected value %q, got %q", val, w.rs.Rows[0][0])
	}
	return nil
}

func (w *world) everyResultValueIsValidUTF8() error {
	if err := w.theQuerySucceeds(); err != nil {
		return err
	}
	for r, row := range w.rs.Rows {
		for c, cell := range row {
			if !utf8.ValidString(cell) {
				return fmt.Errorf("row %d col %d (%s) is not valid UTF-8: %q", r, c, w.rs.Columns[c], cell)
			}
		}
	}
	return nil
}

// ---------- technical: temp dir ----------

func (w *world) emptyDirAsTempDir() error {
	dir, err := os.MkdirTemp("", "eapx-tmpdir-*")
	if err != nil {
		return err
	}
	w.prevTMPDIR, w.hadTMPDIR = os.LookupEnv("TMPDIR")
	if err := os.Setenv("TMPDIR", dir); err != nil {
		return err
	}
	w.tmpDir = dir
	w.tmpDirApplied = true
	return nil
}

func (w *world) iRunTheQueryNTimes(sql string, n int) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	for i := 0; i < n; i++ {
		if _, err := w.conn.Query(sql); err != nil {
			return fmt.Errorf("query %d failed: %w", i, err)
		}
	}
	return nil
}

func (w *world) theTempDirIsStillEmpty() error {
	entries, err := os.ReadDir(w.tmpDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		return fmt.Errorf("temp dir is not empty: %v", names)
	}
	return nil
}

// ---------- technical: no os/exec ----------

func noProjectPackageImportsOsExec(module string) error {
	cmd := exec.Command("go", "list", "-deps",
		"-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./...")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go list failed: %v\n%s", err, out)
	}
	var offenders []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], module) {
			continue
		}
		for _, imp := range fields[1:] {
			if imp == "os/exec" {
				offenders = append(offenders, fields[0])
			}
		}
	}
	if len(offenders) != 0 {
		return fmt.Errorf("packages import os/exec: %v", offenders)
	}
	return nil
}

// ---------- technical: binary linkage ----------

func (w *world) theServerBinaryIsBuilt() error {
	outPath, err := filepath.Abs(filepath.Join(repoRoot, "tmp", "bin", "mcp-server-sparx-ea"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %v\n%s", err, out)
	}
	w.pathArg = outPath
	return nil
}

func (w *world) binaryLinkedAgainst(lib string) error {
	out, err := exec.Command("ldd", w.pathArg).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ldd failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), lib) {
		return fmt.Errorf("binary is not linked against %q:\n%s", lib, out)
	}
	return nil
}

// ---------- technical: cyrillic fixture (skipped when absent) ----------

func theCyrillicFixtureIsAvailable(path string) error {
	if _, err := os.Stat(resolvePath(path)); err != nil {
		return fmt.Errorf("%w: fixture %s is not present (Sparx EA is needed to author a Cyrillic .eapx)", godog.ErrSkip, path)
	}
	return nil
}

// ---------- technical: concurrency ----------

func (w *world) iRunQueryFromNGoroutines(sql string, n int) error {
	if w.conn == nil {
		return errors.New("no open connection")
	}
	w.concN = n
	w.concErrs = make([]error, n)
	w.concVals = make([]string, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			rs, err := w.conn.Query(sql)
			if err != nil {
				w.concErrs[i] = err
				return
			}
			if len(rs.Rows) == 1 && len(rs.Rows[0]) == 1 {
				w.concVals[i] = rs.Rows[0][0]
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func (w *world) allNQueriesSucceed(n int) error {
	if w.concN != n {
		return fmt.Errorf("ran %d goroutines, step expects %d", w.concN, n)
	}
	for i, err := range w.concErrs {
		if err != nil {
			return fmt.Errorf("goroutine %d failed: %w", i, err)
		}
	}
	return nil
}

func (w *world) everyConcurrentQueryReturned(val string) error {
	for i, got := range w.concVals {
		if got != val {
			return fmt.Errorf("goroutine %d returned %q, want %q", i, got, val)
		}
	}
	return nil
}

// ---------- mcp tool ----------

func (w *world) aRunningMCPServer() error {
	c, err := client.NewInProcessClient(mcpserver.New(nil))
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		return err
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "godog", Version: "1.0.0"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		return err
	}
	w.mcpClient = c
	return nil
}

func (w *world) iListTheMCPTools() error {
	res, err := w.mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return err
	}
	w.tools = res.Tools
	return nil
}

func (w *world) theToolIsAvailable(name string) error {
	for _, t := range w.tools {
		if t.Name == name {
			return nil
		}
	}
	return fmt.Errorf("tool %q not found in %d advertised tools", name, len(w.tools))
}

func (w *world) iCallToolWith(name, file, sql string) error {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = map[string]any{
		"file": resolvePath(file),
		"sql":  sql,
	}
	w.toolRes, w.toolErr = w.mcpClient.CallTool(context.Background(), req)
	return nil
}

func (w *world) theToolCallIsNotAnError() error {
	if w.toolErr != nil {
		return fmt.Errorf("transport error: %w", w.toolErr)
	}
	if w.toolRes == nil {
		return errors.New("no tool result")
	}
	if w.toolRes.IsError {
		return fmt.Errorf("tool reported an error: %s", toolText(w.toolRes))
	}
	return nil
}

func (w *world) theToolCallIsAnErrorContaining(msg string) error {
	if w.toolErr != nil {
		return fmt.Errorf("transport error instead of tool error: %w", w.toolErr)
	}
	if w.toolRes == nil {
		return errors.New("no tool result")
	}
	if !w.toolRes.IsError {
		return fmt.Errorf("expected a tool error, got success: %s", toolText(w.toolRes))
	}
	if got := toolText(w.toolRes); !strings.Contains(got, msg) {
		return fmt.Errorf("tool error %q does not contain %q", got, msg)
	}
	return nil
}

func (w *world) theToolJSONFieldEquals(field, value string) error {
	if err := w.theToolCallIsNotAnError(); err != nil {
		return err
	}
	text := toolText(w.toolRes)
	// The handler returns a flat JSON object; do a tolerant check on "field":value.
	needleNum := fmt.Sprintf(`"%s":%s`, field, value)
	needleStr := fmt.Sprintf(`"%s":"%s"`, field, value)
	if !strings.Contains(text, needleNum) && !strings.Contains(text, needleStr) {
		return fmt.Errorf("JSON %s does not contain field %q = %q", text, field, value)
	}
	return nil
}

func (w *world) theToolJSONContains(substr string) error {
	if err := w.theToolCallIsNotAnError(); err != nil {
		return err
	}
	if got := toolText(w.toolRes); !strings.Contains(got, substr) {
		return fmt.Errorf("tool JSON %q does not contain %q", got, substr)
	}
	return nil
}

// ---------- helpers ----------

func toolText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := mcp.AsTextContent(c); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func assertErrContains(err error, want string) error {
	if err == nil {
		return fmt.Errorf("expected an error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", err.Error(), want)
	}
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------- registration ----------

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	// open
	sc.Step(`^the file path "([^"]*)"$`, w.theFilePath)
	sc.Step(`^I open the connection$`, w.iOpenTheConnection)
	sc.Step(`^opening (.+)$`, w.openingOutcome)
	sc.Step(`^a connection to the Sparx EA file "([^"]*)"$`, w.aConnectionTo)

	// query
	sc.Step(`^I run the query "([^"]*)"$`, w.iRunTheQuery)
	sc.Step(`^I run the query "([^"]*)" (\d+) times$`, w.iRunTheQueryNTimes)
	sc.Step(`^I close the connection$`, w.iCloseTheConnection)
	sc.Step(`^the query succeeds$`, w.theQuerySucceeds)
	sc.Step(`^the query fails with an error containing "([^"]*)"$`, w.theQueryFailsContaining)
	sc.Step(`^the result has (\d+) rows$`, w.theResultHasNRows)
	sc.Step(`^the result columns are "([^"]*)"$`, w.theResultColumnsAre)
	sc.Step(`^row (\d+) equals "([^"]*)"$`, w.rowEquals)
	sc.Step(`^the single result value is "([^"]*)"$`, w.theSingleResultValueIs)
	sc.Step(`^every result value is valid UTF-8$`, w.everyResultValueIsValidUTF8)

	// technical: temp dir
	sc.Step(`^an empty directory registered as the process temp dir$`, w.emptyDirAsTempDir)
	sc.Step(`^the temp dir is still empty$`, w.theTempDirIsStillEmpty)

	// technical: os/exec
	sc.Step(`^no package under "([^"]*)" imports "os/exec"$`, noProjectPackageImportsOsExec)

	// technical: binary
	sc.Step(`^the server binary is built$`, w.theServerBinaryIsBuilt)
	sc.Step(`^the binary is dynamically linked against "([^"]*)"$`, w.binaryLinkedAgainst)

	// technical: cyrillic
	sc.Step(`^the Cyrillic fixture "([^"]*)" is available$`, theCyrillicFixtureIsAvailable)

	// technical: concurrency
	sc.Step(`^I run the query "([^"]*)" from (\d+) goroutines concurrently$`, w.iRunQueryFromNGoroutines)
	sc.Step(`^all (\d+) queries succeed$`, w.allNQueriesSucceed)
	sc.Step(`^every concurrent query returned the value "([^"]*)"$`, w.everyConcurrentQueryReturned)

	// mcp
	sc.Step(`^a running MCP server$`, w.aRunningMCPServer)
	sc.Step(`^I list the MCP tools$`, w.iListTheMCPTools)
	sc.Step(`^the tool "([^"]*)" is available$`, w.theToolIsAvailable)
	sc.Step(`^I call "([^"]*)" with file "([^"]*)" and sql "([^"]*)"$`, w.iCallToolWith)
	sc.Step(`^the tool call is not an error$`, w.theToolCallIsNotAnError)
	sc.Step(`^the tool call is an error containing "([^"]*)"$`, w.theToolCallIsAnErrorContaining)
	sc.Step(`^the tool JSON field "([^"]*)" equals "([^"]*)"$`, w.theToolJSONFieldEquals)
	sc.Step(`^the tool JSON contains "(.*)"$`, w.theToolJSONContains)
}
