package sparxtest

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
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
	w := newWorld()

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	registerModelSteps(sc, w)
	registerElementSteps(sc, w)
	registerDiagramSteps(sc, w)
	registerVocabSteps(sc, w)
}
