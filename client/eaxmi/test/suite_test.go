package eaxmitest

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs every scenario under ../features. Each scenario is a Go
// subtest.
//
//	GODOG_FORMAT=progress   change the console format
//	GODOG_STEPS=1           print the step-definition catalogue and exit
func TestFeatures(t *testing.T) {
	_ = os.RemoveAll(tmpDir)

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
		Name:                "eaxmi-codec",
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

	registerParseSteps(sc, w)
	registerNavigateSteps(sc, w)
	registerNewModelSteps(sc, w)
	registerWriteSteps(sc, w)
	registerIdentitySteps(sc, w)
	registerCopySteps(sc, w)
}
