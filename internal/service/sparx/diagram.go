package sparx

import (
	"fmt"
	"sort"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Diagrams: read contents and layout (method 3), create a diagram (method 6),
// and place / move / remove an element on a diagram (method 6).

// archimateDiagramLayers is the set of ArchiMate viewpoints CreateDiagram
// accepts as its layer argument. Each maps to an EA ArchiMate3 MDG diagram
// profile ("Motivation" -> "ArchiMate3::Motivation"), which selects the toolbox
// EA shows when the diagram is opened. The empty layer is also allowed — a
// plain logical diagram, on which ArchiMate elements still render with their
// stereotyped shapes.
var archimateDiagramLayers = map[string]bool{
	"Motivation":               true,
	"Strategy":                 true,
	"Business":                 true,
	"Application":              true,
	"Technology":               true,
	"Physical":                 true,
	"Implementation_Migration": true,
}

// DiagramLayers returns the accepted CreateDiagram layer names, sorted.
func DiagramLayers() []string {
	out := make([]string, 0, len(archimateDiagramLayers))
	for name := range archimateDiagramLayers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// mdgForLayer validates a layer argument and returns the EA MDG diagram type to
// store. "" is valid and yields "" (a plain logical diagram).
func mdgForLayer(layer string) (string, error) {
	layer = strings.TrimSpace(layer)
	if layer == "" {
		return "", nil
	}
	if !archimateDiagramLayers[layer] {
		return "", fmt.Errorf("sparx: %q is not a known ArchiMate diagram layer; use one of %s (or omit it)",
			layer, strings.Join(DiagramLayers(), ", "))
	}
	return "ArchiMate3::" + layer, nil
}

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
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	Verdict string `json:"verdict,omitempty"` // "warn" | "deny" per the ArchiMate rules
	Warning string `json:"warning,omitempty"`
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
				v, rn, sn, tn := s.connectorVerdict(c)
				annotateVerdict(&pl.Verdict, &pl.Warning, v, rn, sn, tn)
				break
			}
		}
		info.Links = append(info.Links, pl)
	}
	return info, nil
}

// CreateDiagram adds an empty diagram named name inside the package at
// parentRef. layer, when given, is an ArchiMate viewpoint (see DiagramLayers)
// that selects EA's toolbox for the diagram; omit it for a plain logical
// diagram. Populate the diagram with AddToDiagram.
func (s *Service) CreateDiagram(parentRef, name, layer string) (*DiagramInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("sparx: diagram name is required")
	}
	mdg, err := mdgForLayer(layer)
	if err != nil {
		return nil, err
	}
	parent := s.resolvePackage(parentRef)
	if parent == nil {
		return nil, fmt.Errorf("sparx: no package for %q", parentRef)
	}
	for _, g := range parent.Diagrams {
		if g.Name == name {
			return nil, fmt.Errorf("sparx: package %q already has a diagram named %q", parent.Name, name)
		}
	}
	g, err := s.doc.AddDiagram(parent.XMIID, name, mdg)
	if err != nil {
		return nil, errWrap(err)
	}
	return s.Diagram(g.XMIID)
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
	if err := s.doc.AddDiagramObject(d.XMIID, el.XMIID, at.Left, at.Top, at.Right, at.Bottom); err != nil {
		return errWrap(err)
	}
	s.showRelatedConnectors(d, el.XMIID)
	return nil
}

// showRelatedConnectors adds a diagram line for every connector between the
// just-placed element newID and an element already on the diagram, so dropping
// two connected elements on a diagram shows the relationship between them.
func (s *Service) showRelatedConnectors(d *eaxmi.Diagram, newID string) {
	placed := map[string]bool{}
	for _, o := range d.Objects {
		placed[o.SubjectID] = true
	}
	for _, c := range s.doc.Connectors() {
		var other string
		switch newID {
		case c.SourceID:
			other = c.TargetID
		case c.TargetID:
			other = c.SourceID
		default:
			continue
		}
		if placed[other] {
			_ = s.doc.AddDiagramLink(d.XMIID, c.XMIID)
		}
	}
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
	if err := s.doc.RemoveDiagramObject(d.XMIID, el.XMIID); err != nil {
		return errWrap(err)
	}
	// Drop any connector lines that can no longer be drawn (one end just left).
	for _, c := range s.doc.Connectors() {
		if c.SourceID == el.XMIID || c.TargetID == el.XMIID {
			s.doc.RemoveDiagramLink(d.XMIID, c.XMIID)
		}
	}
	return nil
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
