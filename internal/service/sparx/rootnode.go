package sparx

import (
	"fmt"
	"strings"
)

// The EA root package (the XMI export's top node). RootNodeNameChange: renaming
// it — optionally with a fresh identity — lets many working copies import into
// one Sparx EA project side by side.

// RootPackages returns the names of the EA root packages in this model (usually
// one).
func (s *Service) RootPackages() []string { return s.doc.RootPackages() }

// SetRootName renames the single EA root package. With freshIdentity=true it
// also regenerates every GUID in the model, so a Saved copy imports into EA as
// a fully independent package that can sit next to the original (an XMI import
// matches by GUID). Returns the new root name.
func (s *Service) SetRootName(newName string, freshIdentity bool) (string, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return "", fmt.Errorf("sparx: root name is required")
	}
	if err := s.doc.RenameRoot(newName, freshIdentity); err != nil {
		return "", fmt.Errorf("sparx: %w", err)
	}
	return newName, nil
}
