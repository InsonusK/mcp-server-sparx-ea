package features

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs the project-wide architectural specs in this directory.
func TestFeatures(t *testing.T) {
	format := "pretty"
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		format = f
	}

	suite := godog.TestSuite{
		Name:                "project-architecture",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"."},
			Tags:     "~@todo",
			Strict:   true,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("cucumber scenarios failed")
	}
}
