package sparxtest

// Steps for features/type_mapping.feature (the ArchiMate vocabulary itself).

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

func isAKnownElementType(ctx context.Context, t string) error {
	if !sparx.IsKnownElementType(t) {
		return fmt.Errorf("%q should be a known element type but is not", t)
	}
	logf(ctx, "%q is a known element type", t)
	return nil
}

func isNotAKnownElementType(ctx context.Context, t string) error {
	if sparx.IsKnownElementType(t) {
		return fmt.Errorf("%q should NOT be a known element type but is", t)
	}
	logf(ctx, "%q is correctly rejected as an element type", t)
	return nil
}

func registerVocabSteps(sc *godog.ScenarioContext, _ *world) {
	sc.Step(`^"([^"]*)" is a known element type$`, isAKnownElementType)
	sc.Step(`^"([^"]*)" is not a known element type$`, isNotAKnownElementType)
}
