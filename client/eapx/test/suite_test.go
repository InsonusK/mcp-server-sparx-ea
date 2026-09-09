package eapxtest

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs every scenario under ../features. Each scenario is a Go
// subtest, so `go test -json` counts them individually.
//
//	GODOG_FORMAT=progress   change the console format
//	GODOG_STEPS=1           print the step-definition catalogue and exit
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
		Name:                "eapx-connector",
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

	registerConnectionSteps(sc, w)
	registerQuerySteps(sc, w)
	registerResultSteps(sc, w)
	registerEncodingSteps(sc, w)
	registerConcurrencySteps(sc, w)
	registerArchitectureSteps(sc, w)
}
