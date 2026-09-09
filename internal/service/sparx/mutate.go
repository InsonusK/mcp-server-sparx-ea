package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Methods 4-6. Every mutation validates against the ArchiMate vocabulary and
// relationship rules before it touches the document. The service never writes
// the original file — Save writes a copy the user imports back into EA.

// Rect is an element's rectangle on a diagram (EA coordinates: origin top-left).
type Rect struct {
	Left, Top, Right, Bottom int
}

// ---------- packages / root ----------

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

// RenamePackage renames a package (by id or path). Unlike SetRootName it keeps
// the package identity.
func (s *Service) RenamePackage(ref, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("sparx: new name is required")
	}
	pkg := s.resolvePackage(ref)
	if pkg == nil {
		return fmt.Errorf("sparx: no package for %q", ref)
	}
	return errWrap(s.doc.SetPackageName(pkg.XMIID, newName))
}

// ---------- method 4: elements ----------

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

// ---------- method 5: relationships ----------

// CreateRelationship adds a relationship of ArchiMate type archimateRelType
// (e.g. "ArchiMate.Serving") from sourceRef to targetRef. It fails if the
// relationship is not permitted between the two element types.
func (s *Service) CreateRelationship(sourceRef, targetRef, archimateRelType, name, documentation string) (*Relation, error) {
	relBare, ok := unqualify(archimateRelType)
	if !ok || !archimateRelationships[relBare] {
		return nil, fmt.Errorf("sparx: %q is not a known ArchiMate relationship type", archimateRelType)
	}
	src, _ := s.resolveElement(sourceRef)
	tgt, _ := s.resolveElement(targetRef)
	if src == nil {
		return nil, fmt.Errorf("sparx: no source element for %q", sourceRef)
	}
	if tgt == nil {
		return nil, fmt.Errorf("sparx: no target element for %q", targetRef)
	}
	srcBare, sok := archimateTypeFromStereotype(src.Stereotype)
	tgtBare, tok := archimateTypeFromStereotype(tgt.Stereotype)
	if !sok || !tok {
		return nil, fmt.Errorf("sparx: %s or %s is not an ArchiMate element; refusing to add a relationship", src.Name, tgt.Name)
	}
	if ok, reason := relationshipAllowed(relBare, srcBare, tgtBare); !ok {
		return nil, fmt.Errorf("sparx: %s from %s (%s) to %s (%s) is not allowed: %s",
			relBare, src.Name, qualify(srcBare), tgt.Name, qualify(tgtBare), reason)
	}
	// (type, source, target) is unique: no duplicate relationship.
	stereo := "ArchiMate_" + relBare
	for _, c := range s.doc.Connectors() {
		if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID && c.Stereotype == stereo {
			return nil, fmt.Errorf("sparx: a %s relationship from %s to %s already exists (%s)",
				relBare, src.Name, tgt.Name, c.XMIID)
		}
	}

	ea := relEA[relBare]
	conn, err := s.doc.AddConnector(eaxmi.ConnectorSpec{
		SourceID: src.XMIID, TargetID: tgt.XMIID, Name: strings.TrimSpace(name),
		EAType: ea.EAType, ModelRepr: ea.ModelRepr,
		Stereotype: stereo, Direction: ea.Direction,
		Documentation: documentation,
	})
	if err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return &Relation{
		ID: conn.XMIID, Type: qualify(relBare), Name: conn.Name, Direction: "outgoing",
		OtherID: tgt.XMIID, OtherName: tgt.Name, OtherType: qualify(tgtBare),
	}, nil
}

// DeleteRelationshipsBetween removes every relationship whose source is
// sourceRef and target is targetRef. Returns how many were removed; it is an
// error if there were none.
func (s *Service) DeleteRelationshipsBetween(sourceRef, targetRef string) (int, error) {
	src, _ := s.resolveElement(sourceRef)
	tgt, _ := s.resolveElement(targetRef)
	if src == nil {
		return 0, fmt.Errorf("sparx: no source element for %q", sourceRef)
	}
	if tgt == nil {
		return 0, fmt.Errorf("sparx: no target element for %q", targetRef)
	}
	var ids []string
	for _, c := range s.doc.Connectors() {
		if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID {
			ids = append(ids, c.XMIID)
		}
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("sparx: no relationship from %s to %s", src.Name, tgt.Name)
	}
	for _, id := range ids {
		if err := s.doc.RemoveConnector(id); err != nil {
			return 0, fmt.Errorf("sparx: %w", err)
		}
	}
	return len(ids), nil
}

// DeleteRelationship removes a relationship by its id.
func (s *Service) DeleteRelationship(id string) error {
	for _, c := range s.doc.Connectors() {
		if c.XMIID == id || strings.EqualFold(c.GUID, id) {
			if err := s.doc.RemoveConnector(c.XMIID); err != nil {
				return fmt.Errorf("sparx: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("sparx: no relationship %q", id)
}

// ---------- method 6: elements on a diagram ----------

// AddToDiagram places an element on a diagram at the given rectangle.
func (s *Service) AddToDiagram(diagramRef, elementRef string, at Rect) error {
	d, _ := s.resolveDiagram(diagramRef)
	if d == nil {
		return fmt.Errorf("sparx: no diagram for %q", diagramRef)
	}
	el, _ := s.resolveElement(elementRef)
	if el == nil {
		return fmt.Errorf("sparx: no element for %q", elementRef)
	}
	for _, o := range d.Objects {
		if o.SubjectID == el.XMIID {
			return fmt.Errorf("sparx: %s is already on diagram %s", el.Name, d.Name)
		}
	}
	if !validRect(at) {
		return fmt.Errorf("sparx: invalid rectangle %+v (need right>left, bottom>top)", at)
	}
	return errWrap(s.doc.AddDiagramObject(d.XMIID, el.XMIID, at.Left, at.Top, at.Right, at.Bottom))
}

// RemoveFromDiagram removes an element's placement from a diagram.
func (s *Service) RemoveFromDiagram(diagramRef, elementRef string) error {
	d, _ := s.resolveDiagram(diagramRef)
	if d == nil {
		return fmt.Errorf("sparx: no diagram for %q", diagramRef)
	}
	el, _ := s.resolveElement(elementRef)
	if el == nil {
		return fmt.Errorf("sparx: no element for %q", elementRef)
	}
	return errWrap(s.doc.RemoveDiagramObject(d.XMIID, el.XMIID))
}

// MoveOnDiagram changes an element's rectangle on a diagram (remove + re-add).
func (s *Service) MoveOnDiagram(diagramRef, elementRef string, at Rect) error {
	if !validRect(at) {
		return fmt.Errorf("sparx: invalid rectangle %+v", at)
	}
	if err := s.RemoveFromDiagram(diagramRef, elementRef); err != nil {
		return err
	}
	return s.AddToDiagram(diagramRef, elementRef, at)
}

// ---------- persistence ----------

// Save writes the edited model to path (never the file it was opened from).
func (s *Service) Save(path string) error {
	if path == s.doc.Path {
		return fmt.Errorf("sparx: refusing to overwrite the source file %q", path)
	}
	return errWrap(s.doc.WriteFile(path))
}

// ---------- helpers ----------

func (s *Service) childElement(p *eaxmi.Package, name string) *eaxmi.Element {
	for _, e := range p.Elements {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func validRect(r Rect) bool { return r.Right > r.Left && r.Bottom > r.Top }

func errWrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("sparx: %w", err)
}
