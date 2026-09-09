package sparx

import "fmt"

// Diagrams: read contents and layout (method 3), and place / move / remove an
// element on a diagram (method 6).

// Rect is an element's rectangle on a diagram (EA coordinates: origin top-left,
// y downward).
type Rect struct {
	Left, Top, Right, Bottom int
}

// DiagramInfo lists what is placed on a diagram and where.
type DiagramInfo struct {
	ID          string          `json:"id"`
	GUID        string          `json:"guid"`
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	DiagramType string          `json:"diagramType"`
	Objects     []PlacedElement `json:"objects"`
	Links       []PlacedLink    `json:"links"`
}

// PlacedElement is an element shown on a diagram, with its rectangle.
type PlacedElement struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Seq    int    `json:"seq"`
	Left   int    `json:"left"`
	Top    int    `json:"top"`
	Right  int    `json:"right"`
	Bottom int    `json:"bottom"`
}

// PlacedLink is a connector shown on a diagram (EA auto-routes it).
type PlacedLink struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// Diagram resolves ref (an ID, a GUID, or a slash path) and returns its contents.
func (s *Service) Diagram(ref string) (*DiagramInfo, error) {
	d, path := s.resolveDiagram(ref)
	if d == nil {
		return nil, fmt.Errorf("sparx: no diagram for %q", ref)
	}
	info := &DiagramInfo{
		ID: d.XMIID, GUID: d.GUID, Name: d.Name, Path: path, DiagramType: d.Type,
		Objects: []PlacedElement{}, Links: []PlacedLink{},
	}
	for _, o := range d.Objects {
		pe := PlacedElement{
			ID: o.SubjectID, Seq: o.Seq,
			Left: o.Left, Top: o.Top, Right: o.Right, Bottom: o.Bottom,
		}
		if el, ok := s.doc.ElementByID(o.SubjectID); ok {
			pe.Name = el.Name
			pe.Type = elementType(el)
		}
		info.Objects = append(info.Objects, pe)
	}
	for _, l := range d.Links {
		pl := PlacedLink{ID: l.ConnectorID}
		for _, c := range s.doc.Connectors() {
			if c.XMIID == l.ConnectorID {
				pl.Type = relationType(c)
				pl.Name = c.Name
				break
			}
		}
		info.Links = append(info.Links, pl)
	}
	return info, nil
}

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

func validRect(r Rect) bool { return r.Right > r.Left && r.Bottom > r.Top }
