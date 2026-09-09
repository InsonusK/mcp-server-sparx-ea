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
// package_name back-reference on its child elements).
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

// RenameRoot renames the single EA root package. With freshIdentity=true it
// first regenerates every GUID in the model (RemapIdentity), so the Saved copy
// imports into EA as a fully independent package next to the original;
// otherwise only the root package gets a fresh id.
func (d *Document) RenameRoot(newName string, freshIdentity bool) error {
	if len(d.Root.Packages) != 1 {
		return fmt.Errorf("eaxmi: expected exactly one root package, found %d", len(d.Root.Packages))
	}
	if freshIdentity {
		if err := d.RemapIdentity(); err != nil {
			return err
		}
		return d.SetPackageName(d.Root.Packages[0].XMIID, newName)
	}
	return d.renameRootWithFreshRootID(newName)
}

// renameRootWithFreshRootID gives just the root package a new name and id and
// repoints its direct children.
func (d *Document) renameRootWithFreshRootID(newName string) error {
	root := d.Root.Packages[0]
	oldID, oldName := root.XMIID, root.Name

	newGUID := NewGUID()
	newID := xmiIDFromGUID(newGUID, "EAPK_")
	oldBody := underscoreBody(root.GUID)

	pe := d.doc.FindElement("//packagedElement[@xmi:id='" + oldID + "']")
	if pe == nil {
		return fmt.Errorf("eaxmi: root package node %q not found", oldID)
	}
	pe.CreateAttr("xmi:id", newID)
	pe.CreateAttr("name", newName)

	if el := d.extension().FindElement("//element[@xmi:idref='" + oldID + "']"); el != nil {
		el.CreateAttr("xmi:idref", newID)
	}
	// direct children reference the root by id in several attributes
	for _, attr := range []string{"package", "package2", "owner"} {
		for _, m := range d.doc.FindElements("//*[@" + attr + "='" + oldID + "']") {
			m.CreateAttr(attr, newID)
		}
		for _, m := range d.doc.FindElements("//*[@" + attr + "='EAID_" + oldBody + "']") {
			m.CreateAttr(attr, "EAID_"+underscoreBody(newGUID))
		}
	}
	for _, ep := range d.extension().FindElements("//extendedProperties[@package_name='" + oldName + "']") {
		ep.CreateAttr("package_name", newName)
	}

	delete(d.packageByID, oldID)
	root.XMIID, root.GUID, root.Name = newID, newGUID, newName
	d.packageByID[newID] = root
	return nil
}
