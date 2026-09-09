package eaxmi

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// Creation of diagrams. A diagram is an <xmi:Extension>/<diagrams>/<diagram>
// record; EA holds no <packagedElement> for it. AddDiagramObject then fills it
// with element placements. The shape follows a real EA "Export Package to XMI"
// (client/eaxmi/test/testdata/CyrillicProject.xml); whether EA's importer needs
// every child is confirmed by round-tripping through EA.

// diagramsBlock returns <xmi:Extension>/<diagrams>, creating it if the export
// (or a from-scratch model) does not carry one yet.
func (d *Document) diagramsBlock() *etree.Element {
	ext := d.extension()
	if b := ext.FindElement("diagrams"); b != nil {
		return b
	}
	return ext.CreateElement("diagrams")
}

// AddDiagram creates an empty diagram named name inside package pkgID and
// returns it. mdgType, when non-empty, is written into the diagram style as
// "MDGDgm=<mdgType>" (e.g. "ArchiMate3::Motivation") so EA opens the diagram
// with the matching ArchiMate toolbox; the diagram's <properties type> stays
// "Logical", as EA exports every extended diagram. Populate it with
// AddDiagramObject.
func (d *Document) AddDiagram(pkgID, name, mdgType string) (*Diagram, error) {
	pkg, ok := d.packageByID[pkgID]
	if !ok {
		return nil, fmt.Errorf("eaxmi: no package %q", pkgID)
	}
	for _, g := range pkg.Diagrams {
		if g.Name == name {
			return nil, fmt.Errorf("eaxmi: package %q already has a diagram named %q", pkg.Name, name)
		}
	}

	guid := NewGUID()
	id := xmiIDFromGUID(guid, "EAID_")

	dg := d.diagramsBlock().CreateElement("diagram")
	dg.CreateAttr("xmi:id", id)

	m := dg.CreateElement("model")
	m.CreateAttr("package", pkgID)
	m.CreateAttr("owner", pkgID)
	m.CreateAttr("localID", "0")

	pr := dg.CreateElement("properties")
	pr.CreateAttr("name", name)
	pr.CreateAttr("type", "Logical")
	pr.CreateAttr("documentation", "")

	proj := dg.CreateElement("project")
	proj.CreateAttr("author", genAuthor)
	proj.CreateAttr("version", "1.0")

	style := "ExcludeRTF=0;DocAll=0;HideQuals=0;AttPkg=1;ShowTests=0;ShowMaint=0;" +
		"AdvancedElementProps=1;AdvancedFeatureProps=1;AdvancedConnectorProps=1;" +
		"m_bElementClassifier=1;SPT=1;ShowNotes=0;SuppressBrackets=0;SuppConnectorLabels=0;"
	if mdgType != "" {
		style += "MDGDgm=" + mdgType + ";"
	}
	dg.CreateElement("style2").CreateAttr("value", style)

	dg.CreateElement("elements")

	g := &Diagram{XMIID: id, GUID: guid, Name: name, Type: "Logical", Package: pkg}
	pkg.Diagrams = append(pkg.Diagrams, g)
	d.diagramByID[id] = g
	return g, nil
}

// AddDiagramLink shows connectorID on diagramID. EA auto-routes the line as long
// as both of the connector's endpoint elements are placed on the diagram; it is
// stored as an <element> with an "SX=" geometry (the shape the parser reads back
// as a DiagramLink). A no-op if the connector is already shown.
func (d *Document) AddDiagramLink(diagramID, connectorID string) error {
	g, ok := d.diagramByID[diagramID]
	if !ok {
		return fmt.Errorf("eaxmi: no diagram %q", diagramID)
	}
	for _, l := range g.Links {
		if l.ConnectorID == connectorID {
			return nil
		}
	}
	dg := d.extension().FindElement("//diagram[@xmi:id='" + diagramID + "']")
	if dg == nil {
		return fmt.Errorf("eaxmi: diagram node %q not found", diagramID)
	}
	els := dg.FindElement("elements")
	if els == nil {
		els = dg.CreateElement("elements")
	}
	o := els.CreateElement("element")
	o.CreateAttr("geometry", "SX=0;SY=0;EX=0;EY=0;EDGE=1;$LLB=;LLT=;LMT=;LMB=;LRT=;LRB=;IRHS=;ILHS=;Path=;")
	o.CreateAttr("subject", connectorID)
	o.CreateAttr("style", "Mode=3;Color=-1;LWidth=0;Hidden=0;")

	g.Links = append(g.Links, &DiagramLink{ConnectorID: connectorID})
	return nil
}

// RemoveDiagramLink hides connectorID on diagramID. A no-op if it is not shown.
func (d *Document) RemoveDiagramLink(diagramID, connectorID string) {
	g, ok := d.diagramByID[diagramID]
	if !ok {
		return
	}
	if dg := d.extension().FindElement("//diagram[@xmi:id='" + diagramID + "']"); dg != nil {
		if els := dg.FindElement("elements"); els != nil {
			for _, o := range els.ChildElements() {
				if o.SelectAttrValue("subject", "") == connectorID &&
					strings.HasPrefix(o.SelectAttrValue("geometry", ""), "SX=") {
					els.RemoveChild(o)
				}
			}
		}
	}
	kept := g.Links[:0]
	for _, l := range g.Links {
		if l.ConnectorID != connectorID {
			kept = append(kept, l)
		}
	}
	g.Links = kept
}
