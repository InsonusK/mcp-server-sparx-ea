package sparx

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Method 1: the model navigator tree — packages, diagrams and elements with
// their EA ids, like the Sparx EA project browser.

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

	// Notices, on the root node only, flags relationships in the model that the
	// ArchiMate rules classify as deny / warn — a heads-up on an imported model.
	Notices []string `json:"notices,omitempty"`
}

// Tree returns the whole model as a navigator tree, rooted at the EA root
// package (skipping the synthetic "EA_Model" wrapper).
func (s *Service) Tree() *Node {
	root := &Node{Kind: KindPackage, Name: s.doc.Root.Name, Path: ""}
	for _, p := range s.doc.Root.Packages {
		root.Children = append(root.Children, s.packageNode(p, ""))
	}
	root.Notices = s.ruleNotices()
	return root
}

// ruleNotices summarises every relationship in the model whose (type, source,
// target) the ArchiMate rules classify as deny or warn.
func (s *Service) ruleNotices() []string {
	var deny, warn int
	for _, c := range s.doc.Connectors() {
		switch v, rn, _, _ := s.connectorVerdict(c); {
		case rn == "":
			continue
		case v == VerdictDeny:
			deny++
		case v == VerdictWarn:
			warn++
		}
	}
	var out []string
	if deny > 0 {
		out = append(out, fmt.Sprintf("%d relationship(s) violate the ArchiMate rules (verdict deny) — read the elements to see which", deny))
	}
	if warn > 0 {
		out = append(out, fmt.Sprintf("%d relationship(s) are discouraged by the ArchiMate rules (verdict warn)", warn))
	}
	return out
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
