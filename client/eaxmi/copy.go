package eaxmi

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/beevik/etree"
)

// CopyPackage deep-copies the package srcPkgID (from src, which may be d itself)
// into d as a child of destParentID. Every id that belongs to the copied subtree
// — packages, elements, connectors, diagrams, association ends, packed <xrefs>
// GUIDs — is regenerated, so the copy is identity-disjoint from the original and
// from everything already in d. References that point OUT of the subtree (a
// connector end or a diagram object that names an element in another package)
// are left untouched, exactly as EA exports them (see example/side1.xml).
//
// Returns the new package.
func (d *Document) CopyPackage(src *Document, srcPkgID, destParentID string) (*Package, error) {
	srcPE := src.doc.FindElement("//packagedElement[@xmi:id='" + srcPkgID + "']")
	if srcPE == nil {
		return nil, fmt.Errorf("eaxmi: no package %q in the source model", srcPkgID)
	}
	destPE := d.doc.FindElement("//packagedElement[@xmi:id='" + destParentID + "']")
	if destPE == nil {
		return nil, fmt.Errorf("eaxmi: no destination package %q", destParentID)
	}

	// 1. ids defined inside the subtree
	ownBody := map[string]bool{}  // "A_B_C_D_E" GUID bodies
	ownEnd := map[string]bool{}   // whole "EAID_src…" / "EAID_dst…" tokens
	ownPkgID := map[string]bool{} // "EAPK_…" of packages in the subtree
	for _, e := range append([]*etree.Element{srcPE}, srcPE.FindElements(".//*")...) {
		id := e.SelectAttrValue("xmi:id", "")
		if id == "" {
			continue
		}
		if isEndToken(id) {
			ownEnd[id] = true
			continue
		}
		if b := guidBody(id); b != "" {
			ownBody[b] = true
			if strings.HasPrefix(id, "EAPK_") ||
				(e.Tag == "packagedElement" && e.SelectAttrValue("xmi:type", "") == "uml:Package") {
				ownPkgID[id] = true
			}
		}
	}

	// diagrams whose owning package is in the subtree
	var srcDiagrams []*etree.Element
	if sExt := src.doc.FindElement("//Extension"); sExt != nil {
		for _, dg := range sExt.FindElements(".//diagram") {
			m := dg.FindElement("model")
			if m == nil {
				continue
			}
			if ownPkgID[m.SelectAttrValue("package", "")] || m.SelectAttrValue("package", "") == srcPkgID {
				srcDiagrams = append(srcDiagrams, dg)
				if b := guidBody(dg.SelectAttrValue("xmi:id", "")); b != "" {
					ownBody[b] = true
				}
			}
		}
	}

	// 2. remap table
	rm := map[string]string{}
	for b := range ownBody {
		nb := underscoreBody(NewGUID())
		rm["EAID_"+b] = "EAID_" + nb
		rm["EAPK_"+b] = "EAPK_" + nb
		rm["{"+dash(b)+"}"] = "{" + dash(nb) + "}"
	}
	for t := range ownEnd {
		rm[t] = xmiIDFromGUID(NewGUID(), "EAID_")
	}

	// 3. model subtree
	newPE, err := remapFragment(srcPE, rm)
	if err != nil {
		return nil, err
	}
	newID := newPE.SelectAttrValue("xmi:id", "")
	destPE.AddChild(newPE)

	// 4. ArchiMate profile applications that reference the subtree
	if sModel := src.doc.FindElement("//Model"); sModel != nil {
		dModel := d.doc.FindElement("//Model")
		for _, pa := range sModel.ChildElements() {
			if pa.Space != "ArchiMate3" {
				continue
			}
			hit := false
			for _, a := range pa.Attr {
				if b := guidBody(a.Value); b != "" && ownBody[b] {
					hit = true
				}
			}
			if !hit {
				continue
			}
			c, err := remapFragment(pa, rm)
			if err != nil {
				return nil, err
			}
			dModel.AddChild(c)
		}
	}

	// 5. <xmi:Extension> records
	dExt := d.doc.FindElement("//Extension")
	sExt := src.doc.FindElement("//Extension")
	if dExt != nil && sExt != nil {
		copyRecs := func(block string, keep func(*etree.Element) bool) error {
			s := sExt.FindElement(block)
			if s == nil {
				return nil
			}
			dst := dExt.FindElement(block)
			if dst == nil {
				dst = dExt.CreateElement(block)
			}
			for _, rec := range s.ChildElements() {
				if !keep(rec) {
					continue
				}
				c, err := remapFragment(rec, rm)
				if err != nil {
					return err
				}
				dst.AddChild(c)
			}
			return nil
		}
		inSubtree := func(rec *etree.Element) bool {
			b := guidBody(rec.SelectAttrValue("xmi:idref", ""))
			return b != "" && ownBody[b]
		}
		if err := copyRecs("elements", inSubtree); err != nil {
			return nil, err
		}
		if err := copyRecs("connectors", inSubtree); err != nil {
			return nil, err
		}
		if dst := dExt.FindElement("diagrams"); dst != nil || len(srcDiagrams) > 0 {
			if dst == nil {
				dst = dExt.CreateElement("diagrams")
			}
			for _, dg := range srcDiagrams {
				c, err := remapFragment(dg, rm)
				if err != nil {
					return nil, err
				}
				dst.AddChild(c)
			}
		}

		// 6. the copied top package now reads as a child of destParentID
		if m := dExt.FindElement(".//element[@xmi:idref='" + newID + "']/model"); m != nil {
			m.CreateAttr("package", destParentID)
		}
		// nothing in the copy is a model root any more — a copied package always
		// lands under a parent. Strip isModel / Recurse from every package record.
		copiedPkgIDs := map[string]bool{newID: true}
		for _, pe := range newPE.FindElements(".//packagedElement[@xmi:type='uml:Package']") {
			copiedPkgIDs[pe.SelectAttrValue("xmi:id", "")] = true
		}
		for id := range copiedPkgIDs {
			if fl := dExt.FindElement(".//element[@xmi:idref='" + id + "']/flags"); fl != nil {
				pf := fl.SelectAttrValue("packageFlags", "")
				pf = strings.ReplaceAll(pf, "isModel=1;", "")
				pf = strings.ReplaceAll(pf, "Recurse=1;", "")
				fl.CreateAttr("packageFlags", pf)
			}
		}
	}

	if err := d.reparse(); err != nil {
		return nil, err
	}
	np, ok := d.packageByID[newID]
	if !ok {
		return nil, fmt.Errorf("eaxmi: copied package %q not found after reparse", newID)
	}
	return np, nil
}

// CopyProfileDefinitions copies the <xmi:Extension>/<profiles> block (the MDG /
// ArchiMate profile *definitions*, not applications) from other into d, unless d
// already has one. EA needs it to recognise the stereotypes an imported package
// carries. It is static, so no id remapping.
func (d *Document) CopyProfileDefinitions(other *Document) error {
	dExt := d.doc.FindElement("//Extension")
	oExt := other.doc.FindElement("//Extension")
	if dExt == nil || oExt == nil {
		return nil
	}
	if dExt.FindElement("profiles") != nil {
		return nil
	}
	src := oExt.FindElement("profiles")
	if src == nil {
		return nil
	}
	dExt.AddChild(src.Copy())
	return d.reparse()
}

// remapFragment serialises el, rewrites every id token per rm (longest key
// first, so "EAID_x" is not clobbered by a shorter prefix), and parses it back.
func remapFragment(el *etree.Element, rm map[string]string) (*etree.Element, error) {
	doc := etree.NewDocument()
	doc.SetRoot(el.Copy())
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("eaxmi: copy serialise: %w", err)
	}
	s := buf.String()
	keys := make([]string, 0, len(rm))
	for k := range rm {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		s = strings.ReplaceAll(s, k, rm[k])
	}
	nd := etree.NewDocument()
	if err := nd.ReadFromString(s); err != nil {
		return nil, fmt.Errorf("eaxmi: copy re-parse: %w", err)
	}
	return nd.Root().Copy(), nil
}

// guidBody returns the "A_B_C_D_E" body of an EAID_/EAPK_ id, or "" if id is not
// GUID-shaped (5 underscore-separated hex groups of 8-4-4-4-12).
func guidBody(id string) string {
	body := strings.TrimPrefix(strings.TrimPrefix(id, "EAID_"), "EAPK_")
	if body == id && !strings.HasPrefix(id, "EAID_") && !strings.HasPrefix(id, "EAPK_") {
		// braced GUID?
		if strings.HasPrefix(id, "{") && strings.HasSuffix(id, "}") {
			body = strings.ReplaceAll(strings.Trim(id, "{}"), "-", "_")
		} else {
			return ""
		}
	}
	parts := strings.Split(body, "_")
	if len(parts) != 5 || len(parts[0]) != 8 {
		return ""
	}
	return body
}

func isEndToken(id string) bool {
	return strings.HasPrefix(id, "EAID_src") || strings.HasPrefix(id, "EAID_dst")
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
