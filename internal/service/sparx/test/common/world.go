// Package common holds the shared parts of the sparx-service Cucumber suite:
// the scenario World, the plumbing steps (load a model, load a working copy,
// save, reload) and the generic data-table comparators. The per-operation
// "I create / I rename / I relate / …" steps live next to their feature file
// in package sparxtest.
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

// TmpDir is where methods 4-6 scenarios save their working copy — gitignored,
// cleared at the start of a run, kept afterwards for manual import into Sparx.
const TmpDir = "tmp"

// Each scenario builds a fully independent working copy of the fixture (its own
// root package, fresh GUIDs) and saves it as one "part" under tmp/.parts/<feature>/.
// Every save also rebuilds tmp/<feature>.xml by merging all of that feature's
// parts, so the file the user imports has one package per scenario and the
// scenarios never influence each other.
var scenarioCount int

// ResetWorkingModels resets the per-run scenario counter. TestFeatures calls it
// once per run (after wiping TmpDir).
func ResetWorkingModels() { scenarioCount = 0 }

// World is one scenario's state.
type World struct {
	ScenarioName string

	// read-only model (the "Given the model file" step)
	Svc     *sparx.Service
	Tree    *sparx.Node
	Element *sparx.ElementInfo
	Diagram *sparx.DiagramInfo
	Err     error

	// working copy (the "Given the working model" step)
	Mut      *sparx.Service
	Source   string
	Output   string // tmp file this scenario's package is merged into
	RootName string // this scenario's root package name (unique within Output)
	Seq      int    // this scenario's ordinal (names the part file)

	LastRelID      string
	LastElemPath   string
	LastDiagramRef string
}

// Rel turns a path written relative to the scenario's package ("Motivation_Package/GoalA")
// into one this scenario's model resolves.
func (w *World) Rel(path string) string {
	if w.RootName == "" || strings.HasPrefix(path, w.RootName+"/") || eaxmiIsID(path) {
		return path
	}
	return w.RootName + "/" + path
}

func eaxmiIsID(ref string) bool {
	return strings.HasPrefix(ref, "EAID_") || strings.HasPrefix(ref, "EAPK_") ||
		(strings.HasPrefix(ref, "{") && strings.HasSuffix(ref, "}"))
}

func NewWorld() *World { return &World{} }

// Active is the model a read step operates on: the in-memory working copy if
// one is loaded, otherwise the read-only fixture.
func (w *World) Active() *sparx.Service {
	if w.Mut != nil {
		return w.Mut
	}
	return w.Svc
}

func (w *World) Reset() {
	name := w.ScenarioName
	*w = World{ScenarioName: name}
}

func Logf(_ context.Context, format string, a ...any) {
	fmt.Printf("      · "+format+"\n", a...)
}

// FixturePath resolves a testdata fixture name.
func FixturePath(name string) string { return filepath.Join("testdata", name) }

// ---------- plumbing steps ----------

func (w *World) theModelFile(ctx context.Context, name string) error {
	svc, err := sparx.Open(FixturePath(name))
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	w.Svc = svc
	Logf(ctx, "loaded read-only model %q", name)
	return nil
}

// theWorkingModel loads a fresh, independent working copy of a fixture for this
// scenario: its own root package (named after the scenario, unique within the
// output file), fresh GUIDs. Edits never touch the fixture. At scenario end the
// package is merged into tmp/<output> alongside the other scenarios'.
//
//	Given the working model:
//	  | source   | TestProject.xml         |
//	  | output   | relationship_create.xml |
//	  | identity | fresh                   |  # optional: "fresh" (default) | "keep"
func (w *World) theWorkingModel(ctx context.Context, table *godog.Table) error {
	cfg := map[string]string{}
	for _, r := range table.Rows {
		if len(r.Cells) == 2 {
			cfg[r.Cells[0].Value] = r.Cells[1].Value
		}
	}
	src, out := cfg["source"], cfg["output"]
	if src == "" || out == "" {
		return fmt.Errorf(`the working model needs "source" and "output" rows`)
	}
	svc, err := sparx.Open(FixturePath(src))
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	scenarioCount++
	w.Seq = scenarioCount
	root := fmt.Sprintf("%02d %s", w.Seq, w.ScenarioName)
	newRoot, err := svc.SetRootName(root, cfg["identity"] != "keep")
	if err != nil {
		return err
	}
	w.Mut, w.Source, w.Output, w.RootName = svc, src, out, newRoot
	Logf(ctx, "working model of %q → package %q → merges into %s/%s (fixture untouched)",
		src, newRoot, TmpDir, out)
	return nil
}

func (w *World) outputPath() string { return filepath.Join(TmpDir, w.Output) }

func (w *World) partsDir() string {
	return filepath.Join(TmpDir, ".parts", strings.TrimSuffix(w.Output, ".xml"))
}

func (w *World) partPath() string {
	return filepath.Join(w.partsDir(), fmt.Sprintf("%02d.xml", w.Seq))
}

// Save persists this scenario's current working copy. Called after every
// successful mutation and from the After hook. It writes the scenario's own
// "part" file, then rebuilds tmp/<output> by merging every part of that feature
// — so the file always reflects the latest state of every scenario and no
// scenario can influence another.
func (w *World) Save(ctx context.Context) error { return w.persist(ctx) }

func (w *World) persist(ctx context.Context) error {
	if w.Mut == nil {
		return nil
	}
	if err := os.MkdirAll(w.partsDir(), 0o755); err != nil {
		return err
	}
	if err := w.Mut.Save(w.partPath()); err != nil {
		return fmt.Errorf("save part %s: %w", w.partPath(), err)
	}

	parts, err := filepath.Glob(filepath.Join(w.partsDir(), "*.xml"))
	if err != nil {
		return err
	}
	sort.Strings(parts)

	combined, err := sparx.Open(parts[0])
	if err != nil {
		return fmt.Errorf("open part %s: %w", parts[0], err)
	}
	for _, p := range parts[1:] {
		part, err := sparx.Open(p)
		if err != nil {
			return fmt.Errorf("open part %s: %w", p, err)
		}
		if err := combined.Absorb(part); err != nil {
			return fmt.Errorf("merge part %s into %s: %w", p, w.Output, err)
		}
	}
	if err := os.MkdirAll(TmpDir, 0o755); err != nil {
		return err
	}
	if err := combined.Save(w.outputPath()); err != nil {
		return fmt.Errorf("save %s: %w", w.Output, err)
	}
	Logf(ctx, "merged %d scenario package(s) into %s (open in Sparx / import to review)",
		len(parts), w.outputPath())
	return nil
}

// Finish persists this scenario's working copy if it has one (call from the
// After hook so every scenario contributes a package, even one with only a
// Background).
func (w *World) Finish(ctx context.Context) error { return w.persist(ctx) }

// Reloaded persists this scenario's working copy, then reopens the merged output
// file so "after reload" assertions run against a real XMI round-trip.
func (w *World) Reloaded(ctx context.Context) (*sparx.Service, error) {
	if err := w.persist(ctx); err != nil {
		return nil, err
	}
	return sparx.Open(w.outputPath())
}

// ---------- generic outcome steps ----------

func (w *World) succeeds(ctx context.Context) error {
	if w.Err != nil {
		return fmt.Errorf("expected success, got: %v", w.Err)
	}
	Logf(ctx, "operation succeeded")
	return nil
}

func (w *World) failsWith(ctx context.Context, want string) error {
	if w.Err == nil {
		return fmt.Errorf("expected a failure containing %q, but it succeeded", want)
	}
	if !strings.Contains(w.Err.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.Err.Error(), want)
	}
	Logf(ctx, "failed as expected: %v", w.Err)
	return nil
}

func (w *World) outcome(ctx context.Context, s string) error {
	if s == "succeeds" {
		return w.succeeds(ctx)
	}
	if rest, ok := strings.CutPrefix(s, "fails with "); ok {
		return w.failsWith(ctx, strings.Trim(rest, `"`))
	}
	return fmt.Errorf("unknown outcome %q", s)
}

// ---------- generic "the element / diagram is …" comparators ----------

func (w *World) elem() *sparx.ElementInfo {
	if w.Element != nil {
		return w.Element
	}
	return nil
}

func (w *World) elementIs(ctx context.Context, table *godog.Table) error {
	return fieldTable(w.elem(), table, "element")
}

func (w *World) elementFieldIs(ctx context.Context, field, want string) error {
	rows, err := ToRows([]any{w.elem()})
	if err != nil {
		return err
	}
	if rows[0][field] != want {
		return fmt.Errorf("element %s = %q, want %q", field, rows[0][field], want)
	}
	return nil
}

func (w *World) elementRelations(ctx context.Context, exact bool, table *godog.Table) error {
	if w.Element == nil {
		return fmt.Errorf("no element read (err: %v)", w.Err)
	}
	rows, err := ToRows(anySlice(w.Element.Relations))
	if err != nil {
		return err
	}
	Logf(ctx, "element has %d relation(s)", len(rows))
	return MatchTable(rows, table, exact)
}

func (w *World) diagramFieldIs(ctx context.Context, field, want string) error {
	if w.Diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.Err)
	}
	rows, _ := ToRows([]any{w.Diagram})
	if rows[0][field] != want {
		return fmt.Errorf("diagram %s = %q, want %q", field, rows[0][field], want)
	}
	return nil
}

func (w *World) diagramHasNObjects(ctx context.Context, n int) error {
	if w.Diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.Err)
	}
	if got := len(w.Diagram.Objects); got != n {
		return fmt.Errorf("diagram has %d placed elements, want %d", got, n)
	}
	return nil
}

func (w *World) diagramHasNLinks(ctx context.Context, n int) error {
	if w.Diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.Err)
	}
	if got := len(w.Diagram.Links); got != n {
		return fmt.Errorf("diagram has %d links, want %d", got, n)
	}
	return nil
}

func (w *World) placedElementsInclude(ctx context.Context, table *godog.Table) error {
	if w.Diagram == nil {
		return fmt.Errorf("no diagram read (err: %v)", w.Err)
	}
	rows, err := ToRows(anySlice(w.Diagram.Objects))
	if err != nil {
		return err
	}
	return MatchTable(rows, table, false)
}

// ---------- "after reload …" comparators ----------

func (w *World) reloadElementIs(ctx context.Context, ref string, table *godog.Table) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.Rel(ref))
	if err != nil {
		return err
	}
	return fieldTable(el, table, "reloaded element")
}

func (w *World) reloadElementFieldIs(ctx context.Context, field, want string) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.Rel(w.LastElemPath))
	if err != nil {
		return err
	}
	rows, _ := ToRows([]any{el})
	if rows[0][field] != want {
		return fmt.Errorf("after reload %s = %q, want %q", field, rows[0][field], want)
	}
	return nil
}

func (w *World) reloadElementGUIDIsNot(ctx context.Context, ref, notWant string) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.Rel(ref))
	if err != nil {
		return err
	}
	if el.GUID == notWant {
		return fmt.Errorf("after reload %s still has guid %s, expected a fresh one", ref, notWant)
	}
	Logf(ctx, "reloaded %q has a fresh guid %s (was %s)", ref, el.GUID, notWant)
	return nil
}

func (w *World) reloadElementNotFound(ctx context.Context, ref string) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Element(w.Rel(ref)); err == nil {
		return fmt.Errorf("element %q still resolves after reload", ref)
	}
	Logf(ctx, "element %q is gone after reload", ref)
	return nil
}

func (w *World) reloadElementNoRelationTo(ctx context.Context, ref, other string) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.Rel(ref))
	if err != nil {
		return err
	}
	for _, r := range el.Relations {
		if r.OtherName == other {
			return fmt.Errorf("%s still has a %s relation to %s after reload", el.Name, r.Type, other)
		}
	}
	Logf(ctx, "%q has no relation to %q after reload", ref, other)
	return nil
}

func (w *World) reloadElementRelationsInclude(ctx context.Context, ref string, table *godog.Table) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	el, err := svc.Element(w.Rel(ref))
	if err != nil {
		return err
	}
	rows, err := ToRows(anySlice(el.Relations))
	if err != nil {
		return err
	}
	return MatchTable(rows, table, false)
}

func (w *World) reloadDiagramHasNObjects(ctx context.Context, ref string, n int) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	d, err := svc.Diagram(w.Rel(ref))
	if err != nil {
		return err
	}
	if len(d.Objects) != n {
		return fmt.Errorf("after reload diagram has %d placed elements, want %d", len(d.Objects), n)
	}
	return nil
}

func (w *World) reloadDiagramObjectsInclude(ctx context.Context, table *godog.Table) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	d, err := svc.Diagram(w.Rel(w.LastDiagramRef))
	if err != nil {
		return err
	}
	rows, err := ToRows(anySlice(d.Objects))
	if err != nil {
		return err
	}
	return MatchTable(rows, table, false)
}

func (w *World) reloadRootIs(ctx context.Context, want string) error {
	svc, err := w.Reloaded(ctx)
	if err != nil {
		return err
	}
	roots := svc.RootPackages()
	for _, r := range roots {
		if r == want {
			Logf(ctx, "reloaded root package is %q", want)
			return nil
		}
	}
	return fmt.Errorf("after reload root packages are %v, want %q", roots, want)
}

// reloadRootIsScenario checks the merged output file carries this scenario's own
// root package (the harness names it "NN <scenario name>").
func (w *World) reloadRootIsScenario(ctx context.Context) error {
	return w.reloadRootIs(ctx, w.RootName)
}

// ---------- registration ----------

// RegisterSharedSteps wires every plumbing / generic-comparator step.
func RegisterSharedSteps(sc *godog.ScenarioContext, w *World) {
	sc.Step(`^the model file "([^"]*)"$`, w.theModelFile)
	sc.Step(`^the working model:$`, w.theWorkingModel)

	sc.Step(`^the (?:read|create|delete|placement|rename) succeeds$`, w.succeeds)
	sc.Step(`^the (?:read|create|delete|placement|rename) fails with "([^"]*)"$`, w.failsWith)
	sc.Step(`^the relate (succeeds|fails with .+)$`, w.outcome)

	sc.Step(`^the element is:$`, w.elementIs)
	sc.Step(`^the element field "([^"]*)" is "([^"]*)"$`, w.elementFieldIs)
	sc.Step(`^the element relations are exactly:$`, func(ctx context.Context, t *godog.Table) error {
		return w.elementRelations(ctx, true, t)
	})
	sc.Step(`^the element relations include:$`, func(ctx context.Context, t *godog.Table) error {
		return w.elementRelations(ctx, false, t)
	})

	sc.Step(`^the diagram field "([^"]*)" is "([^"]*)"$`, w.diagramFieldIs)
	sc.Step(`^the diagram has (\d+) placed elements$`, w.diagramHasNObjects)
	sc.Step(`^the diagram has (\d+) links$`, w.diagramHasNLinks)
	sc.Step(`^the placed elements include:$`, w.placedElementsInclude)

	sc.Step(`^after reload the element "([^"]*)" is:$`, w.reloadElementIs)
	sc.Step(`^after reload the element field "([^"]*)" is "([^"]*)"$`, w.reloadElementFieldIs)
	sc.Step(`^after reload the element "([^"]*)" cannot be found$`, w.reloadElementNotFound)
	sc.Step(`^after reload the element "([^"]*)" guid is not "([^"]*)"$`, w.reloadElementGUIDIsNot)
	sc.Step(`^after reload the element "([^"]*)" has no relation to "([^"]*)"$`, w.reloadElementNoRelationTo)
	sc.Step(`^after reload the element "([^"]*)" relations include:$`, w.reloadElementRelationsInclude)
	sc.Step(`^after reload the diagram "([^"]*)" has (\d+) placed elements$`, w.reloadDiagramHasNObjects)
	sc.Step(`^after reload the diagram placed elements include:$`, w.reloadDiagramObjectsInclude)
	sc.Step(`^after reload the root package is "([^"]*)"$`, w.reloadRootIs)
	sc.Step(`^after reload the root package is this scenario's package$`, w.reloadRootIsScenario)
}

// ---------- table helpers ----------

func fieldTable(v any, table *godog.Table, label string) error {
	if v == nil {
		return fmt.Errorf("no %s available", label)
	}
	rows, err := ToRows([]any{v})
	if err != nil {
		return err
	}
	for _, r := range table.Rows[1:] { // skip "| field | value |"
		f, want := r.Cells[0].Value, r.Cells[1].Value
		if rows[0][f] != want {
			return fmt.Errorf("%s %s = %q, want %q", label, f, rows[0][f], want)
		}
	}
	return nil
}

// ToRows flattens a slice of structs to []map[string]string via JSON so a
// data-table step compares only the columns it names.
func ToRows(list []any) ([]map[string]string, error) {
	out := make([]map[string]string, 0, len(list))
	for _, v := range list {
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		row := map[string]string{}
		for k, val := range m {
			row[k] = stringify(val)
		}
		out = append(out, row)
	}
	return out, nil
}

func anySlice[T any](in []T) []any {
	out := make([]any, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// MatchTable compares actual rows against a data table. exact => bijection
// (order-independent). !exact => every table row must match some distinct
// actual row.
func MatchTable(actual []map[string]string, t *godog.Table, exact bool) error {
	var header []string
	var want [][]string
	for i, r := range t.Rows {
		cells := make([]string, len(r.Cells))
		for j, c := range r.Cells {
			cells[j] = c.Value
		}
		if i == 0 {
			header = cells
			continue
		}
		want = append(want, cells)
	}
	used := make([]bool, len(actual))
	for _, wrow := range want {
		found := -1
		for ai, arow := range actual {
			if used[ai] {
				continue
			}
			if rowMatch(header, wrow, arow) {
				found = ai
				break
			}
		}
		if found < 0 {
			return fmt.Errorf("no row matches %v (columns %v)\n  actual:\n%s", wrow, header, dump(actual))
		}
		used[found] = true
	}
	if exact {
		for ai, u := range used {
			if !u {
				return fmt.Errorf("unexpected extra row: %v\n  actual:\n%s", actual[ai], dump(actual))
			}
		}
	}
	return nil
}

func rowMatch(header, want []string, actual map[string]string) bool {
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

func dump(rows []map[string]string) string {
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "    %v\n", r)
	}
	return b.String()
}

// SplitList splits "a, b, c".
func SplitList(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

var _ = regexp.MustCompile
