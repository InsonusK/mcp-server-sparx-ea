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
	rc := suite.Run()

	// Always assemble the report — even from a partial / failed run. If a
	// scenario broke its file, the report is where you notice, and
	// tmp/scenario/*.xml says which one.
	if err := assembleReport(t); err != nil {
		t.Errorf("assemble tmp/report.xml: %v", err)
	}

	if rc != 0 {
		t.Fatal("cucumber scenarios failed")
	}
}

// assembleReport merges every file under tmp/scenario/ into one importable
// tmp/report.xml — one sub-package per scenario. Scenarios that opt out
// ("| report | no |") save straight to tmp/ and are not here.
func assembleReport(t *testing.T) error {
	srcs, err := filepath.Glob(filepath.Join(common.ScenarioDir, "*.xml"))
	if err != nil {
		return err
	}
	sort.Strings(srcs)
	if len(srcs) == 0 {
		t.Logf("no scenario files in %s — skipping tmp/report.xml", common.ScenarioDir)
		return nil
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
	registerElementEAReprSteps(sc, w)
	registerRelationshipSteps(sc, w)
	registerDiagramSteps(sc, w)
	registerVocabSteps(sc, w)
}
