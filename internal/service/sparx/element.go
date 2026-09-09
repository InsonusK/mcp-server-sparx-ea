package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Elements: read (method 2) and the create / rename / re-note / move / delete
// mutations (method 4). Every mutation validates against the ArchiMate
// vocabulary before it touches the document.

// ElementInfo is the answer to "tell me about this element".
type ElementInfo struct {
	ID            string     `json:"id"`
	GUID          string     `json:"guid"`
	Name          string     `json:"name"`
	Type          string     `json:"type"` // "ArchiMate.Goal", or "unknown:<stereotype>"
	Path          string     `json:"path"`
	Documentation string     `json:"documentation,omitempty"`
	Relations     []Relation `json:"relations"`
}

// Relation is one connector as seen from a particular element.
type Relation struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "ArchiMate.Realization", or "unknown:<stereotype>"
	Name      string `json:"name,omitempty"`
	Direction string `json:"direction"` // "outgoing" | "incoming"
	OtherName string `json:"otherName"`
	OtherID   string `json:"otherId"`
	OtherType string `json:"otherType"`
}

// Element resolves ref (an ID, a GUID, or a slash path) and returns its details.
func (s *Service) Element(ref string) (*ElementInfo, error) {
	el, path := s.resolveElement(ref)
	if el == nil {
		return nil, fmt.Errorf("sparx: no element for %q", ref)
	}
	info := &ElementInfo{
		ID: el.XMIID, GUID: el.GUID, Name: el.Name, Type: elementType(el),
		Path: path, Documentation: el.Documentation,
		Relations: []Relation{},
	}
	for _, c := range s.doc.Connectors() {
		var dir, otherID string
		switch el.XMIID {
		case c.SourceID:
			dir, otherID = "outgoing", c.TargetID
		case c.TargetID:
			dir, otherID = "incoming", c.SourceID
		default:
			continue
		}
		other, _ := s.doc.ElementByID(otherID)
		rel := Relation{
			ID: c.XMIID, Type: relationType(c), Name: c.Name, Direction: dir,
			OtherID: otherID,
		}
		if other != nil {
			rel.OtherName = other.Name
			rel.OtherType = elementType(other)
		}
		info.Relations = append(info.Relations, rel)
	}
	return info, nil
}

// CreateElement adds an element of ArchiMate type archimateType (e.g.
// "ArchiMate.Goal") named name into the package at parentRef (an id or a path).
func (s *Service) CreateElement(parentRef, archimateType, name, documentation string) (*ElementInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("sparx: element name is required")
	}
	bare, ok := unqualify(archimateType)
	if !ok || !archimateElements[bare] {
		return nil, fmt.Errorf("sparx: %q is not a known ArchiMate element type", archimateType)
	}
	pkg := s.resolvePackage(parentRef)
	if pkg == nil {
		return nil, fmt.Errorf("sparx: no package for %q", parentRef)
	}
	if s.childElement(pkg, name) != nil {
		return nil, fmt.Errorf("sparx: package %q already has an element named %q", pkg.Name, name)
	}

	ea := eaForElement(bare)
	el, err := s.doc.AddElement(pkg.XMIID, eaxmi.ElementSpec{
		Name: name, UMLType: ea.UMLType, Stereotype: ea.Stereotype,
		Documentation: documentation,
	})
	if err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Element(el.XMIID)
}

// RenameElement changes an element's name.
func (s *Service) RenameElement(ref, newName string) (*ElementInfo, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, fmt.Errorf("sparx: new name is required")
	}
	el, _ := s.resolveElement(ref)
	if el == nil {
		return nil, fmt.Errorf("sparx: no element for %q", ref)
	}
	if err := s.doc.SetElementName(el.XMIID, newName); err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Element(el.XMIID)
}

// SetElementDocumentation changes an element's note.
func (s *Service) SetElementDocumentation(ref, documentation string) (*ElementInfo, error) {
	el, _ := s.resolveElement(ref)
	if el == nil {
		return nil, fmt.Errorf("sparx: no element for %q", ref)
	}
	if err := s.doc.SetElementDocumentation(el.XMIID, documentation); err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Element(el.XMIID)
}

// MoveElement re-parents an element into the package at newParentRef.
func (s *Service) MoveElement(ref, newParentRef string) (*ElementInfo, error) {
	el, _ := s.resolveElement(ref)
	if el == nil {
		return nil, fmt.Errorf("sparx: no element for %q", ref)
	}
	pkg := s.resolvePackage(newParentRef)
	if pkg == nil {
		return nil, fmt.Errorf("sparx: no package for %q", newParentRef)
	}
	if el.Package != nil && el.Package.XMIID == pkg.XMIID {
		return s.Element(el.XMIID)
	}
	if other := s.childElement(pkg, el.Name); other != nil {
		return nil, fmt.Errorf("sparx: package %q already has an element named %q", pkg.Name, el.Name)
	}
	if err := s.doc.MoveElement(el.XMIID, pkg.XMIID); err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Element(el.XMIID)
}

// DeleteElement removes an element and every relationship attached to it.
func (s *Service) DeleteElement(ref string) error {
	el, _ := s.resolveElement(ref)
	if el == nil {
		return fmt.Errorf("sparx: no element for %q", ref)
	}
	if err := s.doc.RemoveElement(el.XMIID); err != nil {
		return fmt.Errorf("sparx: %w", err)
	}
	return nil
}

func (s *Service) childElement(p *eaxmi.Package, name string) *eaxmi.Element {
	for _, e := range p.Elements {
		if e.Name == name {
			return e
		}
	}
	return nil
}
