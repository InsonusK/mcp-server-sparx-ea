package mcpserver_test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures runs the ea_query tool's Cucumber specs under features/.
func TestFeatures(t *testing.T) {
	format := "pretty"
	if f := os.Getenv("GODOG_FORMAT"); f != "" {
		format = f
	}

	suite := godog.TestSuite{
		Name:                "ea-query-tool",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   format,
			Paths:    []string{"features"},
			Tags:     "~@todo",
			Strict:   true,
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("cucumber scenarios failed")
	}
}
