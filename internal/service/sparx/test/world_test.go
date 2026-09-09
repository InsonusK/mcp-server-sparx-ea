// Package sparxtest is the black-box Cucumber suite for internal/service/sparx.
// Conventions: docs/skills/cucumber-go-testing.md.
package sparxtest

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

type world struct {
	svc *sparx.Service

	tree    *sparx.Node
	element *sparx.ElementInfo
	diagram *sparx.DiagramInfo
	lastErr error
}

func newWorld() *world { return &world{} }

func (w *world) reset() { *w = world{} }

func logf(_ context.Context, format string, a ...any) {
	fmt.Printf("      · "+format+"\n", a...)
}

func fixturePath(name string) string { return filepath.Join("testdata", name) }

// ---- data-table comparison (generic comparator) ----

// toRows flattens a slice of structs to []map[string]string via JSON so a
// data-table step can compare only the columns it names.
func toRows(v any) ([]map[string]string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var list []map[string]any
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	rows := make([]map[string]string, len(list))
	for i, m := range list {
		rows[i] = map[string]string{}
		for k, val := range m {
			rows[i][k] = stringify(val)
		}
	}
	return rows, nil
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func tableHeaderRows(t *godog.Table) (header []string, rows [][]string) {
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
	return
}

// rowMatches reports whether actual satisfies every named (non-empty) cell of want.
func rowMatches(header []string, want []string, actual map[string]string) bool {
	for i, col := range header {
		if want[i] == "" {
			continue
		}
		if actual[col] != want[i] {
			return false
		}
	}
	return true
}

// matchTable compares actual rows against a data table. exact => every actual
// row must be matched by exactly one table row and vice-versa (order-independent,
// count-aware). subset => every table row must match some distinct actual row.
func matchTable(actual []map[string]string, t *godog.Table, exact bool) error {
	header, wantRows := tableHeaderRows(t)
	usedActual := make([]bool, len(actual))

	for _, want := range wantRows {
		found := -1
		for ai, act := range actual {
			if usedActual[ai] {
				continue
			}
			if rowMatches(header, want, act) {
				found = ai
				break
			}
		}
		if found < 0 {
			return fmt.Errorf("no row matches %v\n  columns checked: %v\n  actual rows:\n%s",
				want, header, dumpRows(actual))
		}
		usedActual[found] = true
	}
	if exact {
		for ai, used := range usedActual {
			if !used {
				return fmt.Errorf("unexpected extra row: %v\n  actual rows:\n%s", actual[ai], dumpRows(actual))
			}
		}
	}
	return nil
}

func dumpRows(rows []map[string]string) string {
	var b strings.Builder
	for _, r := range rows {
		b.WriteString("    ")
		b.WriteString(fmt.Sprintf("%v", r))
		b.WriteByte('\n')
	}
	return b.String()
}
