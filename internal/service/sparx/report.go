package sparx

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// ReportRootName is the single root package of an assembled report.
const ReportRootName = "Sparx Service Test Report"

// AssembleReport merges several XMI files into one importable file with a single
// root package (EA's XMI import processes one root): each source's root package
// becomes a child of ReportRootName, keeping its own sub-packages. If the output
// turns out broken in EA, the individual sources say which one is at fault.
//
// Sources must be identity-disjoint (each built with a fresh identity — see
// SetRootName). The first source provides the shared wrapper (profile
// definitions, primitive types).
func AssembleReport(outPath string, sourcePaths []string) error {
	if len(sourcePaths) == 0 {
		return fmt.Errorf("sparx: no sources for the report")
	}

	base, err := eaxmi.Open(sourcePaths[0])
	if err != nil {
		return fmt.Errorf("sparx: report base %q: %w", sourcePaths[0], err)
	}
	// Turn the base into an empty shell: one root package, no content.
	roots := base.RootPackages()
	if len(roots) == 0 {
		return fmt.Errorf("sparx: report base has no root package")
	}
	rootID := base.Root.Packages[0].XMIID
	if err := base.SetPackageName(rootID, ReportRootName); err != nil {
		return fmt.Errorf("sparx: report root: %w", err)
	}
	for _, child := range append([]*eaxmi.Package(nil), base.Root.Packages[0].Packages...) {
		if _, err := base.RemovePackage(child.XMIID); err != nil {
			return fmt.Errorf("sparx: clearing report shell: %w", err)
		}
	}

	var failed []string
	for _, p := range sourcePaths {
		src, err := eaxmi.Open(p)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		if err := base.MergeUnder(src, rootID); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
	}
	if err := base.WriteFile(outPath); err != nil {
		return errWrap(err)
	}
	if len(failed) > 0 {
		return fmt.Errorf("sparx: report written, but %d source(s) could not be merged: %v", len(failed), failed)
	}
	return nil
}
