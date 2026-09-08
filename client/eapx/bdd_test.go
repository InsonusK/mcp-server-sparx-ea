package eapx_test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs the connector's Cucumber specs under features/. godog
// reports each scenario as a Go subtest, so `go test -json` counts them
// individually alongside the plain tests in this package.
func TestFeatures(t *testing.T) {
	format := "pretty"
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		format = f
	}

	suite := godog.TestSuite{
		Name:                "eapx-connector",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"features"},
			Strict:   true,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("cucumber scenarios failed")
	}
}
