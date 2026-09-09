package eaxmi

import "fmt"

// RootPackages returns the names of the EA root packages (the direct children
// of <uml:Model>).
func (d *Document) RootPackages() []string {
	out := make([]string, 0, len(d.Root.Packages))
	for _, p := range d.Root.Packages {
		out = append(out, p.Name)
	}
	return out
}

// SetPackageName renames a package (model-tree name attribute, and the
// package_name back-reference on its direct child elements).
func (d *Document) SetPackageName(id, newName string) error {
	pkg, ok := d.packageByID[id]
	if !ok {
		return fmt.Errorf("eaxmi: no package %q", id)
	}
	old := pkg.Name
	if pe := d.doc.FindElement("//packagedElement[@xmi:id='" + id + "']"); pe != nil {
		pe.CreateAttr("name", newName)
	}
	for _, ep := range d.extension().FindElements("//extendedProperties[@package_name='" + old + "']") {
		ep.CreateAttr("package_name", newName)
	}
	pkg.Name = newName
	return nil
}

// RenameRootPackage gives the single EA root package a new name AND a fresh
// GUID, then repoints every direct child (subpackages, elements, diagrams) and
// its extension records at the new id. A fresh id is what makes an XMI import
// land as a new package next to the original instead of merging into it — used
// by the test suite so many working copies can be imported side by side.
func (d *Document) RenameRootPackage(newName string) error {
	if len(d.Root.Packages) != 1 {
		return fmt.Errorf("eaxmi: expected exactly one root package, found %d", len(d.Root.Packages))
	}
	root := d.Root.Packages[0]
	oldID, oldName := root.XMIID, root.Name

	newGUID := NewGUID()
	newID := xmiIDFromGUID(newGUID, "EAPK_")

	// model tree node
	pe := d.doc.FindElement("//packagedElement[@xmi:id='" + oldID + "']")
	if pe == nil {
		return fmt.Errorf("eaxmi: root package node %q not found", oldID)
	}
	pe.CreateAttr("xmi:id", newID)
	pe.CreateAttr("name", newName)

	// extension <element xmi:idref>
	if el := d.extension().FindElement("//element[@xmi:idref='" + oldID + "']"); el != nil {
		el.CreateAttr("xmi:idref", newID)
	}

	// direct children point their package/owner at the root id
	for _, m := range d.doc.FindElements("//model[@package='" + oldID + "']") {
		m.CreateAttr("package", newID)
	}
	for _, m := range d.doc.FindElements("//model[@owner='" + oldID + "']") {
		m.CreateAttr("owner", newID)
	}
	for _, ep := range d.extension().FindElements("//extendedProperties[@package_name='" + oldName + "']") {
		ep.CreateAttr("package_name", newName)
	}

	// update the parsed view
	delete(d.packageByID, oldID)
	root.XMIID, root.GUID, root.Name = newID, newGUID, newName
	d.packageByID[newID] = root
	return nil
}
