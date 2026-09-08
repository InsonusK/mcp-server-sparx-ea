package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs every .feature file under features/connector and
// features/technical. godog reports each scenario as a Go subtest, so
// `go test -json` counts scenarios individually.
func TestFeatures(t *testing.T) {
	format := "pretty"
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		format = f
	}

	suite := godog.TestSuite{
		Name:                "sparx-ea-connector",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"connector", "technical"},
			Strict:   true,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("cucumber scenarios failed")
	}
}
