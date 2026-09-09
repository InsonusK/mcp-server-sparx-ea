package eaxmi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// connectorUMLTypes are the xmi:type values that denote a relationship rather
// than a model element when they appear as <packagedElement>.
var connectorUMLTypes = map[string]bool{
	"uml:Dependency": true, "uml:Association": true, "uml:Realization": true,
	"uml:Abstraction": true, "uml:Usage": true, "uml:Generalization": true,
	"uml:ControlFlow": true, "uml:ObjectFlow": true, "uml:InformationFlow": true,
	"uml:Deployment": true, "uml:Aggregation": true, "uml:NestingLink": true,
	"uml:PackageMerge": true, "uml:PackageImport": true,
}

func parse(doc *etree.Document) (*Document, error) {
	xmi := doc.SelectElement("XMI")
	if xmi == nil {
		return nil, fmt.Errorf("eaxmi: no <xmi:XMI> root element")
	}
	umlModel := xmi.FindElement("Model")
	if umlModel == nil {
		return nil, fmt.Errorf("eaxmi: no <uml:Model> element")
	}

	d := &Document{
		doc:         doc,
		elementByID: map[string]*Element{},
		diagramByID: map[string]*Diagram{},
		packageByID: map[string]*Package{},
	}
	if docEl := xmi.FindElement("Documentation"); docEl != nil {
		d.ExporterVersion = docEl.SelectAttrValue("exporterVersion", "")
	}

	root := &Package{Name: umlModel.SelectAttrValue("name", "EA_Model")}
	d.Root = root
	d.parseModelTree(umlModel, root)

	if ext := xmi.FindElement("Extension"); ext != nil {
		d.enrichElements(ext)
		d.enrichConnectors(ext)
		d.parseDiagrams(ext)
	}
	return d, nil
}

// parseModelTree walks a package's <packagedElement> children.
func (d *Document) parseModelTree(node *etree.Element, pkg *Package) {
	for _, child := range node.ChildElements() {
		tag := child.Tag
		typ := child.SelectAttrValue("xmi:type", "")

		switch {
		case tag == "packagedElement" && typ == "uml:Package":
			sub := &Package{
				XMIID:  child.SelectAttrValue("xmi:id", ""),
				Name:   child.SelectAttrValue("name", ""),
				Parent: pkg,
			}
			sub.GUID = guidFromXMIID(sub.XMIID)
			pkg.Packages = append(pkg.Packages, sub)
			if sub.XMIID != "" {
				d.packageByID[sub.XMIID] = sub
			}
			d.parseModelTree(child, sub)

		case tag == "edge" || connectorUMLTypes[typ] ||
			child.SelectAttr("source") != nil || child.SelectAttr("supplier") != nil:
			// Relationship. The authoritative record is in <xmi:Extension>;
			// here we only register its id so connectors declared solely in the
			// model tree are not lost.
			id := child.SelectAttrValue("xmi:id", "")
			if id != "" && d.connectorByID(id) == nil {
				d.connectorAll = append(d.connectorAll, &Connector{
					XMIID:  id,
					GUID:   guidFromXMIID(id),
					Name:   child.SelectAttrValue("name", ""),
					EAType: strings.TrimPrefix(typ, "uml:"),
					SourceID: firstNonEmpty(
						child.SelectAttrValue("source", ""),
						child.SelectAttrValue("client", ""),
					),
					TargetID: firstNonEmpty(
						child.SelectAttrValue("target", ""),
						child.SelectAttrValue("supplier", ""),
					),
				})
			}

		case tag == "packagedElement" && typ != "":
			el := &Element{
				XMIID:   child.SelectAttrValue("xmi:id", ""),
				Name:    child.SelectAttrValue("name", ""),
				UMLType: typ,
				Package: pkg,
			}
			el.GUID = guidFromXMIID(el.XMIID)
			pkg.Elements = append(pkg.Elements, el)
			if el.XMIID != "" {
				d.elementByID[el.XMIID] = el
			}
		}
	}
}

func (d *Document) enrichElements(ext *etree.Element) {
	for _, elsBlock := range ext.SelectElements("elements") {
		for _, e := range elsBlock.SelectElements("element") {
			id := e.SelectAttrValue("xmi:idref", "")
			if strings.HasPrefix(id, "EAPK") {
				if pkg := d.packageByID[id]; pkg != nil {
					if m := e.SelectElement("model"); m != nil {
						pkg.EALocalID = m.SelectAttrValue("ea_localid", "")
					}
				}
				continue
			}
			el := d.elementByID[id]
			if el == nil {
				continue
			}
			if m := e.SelectElement("model"); m != nil {
				el.EALocalID = m.SelectAttrValue("ea_localid", "")
			}
			if p := e.SelectElement("properties"); p != nil {
				el.Stereotype = p.SelectAttrValue("stereotype", "")
				el.Documentation = p.SelectAttrValue("documentation", "")
			}
			if pr := e.SelectElement("project"); pr != nil {
				el.Author = pr.SelectAttrValue("author", "")
			}
		}
	}
}

func (d *Document) enrichConnectors(ext *etree.Element) {
	for _, consBlock := range ext.SelectElements("connectors") {
		for _, c := range consBlock.SelectElements("connector") {
			id := c.SelectAttrValue("xmi:idref", "")
			conn := d.connectorByID(id)
			if conn == nil {
				conn = &Connector{XMIID: id, GUID: guidFromXMIID(id)}
				d.connectorAll = append(d.connectorAll, conn)
			}
			if n := c.SelectAttrValue("name", ""); n != "" {
				conn.Name = n
			}
			if s := c.SelectElement("source"); s != nil {
				conn.SourceID = s.SelectAttrValue("xmi:idref", "")
			}
			if t := c.SelectElement("target"); t != nil {
				conn.TargetID = t.SelectAttrValue("xmi:idref", "")
			}
			if m := c.SelectElement("model"); m != nil {
				conn.EALocalID = m.SelectAttrValue("ea_localid", "")
			}
			if p := c.SelectElement("properties"); p != nil {
				if v := p.SelectAttrValue("ea_type", ""); v != "" {
					conn.EAType = v
				}
				conn.Stereotype = p.SelectAttrValue("stereotype", "")
				conn.Direction = p.SelectAttrValue("direction", "")
			}
			if doc := c.SelectElement("documentation"); doc != nil {
				conn.Documentation = doc.SelectAttrValue("value", "")
			}
		}
	}
}

func (d *Document) parseDiagrams(ext *etree.Element) {
	for _, diasBlock := range ext.SelectElements("diagrams") {
		for _, dg := range diasBlock.SelectElements("diagram") {
			g := &Diagram{
				XMIID: dg.SelectAttrValue("xmi:id", ""),
			}
			g.GUID = guidFromXMIID(g.XMIID)
			if m := dg.SelectElement("model"); m != nil {
				g.EALocalID = m.SelectAttrValue("localID", "")
				owner := firstNonEmpty(m.SelectAttrValue("owner", ""), m.SelectAttrValue("package", ""))
				if pkg := d.packageByID[owner]; pkg != nil {
					g.Package = pkg
					pkg.Diagrams = append(pkg.Diagrams, g)
				}
			}
			if p := dg.SelectElement("properties"); p != nil {
				g.Name = p.SelectAttrValue("name", "")
				g.Type = p.SelectAttrValue("type", "")
				g.Documentation = p.SelectAttrValue("documentation", "")
			}
			if elsBlock := dg.SelectElement("elements"); elsBlock != nil {
				for _, e := range elsBlock.SelectElements("element") {
					subj := e.SelectAttrValue("subject", "")
					geo := e.SelectAttrValue("geometry", "")
					if strings.HasPrefix(geo, "SX=") || strings.HasPrefix(geo, "SX =") {
						// connector placement (EA auto-routes; no rectangle)
						g.Links = append(g.Links, &DiagramLink{ConnectorID: subj})
						continue
					}
					obj := &DiagramObject{
						SubjectID: subj,
						Seq:       atoiOr(e.SelectAttrValue("seqno", ""), 0),
					}
					obj.Left, obj.Top, obj.Right, obj.Bottom = parseGeometry(geo)
					g.Objects = append(g.Objects, obj)
				}
			}
			if g.XMIID != "" {
				d.diagramByID[g.XMIID] = g
			}
		}
	}
}

func (d *Document) connectorByID(id string) *Connector {
	for _, c := range d.connectorAll {
		if c.XMIID == id {
			return c
		}
	}
	return nil
}

// parseGeometry parses "Left=520;Top=680;Right=610;Bottom=750;".
func parseGeometry(s string) (left, top, right, bottom int) {
	for _, kv := range strings.Split(s, ";") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		n := atoiOr(v, 0)
		switch k {
		case "Left":
			left = n
		case "Top":
			top = n
		case "Right":
			right = n
		case "Bottom":
			bottom = n
		}
	}
	return
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
