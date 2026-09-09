package eaxmi

import (
	"bytes"
	"fmt"

	"github.com/beevik/etree"
)

// Merge appends other's root packages and their <xmi:Extension> records
// (elements, connectors, diagrams and ArchiMate profile applications) into d,
// so several models become one importable file with one package per source.
//
// other must be identity-disjoint from d and from anything already merged (run
// RemapIdentity on each). The shared parts — <uml:Model> wrapper, primitive
// types, profile definitions — are taken from d only.
func (d *Document) Merge(other *Document) error {
	dModel := d.doc.FindElement("//Model")
	oModel := other.doc.FindElement("//Model")
	dExt := d.doc.FindElement("//Extension")
	oExt := other.doc.FindElement("//Extension")
	if dModel == nil || oModel == nil {
		return fmt.Errorf("eaxmi: merge: missing <uml:Model>")
	}

	for _, pe := range oModel.ChildElements() {
		if pe.Tag == "packagedElement" && pe.SelectAttrValue("xmi:type", "") == "uml:Package" &&
			pe.SelectAttrValue("xmi:id", "") != "EAPrimitiveTypesPackage" {
			dModel.AddChild(pe.Copy())
		}
	}

	if dExt != nil && oExt != nil {
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
		for _, c := range oExt.ChildElements() {
			if c.Space == "ArchiMate3" {
				dExt.AddChild(c.Copy())
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
