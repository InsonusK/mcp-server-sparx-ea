package eaxmi

import (
	"bytes"
	"fmt"

	"github.com/beevik/etree"
)

// Merge appends other's root packages into d, so several independently-built
// models become one importable file with one top-level package per source.
//
// Everything that belongs to a source package travels with it:
//   - its <packagedElement xmi:type="uml:Package"> subtree in <uml:Model>
//     (nested elements, edges, generalizations and association elements included)
//   - the ArchiMate profile applications that live next to it under <uml:Model>
//     (<ArchiMate3:ArchiMate_Goal base_Class="…"/> …)
//   - its <xmi:Extension> records: <elements>, <connectors>, <diagrams>
//
// The shared wrapper (<uml:Model>, primitive types) is kept from d only. other
// must be identity-disjoint from d and from anything already merged (run
// RemapIdentity on each).
func (d *Document) Merge(other *Document) error {
	dModel := d.doc.FindElement("//Model")
	oModel := other.doc.FindElement("//Model")
	if dModel == nil || oModel == nil {
		return fmt.Errorf("eaxmi: merge: missing <uml:Model>")
	}

	for _, c := range oModel.ChildElements() {
		switch {
		case c.Tag == "packagedElement" && c.SelectAttrValue("xmi:type", "") == "uml:Package" &&
			c.SelectAttrValue("xmi:id", "") != "EAPrimitiveTypesPackage":
			dModel.AddChild(c.Copy())
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
	}

	// re-parse so the model view reflects the merge
	var buf bytes.Buffer
	if _, err := d.doc.WriteTo(&buf); err != nil {
		return fmt.Errorf("eaxmi: merge serialise: %w", err)
	}
	nd := etree.NewDocument()
	nd.ReadSettings = d.doc.ReadSettings
	if err := nd.ReadFromBytes(buf.Bytes()); err != nil {
		return fmt.Errorf("eaxmi: merge re-parse: %w", err)
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
