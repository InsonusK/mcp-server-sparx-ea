package sparxtest

// The ArchiMate type vocabulary itself (type_vocabulary.feature).

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func registerVocabSteps(sc *godog.ScenarioContext, _ *common.World) {
	sc.Step(`^"([^"]*)" is a known element type$`, func(ctx context.Context, t string) error {
		if !sparx.IsKnownElementType(t) {
			return fmt.Errorf("%q should be a known element type but is not", t)
		}
		common.Logf(ctx, "%q is a known element type", t)
		return nil
	})
	sc.Step(`^"([^"]*)" is not a known element type$`, func(ctx context.Context, t string) error {
		if sparx.IsKnownElementType(t) {
			return fmt.Errorf("%q should NOT be a known element type but is", t)
		}
		common.Logf(ctx, "%q is correctly rejected", t)
		return nil
	})
	sc.Step(`^"([^"]*)" is a known relationship type$`, func(ctx context.Context, t string) error {
		if !sparx.IsKnownRelationshipType(t) {
			return fmt.Errorf("%q should be a known relationship type but is not", t)
		}
		return nil
	})

	sc.Step(`^"([^"]*)" from "([^"]*)" to "([^"]*)" is "(allow|warn|deny)"$`,
		func(ctx context.Context, rel, src, tgt, want string) error {
			got := sparx.RelationshipVerdictName(rel, src, tgt)
			if got != want {
				return fmt.Errorf("%s %s -> %s: verdict %q, want %q", rel, src, tgt, got, want)
			}
			common.Logf(ctx, "%s %s -> %s: %s", rel, src, tgt, want)
			return nil
		})
}
