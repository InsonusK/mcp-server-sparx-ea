package eaxmi

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// This file adds mutation to the codec. Edits are applied to the underlying
// etree document so the parts eaxmi does not model are preserved; WriteFile
// serialises the whole thing back. The generated shape follows a real EA
// "Export Package to XMI" (see example/TestProject2-Model.xml); whether EA's
// importer needs every detail is confirmed by round-tripping through EA.

// NewGUID returns a fresh EA-style GUID, e.g. "{A0D485C0-C57E-4A4D-8FF1-66D594FA6778}".
func NewGUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("{%X-%X-%X-%X-%X}", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// xmiIDFromGUID converts "{A-B-C-D-E}" to "EAID_A_B_C_D_E" (or the EAPK_ form).
func xmiIDFromGUID(guid, prefix string) string {
	body := strings.TrimSuffix(strings.TrimPrefix(guid, "{"), "}")
	return prefix + strings.ReplaceAll(body, "-", "_")
}

// ElementSpec describes an element to create.
type ElementSpec struct {
	Name          string
	UMLType       string // "uml:Class" / "uml:Activity"
	Stereotype    string // "ArchiMate_Goal"
	Documentation string
	Author        string
}

// ConnectorSpec describes a relationship to create.
type ConnectorSpec struct {
	SourceID      string
	TargetID      string
	Name          string
	EAType        string // "Association" / "ControlFlow" / "Dependency" / "Generalization"
	ModelRepr     string // "association" / "controlflow" / "dependency" / "generalization"
	Stereotype    string // "ArchiMate_Composition"
	Direction     string
	Documentation string
	Author        string
}

const genAuthor = "mcp-server-sparx-ea"

// modelPackageElement returns the <packagedElement xmi:type="uml:Package"> node
// for a package id, or the <uml:Model> node for the root.
func (d *Document) modelPackageElement(pkgID string) *etree.Element {
	umlModel := d.doc.FindElement("//Model")
	if pkgID == "" {
		return umlModel
	}
	return umlModel.FindElement("//packagedElement[@xmi:id='" + pkgID + "']")
}

func (d *Document) extension() *etree.Element {
	return d.doc.FindElement("//Extension")
}

// AddElement creates an element in the package pkgID (an "" pkgID means the
// model root). It updates both the model tree and the <xmi:Extension> block and
// returns the new Element.
func (d *Document) AddElement(pkgID string, spec ElementSpec) (*Element, error) {
	host := d.modelPackageElement(pkgID)
	if host == nil {
		return nil, fmt.Errorf("eaxmi: no package %q", pkgID)
	}
	pkg := d.packageByID[pkgID]
	if pkgID == "" {
		pkg = d.Root
	}

	guid := NewGUID()
	id := xmiIDFromGUID(guid, "EAID_")
	author := spec.Author
	if author == "" {
		author = genAuthor
	}

	// 1. model tree
	pe := host.CreateElement("packagedElement")
	pe.CreateAttr("xmi:type", spec.UMLType)
	pe.CreateAttr("xmi:id", id)
	pe.CreateAttr("name", spec.Name)
	pe.CreateAttr("visibility", "public")

	// 2. extension <elements><element>
	els := d.extensionElementsBlock()
	el := els.CreateElement("element")
	el.CreateAttr("xmi:idref", id)
	el.CreateAttr("xmi:type", spec.UMLType)
	el.CreateAttr("name", spec.Name)
	el.CreateAttr("scope", "public")
	m := el.CreateElement("model")
	m.CreateAttr("package", pkgID)
	m.CreateAttr("tpos", "0")
	m.CreateAttr("ea_localid", "0")
	m.CreateAttr("ea_eleType", "element")
	pr := el.CreateElement("properties")
	if spec.Documentation != "" {
		pr.CreateAttr("documentation", spec.Documentation)
	}
	pr.CreateAttr("isSpecification", "false")
	pr.CreateAttr("sType", strings.TrimPrefix(spec.UMLType, "uml:"))
	pr.CreateAttr("nType", "0")
	pr.CreateAttr("scope", "public")
	if spec.Stereotype != "" {
		pr.CreateAttr("stereotype", spec.Stereotype)
	}
	pr.CreateAttr("isRoot", "false")
	pr.CreateAttr("isLeaf", "false")
	pr.CreateAttr("isAbstract", "false")
	pr.CreateAttr("isActive", "false")
	proj := el.CreateElement("project")
	proj.CreateAttr("author", author)
	proj.CreateAttr("version", "1.0")
	proj.CreateAttr("phase", "1.0")
	proj.CreateAttr("status", "Proposed")
	st := el.CreateElement("style")
	st.CreateAttr("appearance", "BackColor=-1;BorderColor=-1;BorderWidth=1;FontColor=-1;VSwimLanes=1;HSwimLanes=1;BorderStyle=0;")
	ep := el.CreateElement("extendedProperties")
	ep.CreateAttr("tagged", "0")
	if pkg != nil {
		ep.CreateAttr("package_name", pkg.Name)
	}
	el.CreateElement("links")

	// 3. profile application
	if spec.Stereotype != "" {
		d.addProfileApplication(spec.Stereotype, "base_"+strings.TrimPrefix(spec.UMLType, "uml:"), id)
	}

	// 4. update the parsed model
	newEl := &Element{
		XMIID: id, GUID: guid, Name: spec.Name, UMLType: spec.UMLType,
		Stereotype: spec.Stereotype, Documentation: spec.Documentation,
		Author: author, Package: pkg,
	}
	d.elementByID[id] = newEl
	if pkg != nil {
		pkg.Elements = append(pkg.Elements, newEl)
	}
	return newEl, nil
}

// AddConnector creates a relationship. It returns the new Connector.
func (d *Document) AddConnector(spec ConnectorSpec) (*Connector, error) {
	src, ok := d.elementByID[spec.SourceID]
	if !ok {
		return nil, fmt.Errorf("eaxmi: no source element %q", spec.SourceID)
	}
	if _, ok := d.elementByID[spec.TargetID]; !ok {
		return nil, fmt.Errorf("eaxmi: no target element %q", spec.TargetID)
	}
	guid := NewGUID()
	id := xmiIDFromGUID(guid, "EAID_")
	author := spec.Author
	if author == "" {
		author = genAuthor
	}

	hostPkgID := ""
	if src.Package != nil {
		hostPkgID = src.Package.XMIID
	}
	host := d.modelPackageElement(hostPkgID)

	// 1. model tree representation
	switch spec.ModelRepr {
	case "controlflow":
		e := host.CreateElement("edge")
		e.CreateAttr("xmi:type", "uml:"+spec.EAType)
		e.CreateAttr("xmi:id", id)
		if spec.Name != "" {
			e.CreateAttr("name", spec.Name)
		}
		e.CreateAttr("visibility", "public")
		e.CreateAttr("source", spec.SourceID)
		e.CreateAttr("target", spec.TargetID)
	case "dependency":
		pe := host.CreateElement("packagedElement")
		pe.CreateAttr("xmi:type", "uml:Dependency")
		pe.CreateAttr("xmi:id", id)
		if spec.Name != "" {
			pe.CreateAttr("name", spec.Name)
		}
		pe.CreateAttr("visibility", "public")
		pe.CreateAttr("supplier", spec.TargetID)
		pe.CreateAttr("client", spec.SourceID)
	case "generalization":
		srcPE := d.modelPackageElement(hostPkgID).FindElement("packagedElement[@xmi:id='" + spec.SourceID + "']")
		if srcPE == nil {
			return nil, fmt.Errorf("eaxmi: source element node not found for generalization")
		}
		g := srcPE.CreateElement("generalization")
		g.CreateAttr("xmi:type", "uml:Generalization")
		g.CreateAttr("xmi:id", id)
		g.CreateAttr("general", spec.TargetID)
	default: // "association"
		pe := host.CreateElement("packagedElement")
		pe.CreateAttr("xmi:type", "uml:Association")
		pe.CreateAttr("xmi:id", id)
		if spec.Name != "" {
			pe.CreateAttr("name", spec.Name)
		}
		pe.CreateAttr("visibility", "public")
		srcEnd := "EAID_src" + strings.TrimPrefix(id, "EAID_")[2:]
		dstEnd := "EAID_dst" + strings.TrimPrefix(id, "EAID_")[2:]
		pe.CreateElement("memberEnd").CreateAttr("xmi:idref", dstEnd)
		pe.CreateElement("memberEnd").CreateAttr("xmi:idref", srcEnd)
		d.ownedEnd(pe, srcEnd, id, spec.SourceID)
		d.ownedEnd(pe, dstEnd, id, spec.TargetID)
	}

	// 2. extension <connectors><connector>
	cons := d.extensionConnectorsBlock()
	c := cons.CreateElement("connector")
	c.CreateAttr("xmi:idref", id)
	if spec.Name != "" {
		c.CreateAttr("name", spec.Name)
	}
	s := c.CreateElement("source")
	s.CreateAttr("xmi:idref", spec.SourceID)
	d.connEndModel(s, spec.SourceID)
	t := c.CreateElement("target")
	t.CreateAttr("xmi:idref", spec.TargetID)
	d.connEndModel(t, spec.TargetID)
	c.CreateElement("model").CreateAttr("ea_localid", "0")
	p := c.CreateElement("properties")
	p.CreateAttr("ea_type", spec.EAType)
	if spec.Stereotype != "" {
		p.CreateAttr("stereotype", spec.Stereotype)
	}
	if spec.Direction != "" {
		p.CreateAttr("direction", spec.Direction)
	}
	ap := c.CreateElement("appearance")
	ap.CreateAttr("linemode", "3")
	ap.CreateAttr("linecolor", "-1")
	ap.CreateAttr("linewidth", "0")
	ap.CreateAttr("seqno", "0")
	ap.CreateAttr("headStyle", "0")
	ap.CreateAttr("lineStyle", "0")
	if spec.Documentation != "" {
		c.CreateElement("documentation").CreateAttr("value", spec.Documentation)
	}

	// 3. profile application
	if spec.Stereotype != "" {
		d.addProfileApplication(spec.Stereotype, "base_"+spec.EAType, id)
	}

	// 4. source element <links>
	if links := d.extension().FindElement("//element[@xmi:idref='" + spec.SourceID + "']/links"); links != nil {
		l := links.CreateElement(spec.EAType)
		l.CreateAttr("xmi:id", id)
		l.CreateAttr("start", spec.SourceID)
		l.CreateAttr("end", spec.TargetID)
	}

	conn := &Connector{
		XMIID: id, GUID: guid, Name: spec.Name, Documentation: spec.Documentation,
		EAType: spec.EAType, Stereotype: spec.Stereotype, Direction: spec.Direction,
		SourceID: spec.SourceID, TargetID: spec.TargetID,
	}
	d.connectorAll = append(d.connectorAll, conn)
	return conn, nil
}

func (d *Document) ownedEnd(assoc *etree.Element, endID, assocID, typeID string) {
	oe := assoc.CreateElement("ownedEnd")
	oe.CreateAttr("xmi:type", "uml:Property")
	oe.CreateAttr("xmi:id", endID)
	oe.CreateAttr("visibility", "public")
	oe.CreateAttr("association", assocID)
	oe.CreateAttr("aggregation", "none")
	oe.CreateElement("type").CreateAttr("xmi:idref", typeID)
}

func (d *Document) connEndModel(end *etree.Element, elemID string) {
	m := end.CreateElement("model")
	if el, ok := d.elementByID[elemID]; ok {
		m.CreateAttr("type", strings.TrimPrefix(el.UMLType, "uml:"))
		m.CreateAttr("name", el.Name)
	}
	r := end.CreateElement("role")
	r.CreateAttr("visibility", "Public")
	r.CreateAttr("targetScope", "instance")
	end.CreateElement("type").CreateAttr("containment", "Unspecified")
}

// RemoveElement deletes an element and every connector that touches it.
func (d *Document) RemoveElement(id string) error {
	el, ok := d.elementByID[id]
	if !ok {
		return fmt.Errorf("eaxmi: no element %q", id)
	}
	for _, c := range append([]*Connector(nil), d.connectorAll...) {
		if c.SourceID == id || c.TargetID == id {
			_ = d.RemoveConnector(c.XMIID)
		}
	}
	d.deleteByID("packagedElement", id)
	d.deleteExtByIDRef("element", id)
	d.deleteProfileApplication(id)
	delete(d.elementByID, id)
	if el.Package != nil {
		el.Package.Elements = removeElement(el.Package.Elements, el)
	}
	return nil
}

// RemoveConnector deletes a relationship.
func (d *Document) RemoveConnector(id string) error {
	var conn *Connector
	for _, c := range d.connectorAll {
		if c.XMIID == id {
			conn = c
			break
		}
	}
	if conn == nil {
		return fmt.Errorf("eaxmi: no connector %q", id)
	}
	d.deleteByID("packagedElement", id)
	d.deleteByID("edge", id)
	d.deleteByID("generalization", id)
	d.deleteExtByIDRef("connector", id)
	d.deleteProfileApplication(id)
	for _, links := range d.doc.FindElements("//links") {
		for _, l := range links.ChildElements() {
			if l.SelectAttrValue("xmi:id", "") == id {
				links.RemoveChild(l)
			}
		}
	}
	d.connectorAll = removeConnector(d.connectorAll, conn)
	return nil
}

// SetElementName / SetElementDocumentation change one basic property.
func (d *Document) SetElementName(id, name string) error {
	el, ok := d.elementByID[id]
	if !ok {
		return fmt.Errorf("eaxmi: no element %q", id)
	}
	if pe := d.doc.FindElement("//packagedElement[@xmi:id='" + id + "']"); pe != nil {
		pe.CreateAttr("name", name)
	}
	if e := d.extension().FindElement("//element[@xmi:idref='" + id + "']"); e != nil {
		e.CreateAttr("name", name)
	}
	el.Name = name
	return nil
}

func (d *Document) SetElementDocumentation(id, doc string) error {
	el, ok := d.elementByID[id]
	if !ok {
		return fmt.Errorf("eaxmi: no element %q", id)
	}
	e := d.extension().FindElement("//element[@xmi:idref='" + id + "']/properties")
	if e != nil {
		e.CreateAttr("documentation", doc)
	}
	el.Documentation = doc
	return nil
}

// AddDiagramObject places an element on a diagram.
func (d *Document) AddDiagramObject(diagramID, elementID string, left, top, right, bottom int) error {
	g, ok := d.diagramByID[diagramID]
	if !ok {
		return fmt.Errorf("eaxmi: no diagram %q", diagramID)
	}
	if _, ok := d.elementByID[elementID]; !ok {
		return fmt.Errorf("eaxmi: no element %q", elementID)
	}
	dg := d.extension().FindElement("//diagram[@xmi:id='" + diagramID + "']")
	if dg == nil {
		return fmt.Errorf("eaxmi: diagram node %q not found", diagramID)
	}
	els := dg.FindElement("elements")
	if els == nil {
		els = dg.CreateElement("elements")
	}
	seq := len(g.Objects) + 1
	o := els.CreateElement("element")
	o.CreateAttr("geometry", fmt.Sprintf("Left=%d;Top=%d;Right=%d;Bottom=%d;", left, top, right, bottom))
	o.CreateAttr("subject", elementID)
	o.CreateAttr("seqno", fmt.Sprintf("%d", seq))
	o.CreateAttr("style", "DUID="+strings.ToUpper(strings.TrimPrefix(NewGUID(), "{")[:8])+";UCRect=1;HideIcon=0;")

	g.Objects = append(g.Objects, &DiagramObject{
		SubjectID: elementID, Seq: seq,
		Left: left, Top: top, Right: right, Bottom: bottom,
	})
	return nil
}

// RemoveDiagramObject removes an element's placement from a diagram (the element
// itself is untouched).
func (d *Document) RemoveDiagramObject(diagramID, elementID string) error {
	g, ok := d.diagramByID[diagramID]
	if !ok {
		return fmt.Errorf("eaxmi: no diagram %q", diagramID)
	}
	dg := d.extension().FindElement("//diagram[@xmi:id='" + diagramID + "']")
	if dg == nil {
		return fmt.Errorf("eaxmi: diagram node %q not found", diagramID)
	}
	removed := false
	if els := dg.FindElement("elements"); els != nil {
		for _, o := range els.ChildElements() {
			if o.SelectAttrValue("subject", "") == elementID && !strings.HasPrefix(o.SelectAttrValue("geometry", ""), "SX=") {
				els.RemoveChild(o)
				removed = true
			}
		}
	}
	if !removed {
		return fmt.Errorf("eaxmi: element %q is not on diagram %q", elementID, diagramID)
	}
	var kept []*DiagramObject
	for _, o := range g.Objects {
		if o.SubjectID != elementID {
			kept = append(kept, o)
		}
	}
	g.Objects = kept
	return nil
}

// WriteFile serialises the (possibly edited) document to path. etree writes
// UTF-8, so the XML declaration is normalised to UTF-8 (EA reads either, and
// keeping a windows-1252 declaration on UTF-8 bytes would corrupt Cyrillic).
func (d *Document) WriteFile(path string) error {
	for _, tok := range d.doc.Child {
		if pi, ok := tok.(*etree.ProcInst); ok && pi.Target == "xml" {
			pi.Inst = `version="1.0" encoding="UTF-8"`
		}
	}
	d.doc.Indent(2)
	return d.doc.WriteToFile(path)
}

// ---------- low-level helpers ----------

func (d *Document) extensionElementsBlock() *etree.Element {
	ext := d.extension()
	if b := ext.FindElement("elements"); b != nil {
		return b
	}
	return ext.CreateElement("elements")
}

func (d *Document) extensionConnectorsBlock() *etree.Element {
	ext := d.extension()
	if b := ext.FindElement("connectors"); b != nil {
		return b
	}
	return ext.CreateElement("connectors")
}

func (d *Document) addProfileApplication(stereotype, baseAttr, id string) {
	ext := d.extension()
	pa := ext.CreateElement("ArchiMate3:" + stereotype)
	pa.CreateAttr(baseAttr, id)
}

func (d *Document) deleteProfileApplication(id string) {
	for _, pa := range d.extension().ChildElements() {
		if pa.Space != "ArchiMate3" {
			continue
		}
		for _, a := range pa.Attr {
			if a.Value == id {
				d.extension().RemoveChild(pa)
			}
		}
	}
}

func (d *Document) deleteByID(tag, id string) {
	for _, e := range d.doc.FindElements("//" + tag + "[@xmi:id='" + id + "']") {
		if e.Parent() != nil {
			e.Parent().RemoveChild(e)
		}
	}
}

func (d *Document) deleteExtByIDRef(tag, id string) {
	for _, e := range d.extension().FindElements("//" + tag + "[@xmi:idref='" + id + "']") {
		if e.Parent() != nil {
			e.Parent().RemoveChild(e)
		}
	}
}

func removeElement(s []*Element, x *Element) []*Element {
	out := s[:0]
	for _, e := range s {
		if e != x {
			out = append(out, e)
		}
	}
	return out
}

func removeConnector(s []*Connector, x *Connector) []*Connector {
	out := s[:0]
	for _, c := range s {
		if c != x {
			out = append(out, c)
		}
	}
	return out
}
