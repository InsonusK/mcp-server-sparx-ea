package sparx

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// ReportRootName is the single root package of an assembled report.
const ReportRootName = "Sparx Service Test Report"

// AssembleReport combines several XMI files into one importable file, built from
// scratch so it carries none of the sources' model-root flags:
//
//  1. start an empty model whose one root package is ReportRootName
//  2. for each source, deep-copy its root package (with all its sub-packages) in
//     as a child of the report root, regenerating every id the subtree owns so
//     nothing collides and cross-references inside the source stay intact
//     (eaxmi.CopyPackage — it also strips the copied package's model-root flags)
//
// EA's "Import Package from XMI" processes a single root, so the report has
// exactly one. If it turns out broken, the individual sources say which one.
func AssembleReport(outPath string, sourcePaths []string) error {
	if len(sourcePaths) == 0 {
		return fmt.Errorf("sparx: no sources for the report")
	}

	base, err := eaxmi.NewModel(ReportRootName)
	if err != nil {
		return fmt.Errorf("sparx: report shell: %w", err)
	}
	rootID := base.Root.Packages[0].XMIID

	// carry the ArchiMate profile definitions so EA recognises the stereotypes
	if first, err := eaxmi.Open(sourcePaths[0]); err == nil {
		_ = base.CopyProfileDefinitions(first)
	}

	var failed []string
	for _, p := range sourcePaths {
		src, err := eaxmi.Open(p)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
			continue
		}
		if len(src.Root.Packages) == 0 {
			failed = append(failed, p+": no root package")
			continue
		}
		if _, err := base.CopyPackage(src, src.Root.Packages[0].XMIID, rootID); err != nil {
			failed = append(failed, fmt.Sprintf("%s: %v", p, err))
		}
	}

	if err := base.WriteFile(outPath); err != nil {
		return errWrap(err)
	}
	if len(failed) > 0 {
		return fmt.Errorf("sparx: report written, but %d part(s) failed: %v", len(failed), failed)
	}
	return nil
}
