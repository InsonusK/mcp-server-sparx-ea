package eaxmi

import "fmt"

// Mutation of packages: create, move (re-parent) and remove (recursively). The
// model-tree <packagedElement xmi:type="uml:Package"> and the matching
// <xmi:Extension>/<elements>/<element ea_eleType="package"> record are kept in
// step, as is the parsed model.

// AddPackage creates a package named name inside parentID and returns it.
func (d *Document) AddPackage(parentID, name string) (*Package, error) {
	parent, ok := d.packageByID[parentID]
	if !ok {
		return nil, fmt.Errorf("eaxmi: no package %q", parentID)
	}
	host := d.doc.FindElement("//packagedElement[@xmi:id='" + parentID + "']")
	if host == nil {
		return nil, fmt.Errorf("eaxmi: package node %q not found", parentID)
	}

	guid := NewGUID()
	id := xmiIDFromGUID(guid, "EAPK_")

	// 1. model tree
	pe := host.CreateElement("packagedElement")
	pe.CreateAttr("xmi:type", "uml:Package")
	pe.CreateAttr("xmi:id", id)
	pe.CreateAttr("name", name)
	pe.CreateAttr("visibility", "public")

	// 2. extension record
	els := d.extensionElementsBlock()
	el := els.CreateElement("element")
	el.CreateAttr("xmi:idref", id)
	el.CreateAttr("xmi:type", "uml:Package")
	el.CreateAttr("name", name)
	el.CreateAttr("scope", "public")
	m := el.CreateElement("model")
	m.CreateAttr("package2", "EAID_"+underscoreBody(guid))
	m.CreateAttr("package", parentID)
	m.CreateAttr("tpos", "0")
	m.CreateAttr("ea_localid", "0")
	m.CreateAttr("ea_eleType", "package")
	pr := el.CreateElement("properties")
	pr.CreateAttr("isSpecification", "false")
	pr.CreateAttr("sType", "Package")
	pr.CreateAttr("nType", "0")
	pr.CreateAttr("scope", "public")
	proj := el.CreateElement("project")
	proj.CreateAttr("author", genAuthor)
	proj.CreateAttr("version", "1.0")
	proj.CreateAttr("phase", "1.0")
	proj.CreateAttr("status", "Proposed")
	st := el.CreateElement("style")
	st.CreateAttr("appearance", "BackColor=-1;BorderColor=-1;BorderWidth=1;FontColor=-1;VSwimLanes=1;HSwimLanes=1;BorderStyle=0;")

	// 3. parsed model
	np := &Package{XMIID: id, GUID: guid, Name: name, Parent: parent}
	parent.Packages = append(parent.Packages, np)
	d.packageByID[id] = np
	return np, nil
}

// MovePackage re-parents pkgID under newParentID.
func (d *Document) MovePackage(pkgID, newParentID string) error {
	pkg, ok := d.packageByID[pkgID]
	if !ok {
		return fmt.Errorf("eaxmi: no package %q", pkgID)
	}
	newParent, ok := d.packageByID[newParentID]
	if !ok {
		return fmt.Errorf("eaxmi: no package %q", newParentID)
	}
	pe := d.doc.FindElement("//packagedElement[@xmi:id='" + pkgID + "']")
	newHost := d.doc.FindElement("//packagedElement[@xmi:id='" + newParentID + "']")
	if pe == nil || newHost == nil {
		return fmt.Errorf("eaxmi: package node not found")
	}
	if p := pe.Parent(); p != nil {
		p.RemoveChild(pe)
	}
	newHost.AddChild(pe)

	if m := d.extension().FindElement("//element[@xmi:idref='" + pkgID + "']/model"); m != nil {
		m.CreateAttr("package", newParentID)
	}

	if pkg.Parent != nil {
		pkg.Parent.Packages = removePackage(pkg.Parent.Packages, pkg)
	}
	newParent.Packages = append(newParent.Packages, pkg)
	pkg.Parent = newParent
	return nil
}

// RemovePackage removes pkgID and everything under it: sub-packages, elements
// (with every relationship attached to them) and diagrams. It returns the number
// of model objects removed, the package itself included.
func (d *Document) RemovePackage(pkgID string) (int, error) {
	pkg, ok := d.packageByID[pkgID]
	if !ok {
		return 0, fmt.Errorf("eaxmi: no package %q", pkgID)
	}

	removed := 0
	var rec func(p *Package)
	rec = func(p *Package) {
		for _, e := range append([]*Element(nil), p.Elements...) {
			if err := d.RemoveElement(e.XMIID); err == nil {
				removed++
			}
		}
		for _, g := range append([]*Diagram(nil), p.Diagrams...) {
			d.removeDiagram(g)
			removed++
		}
		for _, sub := range append([]*Package(nil), p.Packages...) {
			rec(sub)
		}
		d.deleteByID("packagedElement", p.XMIID)
		d.deleteExtByIDRef("element", p.XMIID)
		delete(d.packageByID, p.XMIID)
		removed++
	}
	rec(pkg)

	if pkg.Parent != nil {
		pkg.Parent.Packages = removePackage(pkg.Parent.Packages, pkg)
	}
	return removed, nil
}

// removeDiagram drops a diagram's <xmi:Extension> record and detaches it from
// its package.
func (d *Document) removeDiagram(g *Diagram) {
	for _, dg := range d.extension().FindElements("//diagram[@xmi:id='" + g.XMIID + "']") {
		if p := dg.Parent(); p != nil {
			p.RemoveChild(dg)
		}
	}
	delete(d.diagramByID, g.XMIID)
	if g.Package != nil {
		g.Package.Diagrams = removeDiagramFrom(g.Package.Diagrams, g)
	}
}

func removePackage(s []*Package, x *Package) []*Package {
	out := s[:0]
	for _, p := range s {
		if p != x {
			out = append(out, p)
		}
	}
	return out
}

func removeDiagramFrom(s []*Diagram, x *Diagram) []*Diagram {
	out := s[:0]
	for _, g := range s {
		if g != x {
			out = append(out, g)
		}
	}
	return out
}
