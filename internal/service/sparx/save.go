package sparx

import "fmt"

// Save writes the edited model to path. It refuses to overwrite the file the
// service was opened from — a working copy is always a separate file the user
// imports back into EA.
func (s *Service) Save(path string) error {
	if path == s.doc.Path {
		return fmt.Errorf("sparx: refusing to overwrite the source file %q", path)
	}
	return errWrap(s.doc.WriteFile(path))
}
