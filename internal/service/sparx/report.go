package sparx

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// AssembleReport merges several XMI files into one importable file: each
// source's root package becomes a top-level package in the output, so a
// reviewer imports one file instead of many. If the output turns out broken in
// EA, the individual sources say which one is at fault.
//
// Sources must be identity-disjoint (each built with a fresh identity — see
// SetRootName). The first source provides the shared wrapper.
func AssembleReport(outPath string, sourcePaths []string) error {
	if len(sourcePaths) == 0 {
		return fmt.Errorf("sparx: no sources for the report")
	}
	base, err := eaxmi.Open(sourcePaths[0])
	if err != nil {
		return fmt.Errorf("sparx: report base %q: %w", sourcePaths[0], err)
	}
	for _, p := range sourcePaths[1:] {
		next, err := eaxmi.Open(p)
		if err != nil {
			return fmt.Errorf("sparx: report source %q: %w", p, err)
		}
		if err := base.Merge(next); err != nil {
			return fmt.Errorf("sparx: merge %q: %w", p, err)
		}
	}
	return errWrap(base.WriteFile(outPath))
}
