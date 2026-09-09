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
	"testing"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func TestFeatures(t *testing.T) {
	// tmp/ holds the saved working copies from methods 4-6 scenarios: cleared
	// at the start of every run, kept afterwards for import into Sparx EA. One
	// file per feature, accumulating that feature's changes.
	_ = os.RemoveAll(common.TmpDir)
	common.ResetWorkingModels()

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
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := common.NewWorld()

	sc.Before(func(ctx context.Context, s *godog.Scenario) (context.Context, error) {
		w.ScenarioName = s.Name
		w.Reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if err == nil {
			_ = w.Finish(ctx) // merge this scenario's package into its output file
		}
		w.Reset()
		return ctx, nil
	})

	common.RegisterSharedSteps(sc, w)
	registerTreeSteps(sc, w)
	registerElementSteps(sc, w)
	registerRelationshipSteps(sc, w)
	registerDiagramSteps(sc, w)
	registerVocabSteps(sc, w)
}
