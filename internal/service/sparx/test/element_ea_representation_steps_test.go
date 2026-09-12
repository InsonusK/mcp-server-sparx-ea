package sparxtest

// The EA representation behind each ArchiMate element type
// (element_ea_representation.feature): EA only renders an element's real
// ArchiMate shape when its stereotype is applied to the UML base type EA's
// own ArchiMate3 profile expects for it — uml:Class for most types, but
// uml:Interface / uml:Component / uml:Activity for a few. Get the base type
// wrong and EA renders the element as an anonymous, unstyled class instead of
// (say) a Technology Interface. See tmp/examples/sandbox.xml for the real EA
// export these steps guard against regressing.

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx/test/common"
)

func registerElementEAReprSteps(sc *godog.ScenarioContext, w *common.World) {
	sc.Step(`^I create one element of every ArchiMate type in "([^"]*)"$`,
		func(ctx context.Context, pkg string) error {
			types := sparx.ElementTypes()
			for _, qualified := range types {
				bare := strings.TrimPrefix(qualified, "ArchiMate.")
				if _, err := w.Mut.CreateElement(pkg, qualified, bare, ""); err != nil {
					w.Err = fmt.Errorf("create %s: %w", qualified, err)
					return nil
				}
			}
			w.Err = nil
			common.Logf(ctx, "created %d elements (one per ArchiMate type) in %q", len(types), pkg)
			return w.Save(ctx)
		})

	sc.Step(`^after reload the element "([^"]*)" has EA type "([^"]*)"$`,
		func(ctx context.Context, ref, wantUMLType string) error {
			doc, err := w.ReloadedDoc(ctx)
			if err != nil {
				return err
			}
			el := doc.ResolveElement(ref)
			if el == nil {
				return fmt.Errorf("no element for %q after reload", ref)
			}
			if el.UMLType != wantUMLType {
				return fmt.Errorf("%s: EA base type is %q, want %q", ref, el.UMLType, wantUMLType)
			}
			common.Logf(ctx, "%s is EA type %s, as expected", ref, el.UMLType)
			return nil
		})

	sc.Step(`^after reload every element in "([^"]*)" has EA's correct UML base type$`,
		func(ctx context.Context, pkg string) error {
			doc, err := w.ReloadedDoc(ctx)
			if err != nil {
				return err
			}
			types := sparx.ElementTypes()
			for _, qualified := range types {
				bare := strings.TrimPrefix(qualified, "ArchiMate.")
				ref := pkg + "/" + bare
				el := doc.ResolveElement(ref)
				if el == nil {
					return fmt.Errorf("no element for %q after reload", ref)
				}
				wantUML, wantStereo := sparx.EAElementForType(bare)
				if el.UMLType != wantUML {
					return fmt.Errorf("%s: EA base type is %q, want %q", ref, el.UMLType, wantUML)
				}
				if el.Stereotype != wantStereo {
					return fmt.Errorf("%s: stereotype is %q, want %q", ref, el.Stereotype, wantStereo)
				}
			}
			common.Logf(ctx, "all %d ArchiMate element types map to EA's correct UML base type", len(types))
			return nil
		})
}
