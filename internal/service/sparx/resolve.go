package sparx

import (
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// A ref is one of:
//   - an EA xmi:id      "EAID_F3C2209E_B146_4328_863E_622BABD2B43E"
//   - a braced GUID     "{F3C2209E-B146-4328-863E-622BABD2B43E}"
//   - a slash path      "Model/Motivation_Package/Goal1" (from the EA root package)

func looksLikeID(ref string) bool {
	return strings.HasPrefix(ref, "EAID_") || strings.HasPrefix(ref, "EAPK_") ||
		(strings.HasPrefix(ref, "{") && strings.HasSuffix(ref, "}"))
}

func (s *Service) resolveElement(ref string) (*eaxmi.Element, string) {
	if looksLikeID(ref) {
		for _, e := range s.allElements() {
			if e.XMIID == ref || e.GUID == ref || strings.EqualFold(e.GUID, ref) {
				return e, s.pathOfElement(e)
			}
		}
		return nil, ""
	}
	segs := splitPath(ref)
	if len(segs) == 0 {
		return nil, ""
	}
	pkg := s.walkPackages(segs[:len(segs)-1])
	if pkg == nil {
		return nil, ""
	}
	last := segs[len(segs)-1]
	for _, e := range pkg.Elements {
		if e.Name == last {
			return e, ref
		}
	}
	return nil, ""
}

func (s *Service) resolveDiagram(ref string) (*eaxmi.Diagram, string) {
	if looksLikeID(ref) {
		for _, p := range s.allPackages() {
			for _, d := range p.Diagrams {
				if d.XMIID == ref || strings.EqualFold(d.GUID, ref) {
					return d, s.pathOfDiagram(d)
				}
			}
		}
		return nil, ""
	}
	segs := splitPath(ref)
	if len(segs) == 0 {
		return nil, ""
	}
	pkg := s.walkPackages(segs[:len(segs)-1])
	if pkg == nil {
		return nil, ""
	}
	last := segs[len(segs)-1]
	for _, d := range pkg.Diagrams {
		if d.Name == last {
			return d, ref
		}
	}
	return nil, ""
}

// walkPackages follows package-name segments from the EA root package downward.
func (s *Service) walkPackages(segs []string) *eaxmi.Package {
	cur := s.doc.Root // the synthetic "EA_Model"; its children are EA root packages
	for _, seg := range segs {
		var next *eaxmi.Package
		for _, p := range cur.Packages {
			if p.Name == seg {
				next = p
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func (s *Service) allPackages() []*eaxmi.Package {
	var out []*eaxmi.Package
	var rec func(p *eaxmi.Package)
	rec = func(p *eaxmi.Package) {
		for _, sub := range p.Packages {
			out = append(out, sub)
			rec(sub)
		}
	}
	rec(s.doc.Root)
	return out
}

func (s *Service) allElements() []*eaxmi.Element {
	var out []*eaxmi.Element
	for _, p := range s.allPackages() {
		out = append(out, p.Elements...)
	}
	return out
}

func (s *Service) pathOfPackage(p *eaxmi.Package) string {
	var segs []string
	for cur := p; cur != nil && cur != s.doc.Root; cur = cur.Parent {
		segs = append([]string{cur.Name}, segs...)
	}
	return strings.Join(segs, "/")
}

func (s *Service) pathOfElement(e *eaxmi.Element) string {
	// The parser always attaches an element to its package.
	return joinPath(s.pathOfPackage(e.Package), e.Name)
}

func (s *Service) pathOfDiagram(d *eaxmi.Diagram) string {
	if d.Package == nil {
		return d.Name
	}
	return joinPath(s.pathOfPackage(d.Package), d.Name)
}

func splitPath(ref string) []string {
	var out []string
	for _, seg := range strings.Split(ref, "/") {
		if seg = strings.TrimSpace(seg); seg != "" {
			out = append(out, seg)
		}
	}
	return out
}
