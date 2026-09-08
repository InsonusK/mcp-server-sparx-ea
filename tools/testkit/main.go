// Command testkit normalizes Go test / coverage / gremlins output into the
// stack-independent report contract used by the Makefile:
//
//	tmp/result/unit-test.json      {total,passed,failed}
//	tmp/result/coverage-test.json  {linePct}
//	tmp/result/mutation-test.json  {killed,survived,timedout,noCoverage,score}
//
// and assembles the publishable public/ directory (per-kind report copies plus
// shields.io endpoint badges and a copied landing page).
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: testkit <unit-test|coverage|mutation|report> [args]")
	}
	var err error
	switch os.Args[1] {
	case "unit-test":
		err = cmdUnitTest(os.Args[2:])
	case "coverage":
		err = cmdCoverage(os.Args[2:])
	case "mutation":
		err = cmdMutation(os.Args[2:])
	case "report":
		err = cmdReport(os.Args[2:])
	default:
		fail("unknown subcommand %q", os.Args[1])
	}
	if err != nil {
		fail("%v", err)
	}
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "testkit: "+format+"\n", a...)
	os.Exit(1)
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// ---------- unit-test ----------

type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

func cmdUnitTest(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: testkit unit-test <go-test-json>")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	// terminal[key] = last pass/fail/skip action seen for that test
	terminal := map[string]string{}
	names := map[string][]string{} // package -> test names
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var e goTestEvent
		if json.Unmarshal([]byte(line), &e) != nil || e.Test == "" {
			continue
		}
		key := e.Package + "\x00" + e.Test
		switch e.Action {
		case "pass", "fail", "skip":
			terminal[key] = e.Action
		}
		names[e.Package] = append(names[e.Package], e.Test)
	}
	if err := sc.Err(); err != nil {
		return err
	}

	isParent := func(pkg, test string) bool {
		for _, other := range names[pkg] {
			if strings.HasPrefix(other, test+"/") {
				return true
			}
		}
		return false
	}

	var passed, failed, skipped int
	for key, action := range terminal {
		parts := strings.SplitN(key, "\x00", 2)
		if isParent(parts[0], parts[1]) {
			continue // count only leaf tests / scenarios
		}
		switch action {
		case "pass":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}

	result := map[string]int{
		"total":  passed + failed + skipped,
		"passed": passed,
		"failed": failed,
	}
	if err := writeJSON("tmp/result/unit-test.json", result); err != nil {
		return err
	}
	summary := fmt.Sprintf("tests: %d total, %d passed, %d failed, %d skipped\n",
		passed+failed+skipped, passed, failed, skipped)
	if err := os.MkdirAll("tmp/report/tests", 0o755); err != nil {
		return err
	}
	fmt.Print(summary)
	return os.WriteFile("tmp/report/tests/summary.txt", []byte(summary), 0o644)
}

// ---------- coverage ----------

var coverTotalRe = regexp.MustCompile(`total:\s+\(statements\)\s+([0-9.]+)%`)

func cmdCoverage(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: testkit coverage <go-tool-cover-func-output>")
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	m := coverTotalRe.FindSubmatch(b)
	if m == nil {
		return fmt.Errorf("could not find a total coverage line in %s", args[0])
	}
	pct, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		return err
	}
	fmt.Printf("coverage: %.1f%%\n", pct)
	return writeJSON("tmp/result/coverage-test.json", map[string]float64{"linePct": pct})
}

// ---------- mutation ----------

type gremlinsReport struct {
	Files []struct {
		Mutations []struct {
			Status string `json:"status"`
		} `json:"mutations"`
	} `json:"files"`
}

func cmdMutation(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: testkit mutation <gremlins-json>...")
	}

	var killed, survived, timedout, noCoverage int
	for _, path := range args {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var rep gremlinsReport
		if err := json.Unmarshal(b, &rep); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for _, file := range rep.Files {
			for _, mut := range file.Mutations {
				switch strings.ToUpper(mut.Status) {
				case "KILLED":
					killed++
				case "LIVED":
					survived++
				case "TIMED OUT", "TIMEDOUT":
					timedout++
				case "NOT COVERED", "NOTCOVERED":
					noCoverage++
				}
			}
		}
	}

	denom := killed + survived + timedout + noCoverage
	score := 0.0
	if denom > 0 {
		score = math.Round(float64(killed)/float64(denom)*1000) / 10
	}
	fmt.Printf("mutation: killed=%d survived=%d timedout=%d noCoverage=%d score=%.1f\n",
		killed, survived, timedout, noCoverage, score)
	return writeJSON("tmp/result/mutation-test.json", map[string]any{
		"killed":     killed,
		"survived":   survived,
		"timedout":   timedout,
		"noCoverage": noCoverage,
		"score":      score,
	})
}

// ---------- report ----------

func cmdReport(_ []string) error {
	if err := os.RemoveAll("public"); err != nil {
		return err
	}
	if err := os.MkdirAll("public", 0o755); err != nil {
		return err
	}

	// per-kind native report copies
	for _, kind := range []string{"tests", "coverage", "mutation"} {
		src := filepath.Join("tmp", "report", kind)
		if info, err := os.Stat(src); err == nil && info.IsDir() {
			if err := copyTree(src, filepath.Join("public", kind)); err != nil {
				return err
			}
		}
	}

	// badges
	if u := readInts("tmp/result/unit-test.json"); u != nil {
		msg := fmt.Sprintf("%d passed", u["passed"])
		color := "brightgreen"
		if u["failed"] > 0 {
			msg = fmt.Sprintf("%d failed", u["failed"])
			color = "red"
		}
		if err := writeBadge("public/tests-badge.json", "tests", msg, color); err != nil {
			return err
		}
	}
	if c := readFloat("tmp/result/coverage-test.json", "linePct"); c != nil {
		if err := writeBadge("public/coverage-badge.json", "coverage",
			fmt.Sprintf("%.1f%%", *c), pctColor(*c)); err != nil {
			return err
		}
	}
	if s := readFloat("tmp/result/mutation-test.json", "score"); s != nil {
		if err := writeBadge("public/mutation-badge.json", "mutation score",
			fmt.Sprintf("%.1f%%", *s), pctColor(*s)); err != nil {
			return err
		}
	}

	// landing page: copied verbatim, never generated
	page, err := os.ReadFile("report-template/index.html")
	if err != nil {
		return fmt.Errorf("report-template/index.html is required: %w", err)
	}
	return os.WriteFile("public/index.html", page, 0o644)
}

func pctColor(v float64) string {
	switch {
	case v >= 80:
		return "brightgreen"
	case v >= 60:
		return "yellowgreen"
	default:
		return "red"
	}
}

func writeBadge(path, label, message, color string) error {
	return writeJSON(path, map[string]any{
		"schemaVersion": 1,
		"label":         label,
		"message":       message,
		"color":         color,
	})
}

func readInts(path string) map[string]int {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var m map[string]int
	if json.Unmarshal(b, &m) != nil {
		return nil
	}
	return m
}

func readFloat(path, key string) *float64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(b, &m) != nil {
		return nil
	}
	raw, ok := m[key]
	if !ok {
		return nil
	}
	var f float64
	if json.Unmarshal(raw, &f) != nil {
		return nil
	}
	return &f
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
