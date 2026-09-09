package eaxmi

import (
	"bytes"
	"fmt"

	"github.com/beevik/etree"
)

// MergeUnder nests other's root package into d, as a sub-package of parentID.
// Several independently-built models become one importable file with a single
// root package (EA's XMI import processes one root) whose children are the
// sources.
//
// Everything that belongs to the source package travels with it:
//   - its <packagedElement xmi:type="uml:Package"> subtree in <uml:Model>
//     (nested elements, edges, generalizations and association elements included)
//   - the ArchiMate profile applications next to it under <uml:Model>
//   - its <xmi:Extension> records: <elements>, <connectors>, <diagrams>
//
// The source package's extension record is re-pointed at parentID so it reads
// as a child. other must be identity-disjoint from d and from anything already
// merged (run RemapIdentity on each).
func (d *Document) MergeUnder(other *Document, parentID string) error {
	dModel := d.doc.FindElement("//Model")
	oModel := other.doc.FindElement("//Model")
	host := d.doc.FindElement("//packagedElement[@xmi:id='" + parentID + "']")
	if dModel == nil || oModel == nil {
		return fmt.Errorf("eaxmi: merge: missing <uml:Model>")
	}
	if host == nil {
		return fmt.Errorf("eaxmi: merge: no parent package %q", parentID)
	}

	var nestedRootIDs []string
	for _, c := range oModel.ChildElements() {
		switch {
		case c.Tag == "packagedElement" && c.SelectAttrValue("xmi:type", "") == "uml:Package" &&
			c.SelectAttrValue("xmi:id", "") != "EAPrimitiveTypesPackage":
			host.AddChild(c.Copy())
			nestedRootIDs = append(nestedRootIDs, c.SelectAttrValue("xmi:id", ""))
		case c.Space == "ArchiMate3":
			dModel.AddChild(c.Copy())
		}
	}

	if dExt, oExt := d.doc.FindElement("//Extension"), other.doc.FindElement("//Extension"); dExt != nil && oExt != nil {
		for _, block := range []string{"elements", "connectors", "diagrams"} {
			src := oExt.FindElement(block)
			if src == nil {
				continue
			}
			dst := dExt.FindElement(block)
			if dst == nil {
				dst = dExt.CreateElement(block)
			}
			for _, c := range src.ChildElements() {
				dst.AddChild(c.Copy())
			}
		}
		// re-point the nested source roots at their new parent
		for _, id := range nestedRootIDs {
			if m := dExt.FindElement("//element[@xmi:idref='" + id + "']/model"); m != nil {
				m.CreateAttr("package", parentID)
			}
		}
	}

	return d.reparse()
}

// reparse rebuilds the parsed model view from the (edited) etree document.
func (d *Document) reparse() error {
	var buf bytes.Buffer
	if _, err := d.doc.WriteTo(&buf); err != nil {
		return fmt.Errorf("eaxmi: reparse serialise: %w", err)
	}
	nd := etree.NewDocument()
	nd.ReadSettings = d.doc.ReadSettings
	if err := nd.ReadFromBytes(buf.Bytes()); err != nil {
		return fmt.Errorf("eaxmi: reparse: %w", err)
	}
	np, err := parse(nd)
	if err != nil {
		return err
	}
	path := d.Path
	*d = *np
	d.Path = path
	return nil
}
