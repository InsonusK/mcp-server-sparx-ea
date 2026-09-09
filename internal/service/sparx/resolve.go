package sparx

import "github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"

// Thin wrappers over eaxmi navigation (client/eaxmi/navigate.go). A ref is an
// xmi:id, a braced GUID, or a slash path from an EA root package.

func looksLikeID(ref string) bool { return eaxmi.IsRefID(ref) }

func splitPath(ref string) []string { return eaxmi.SplitPath(ref) }

func (s *Service) resolveElement(ref string) (*eaxmi.Element, string) {
	e := s.doc.ResolveElement(ref)
	if e == nil {
		return nil, ""
	}
	return e, s.doc.PathOfElement(e)
}

func (s *Service) resolveDiagram(ref string) (*eaxmi.Diagram, string) {
	g := s.doc.ResolveDiagram(ref)
	if g == nil {
		return nil, ""
	}
	return g, s.doc.PathOfDiagram(g)
}

func (s *Service) resolvePackage(ref string) *eaxmi.Package {
	return s.doc.ResolvePackage(ref)
}

func (s *Service) allPackages() []*eaxmi.Package { return s.doc.AllPackages() }
