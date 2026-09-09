// Package sparx is the ArchiMate-facing service over an EA XMI file. It gives
// the MCP server simple, validated operations instead of raw SQL / raw XML:
// element types are ArchiMate types (the uml:Class + stereotype pair is hidden),
// and every mutation is validated before it touches the file.
//
// It reads (and, iteratively, writes) EA "Export Package to XMI" files via
// client/eaxmi.
package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Service is bound to one XMI file.
type Service struct {
	doc *eaxmi.Document
}

// Open loads the XMI file at path.
func Open(path string) (*Service, error) {
	doc, err := eaxmi.Open(path)
	if err != nil {
		return nil, err
	}
	return &Service{doc: doc}, nil
}

// ---------- method 1: model tree ----------

// NodeKind is "package", "diagram" or "element".
type NodeKind string

const (
	KindPackage NodeKind = "package"
	KindDiagram NodeKind = "diagram"
	KindElement NodeKind = "element"
)

// Node is one entry in the model navigator tree.
type Node struct {
	Kind     NodeKind `json:"kind"`
	Name     string   `json:"name"`
	ID       string   `json:"id"`             // EA xmi:id (EAPK_/EAID_)
	GUID     string   `json:"guid"`           // {…}
	Type     string   `json:"type,omitempty"` // ArchiMate type for elements, e.g. "ArchiMate.Goal"
	Path     string   `json:"path"`
	Children []*Node  `json:"children,omitempty"`
}

// Tree returns the whole model as a navigator tree, rooted at the EA root
// package (skipping the synthetic "EA_Model" wrapper).
func (s *Service) Tree() *Node {
	root := &Node{Kind: KindPackage, Name: s.doc.Root.Name, Path: ""}
	for _, p := range s.doc.Root.Packages {
		root.Children = append(root.Children, s.packageNode(p, ""))
	}
	return root
}

func (s *Service) packageNode(p *eaxmi.Package, parentPath string) *Node {
	path := joinPath(parentPath, p.Name)
	n := &Node{Kind: KindPackage, Name: p.Name, ID: p.XMIID, GUID: p.GUID, Path: path}
	for _, sub := range p.Packages {
		n.Children = append(n.Children, s.packageNode(sub, path))
	}
	for _, d := range p.Diagrams {
		n.Children = append(n.Children, &Node{
			Kind: KindDiagram, Name: d.Name, ID: d.XMIID, GUID: d.GUID,
			Path: joinPath(path, d.Name),
		})
	}
	for _, e := range p.Elements {
		n.Children = append(n.Children, &Node{
			Kind: KindElement, Name: e.Name, ID: e.XMIID, GUID: e.GUID,
			Type: elementType(e), Path: joinPath(path, e.Name),
		})
	}
	return n
}

// ---------- method 2: element info ----------

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

// ---------- method 3: diagram contents ----------

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

// PlacedElement is an element shown on a diagram, with its rectangle (EA
// coordinates: origin top-left, y downward).
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

// ---------- helpers ----------

func elementType(e *eaxmi.Element) string {
	name, ok := archimateTypeFromStereotype(e.Stereotype)
	if !ok || !archimateElements[name] {
		if e.Stereotype == "" {
			return "unknown:" + strings.TrimPrefix(e.UMLType, "uml:")
		}
		return "unknown:" + e.Stereotype
	}
	return qualify(name)
}

func relationType(c *eaxmi.Connector) string {
	name, ok := archimateTypeFromStereotype(c.Stereotype)
	if !ok || !archimateRelationships[name] {
		if c.Stereotype == "" {
			return "unknown:" + c.EAType
		}
		return "unknown:" + c.Stereotype
	}
	return qualify(name)
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}
