// Package eaxmitest is the black-box Cucumber suite for the eaxmi codec.
// It imports client/eaxmi and exercises only its public API.
//
// Conventions: docs/skills/cucumber-go-testing.md
//   - every step takes ctx first and logs what it did via logf
//   - expected data lives in the .feature (data tables), not hard-coded here
//   - fixtures live in testdata/; write tests save to tmp/ (gitignored)
package eaxmitest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

const tmpDir = "tmp"

// world is one scenario's state.
type world struct {
	doc *eaxmi.Document
	err error

	// last resolved things
	elem *eaxmi.Element
	conn *eaxmi.Connector
	pkg  *eaxmi.Package
	diag *eaxmi.Diagram
	path string

	// a second document, for cross-document copy
	other *eaxmi.Document

	// a reopened working copy (after WriteFile)
	reloaded  *eaxmi.Document
	savedPath string

	// snapshot of GUIDs by path, for identity assertions
	guidBefore map[string]string
}

func newWorld() *world { return &world{} }

func (w *world) reset() { *w = world{} }

func logf(_ context.Context, format string, a ...any) {
	fmt.Printf("      · "+format+"\n", a...)
}

func fixturePath(name string) string {
	if name == "" || filepath.IsAbs(name) {
		return name
	}
	return filepath.Join("testdata", name)
}

// active is the document a step operates on: the reopened working copy if there
// is one, else the in-memory doc.
func (w *world) active() *eaxmi.Document {
	if w.reloaded != nil {
		return w.reloaded
	}
	return w.doc
}

// open loads a fixture into w.doc, failing the scenario if it cannot be parsed.
func (w *world) open(ctx context.Context, name string) error {
	d, err := eaxmi.Open(fixturePath(name))
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	w.doc = d
	logf(ctx, "opened %q → root %q, %d element(s), %d connector(s)",
		name, d.Root.Name, len(d.AllElements()), len(d.Connectors()))
	return nil
}

// saveAndReopen writes w.doc to tmp/<name> and reopens it into w.reloaded.
func (w *world) saveAndReopen(ctx context.Context, name string) error {
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	out := filepath.Join(tmpDir, name)
	if err := w.doc.WriteFile(out); err != nil {
		return fmt.Errorf("write %s: %w", out, err)
	}
	d, err := eaxmi.Open(out)
	if err != nil {
		return fmt.Errorf("reopen %s: %w", out, err)
	}
	w.reloaded = d
	w.savedPath = out
	logf(ctx, "wrote and reopened %s", out)
	return nil
}

// ---- table helpers ----

func tableData(t *godog.Table) (header []string, rows [][]string) {
	for i, r := range t.Rows {
		cells := make([]string, len(r.Cells))
		for j, c := range r.Cells {
			cells[j] = c.Value
		}
		if i == 0 {
			header = cells
			continue
		}
		rows = append(rows, cells)
	}
	return header, rows
}

// checkFields compares a "| field | value |" table against a lookup function.
func checkFields(t *godog.Table, get func(field string) (string, bool)) error {
	_, rows := tableData(t)
	for _, r := range rows {
		field, want := r[0], r[1]
		got, ok := get(field)
		if !ok {
			return fmt.Errorf("unknown field %q", field)
		}
		if got != want {
			return fmt.Errorf("%s = %q, want %q", field, got, want)
		}
	}
	return nil
}

func atoi(s string) int { n, _ := strconv.Atoi(strings.TrimSpace(s)); return n }

func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func names(elems []*eaxmi.Element) []string {
	out := make([]string, len(elems))
	for i, e := range elems {
		out[i] = e.Name
	}
	return out
}

func pkgNames(pkgs []*eaxmi.Package) []string {
	out := make([]string, len(pkgs))
	for i, p := range pkgs {
		out[i] = p.Name
	}
	return out
}

func sameSet(got, want []string) error {
	g := map[string]int{}
	for _, v := range got {
		g[v]++
	}
	for _, v := range want {
		if g[v] == 0 {
			return fmt.Errorf("missing %q (have %v)", v, got)
		}
		g[v]--
	}
	var extra []string
	for v, n := range g {
		for ; n > 0; n-- {
			extra = append(extra, v)
		}
	}
	if len(extra) > 0 {
		return fmt.Errorf("unexpected %v (have %v, want %v)", extra, got, want)
	}
	return nil
}
