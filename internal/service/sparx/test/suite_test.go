// Package sparxtest is the black-box Cucumber suite for internal/service/sparx.
//
// Layout (docs/skills/cucumber-go-testing.md):
//   - features/            one .feature per operation
//   - test/common/         the World, plumbing steps, generic data-table comparators
//   - test/*_steps_test.go  the per-operation action steps ("I create", "I relate", …)
package sparxtest

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func TestFeatures(t *testing.T) {
	// tmp/ holds the saved working copies: cleared at the start of every run,
	// kept afterwards for import into Sparx EA (tmp/scenario/*.xml per scenario,
	// tmp/report.xml all of them merged).
	_ = os.RemoveAll(common.TmpDir)

	opts := godog.Options{
		Format:   "pretty",
		Paths:    []string{"../features"},
		Tags:     "~@todo",
		Strict:   true,
		TestingT: t,
	}
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		opts.Format = f
	}
	if os.Getenv("GODOG_STEPS") != "" {
		opts.ShowStepDefinitions = true
	}

	suite := godog.TestSuite{
		Name:                "sparx-service",
		ScenarioInitializer: initializeScenario,
		Options:             &opts,
	}
	if suite.Run() != 0 {
		t.Fatal("cucumber scenarios failed")
	}

	if err := assembleReport(t); err != nil {
		t.Fatalf("assemble tmp/report.xml: %v", err)
	}
}

// assembleReport merges every scenario file (except the root-rename ones, which
// deliberately keep the fixture identity and would collide) into one importable
// tmp/report.xml — each scenario a top-level package.
func assembleReport(t *testing.T) error {
	files, err := filepath.Glob(filepath.Join(common.ScenarioDir, "*.xml"))
	if err != nil || len(files) == 0 {
		return err
	}
	sort.Strings(files)
	var srcs []string
	for _, f := range files {
		if strings.HasPrefix(filepath.Base(f), "root_rename") {
			continue
		}
		srcs = append(srcs, f)
	}
	out := filepath.Join(common.TmpDir, "report.xml")
	if err := sparx.AssembleReport(out, srcs); err != nil {
		return err
	}
	t.Logf("merged %d scenario file(s) → %s", len(srcs), out)
	return nil
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := common.NewWorld()

	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		w.ScenarioName = s.Name
		w.Reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.Reset()
		return ctx, nil
	})

	common.RegisterSharedSteps(sc, w)
	registerTreeSteps(sc, w)
	registerPackageSteps(sc, w)
	registerElementSteps(sc, w)
	registerRelationshipSteps(sc, w)
	registerDiagramSteps(sc, w)
	registerVocabSteps(sc, w)
}
