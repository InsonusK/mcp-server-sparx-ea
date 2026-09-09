package eaxmi

import (
	"fmt"
	"strings"
)

// NewModel builds a valid EA "Export Package to XMI" document from scratch — not
// a copy of a fixture. It contains one root package named rootName (marked as an
// EA model root), ready for AddPackage / AddElement. Save writes it like any
// working copy.
//
// It has no ArchiMate profile block; plain packages and UML elements work as-is,
// ArchiMate elements would need the profile definitions added.
func NewModel(rootName string) (*Document, error) {
	rootName = strings.TrimSpace(rootName)
	if rootName == "" {
		return nil, fmt.Errorf("eaxmi: root name is required")
	}
	body := underscoreBody(NewGUID())
	n := escapeXMLAttr(rootName)
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<xmi:XMI xmi:version="2.1" xmlns:uml="http://schema.omg.org/spec/UML/2.1" xmlns:xmi="http://schema.omg.org/spec/XMI/2.1" xmlns:ArchiMate3="http://www.sparxsystems.com/profiles/ArchiMate3/1.0">
	<xmi:Documentation exporter="Enterprise Architect" exporterVersion="6.5"/>
	<uml:Model xmi:type="uml:Model" name="EA_Model" visibility="public">
		<packagedElement xmi:type="uml:Package" xmi:id="EAPK_` + body + `" name="` + n + `" visibility="public"/>
	</uml:Model>
	<xmi:Extension extender="Enterprise Architect" extenderID="6.5">
		<elements>
			<element xmi:idref="EAPK_` + body + `" xmi:type="uml:Package" name="` + n + `" scope="public">
				<model package2="EAID_` + body + `" tpos="0" ea_localid="1" ea_eleType="package"/>
				<properties isSpecification="false" sType="Package" nType="0" scope="public"/>
				<project author="` + genAuthor + `" version="1.0" phase="1.0" status="Proposed"/>
				<style appearance="BackColor=-1;BorderColor=-1;BorderWidth=-1;FontColor=-1;VSwimLanes=1;HSwimLanes=1;BorderStyle=0;"/>
				<extendedProperties tagged="0"/>
				<flags iscontrolled="FALSE" isprotected="FALSE" batchsave="1" batchload="1" usedtd="FALSE" logxml="FALSE" packageFlags="Recurse=1;isModel=1;"/>
			</element>
		</elements>
		<connectors/>
		<primitivetypes>
			<packagedElement xmi:type="uml:Package" xmi:id="EAPrimitiveTypesPackage" name="EA_PrimitiveTypes_Package" visibility="public"/>
		</primitivetypes>
		<diagrams/>
	</xmi:Extension>
</xmi:XMI>`
	return Read(strings.NewReader(xml))
}

// AddRootPackage adds a package named name directly under <uml:Model> — a new
// top-level package alongside the model root (not marked as a model). Returns it.
func (d *Document) AddRootPackage(name string) (*Package, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("eaxmi: package name is required")
	}
	for _, p := range d.Root.Packages {
		if p.Name == name {
			return nil, fmt.Errorf("eaxmi: a root package named %q already exists", name)
		}
	}
	umlModel := d.doc.FindElement("//Model")
	if umlModel == nil {
		return nil, fmt.Errorf("eaxmi: no <uml:Model>")
	}
	guid := NewGUID()
	id := xmiIDFromGUID(guid, "EAPK_")

	pe := umlModel.CreateElement("packagedElement")
	pe.CreateAttr("xmi:type", "uml:Package")
	pe.CreateAttr("xmi:id", id)
	pe.CreateAttr("name", name)
	pe.CreateAttr("visibility", "public")

	el := d.extensionElementsBlock().CreateElement("element")
	el.CreateAttr("xmi:idref", id)
	el.CreateAttr("xmi:type", "uml:Package")
	el.CreateAttr("name", name)
	el.CreateAttr("scope", "public")
	m := el.CreateElement("model")
	m.CreateAttr("package2", "EAID_"+underscoreBody(guid))
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
	el.CreateElement("style").CreateAttr("appearance", "BackColor=-1;BorderColor=-1;BorderWidth=-1;FontColor=-1;VSwimLanes=1;HSwimLanes=1;BorderStyle=0;")
	el.CreateElement("extendedProperties").CreateAttr("tagged", "0")

	np := &Package{XMIID: id, GUID: guid, Name: name, Parent: d.Root}
	d.Root.Packages = append(d.Root.Packages, np)
	d.packageByID[id] = np
	return np, nil
}

func escapeXMLAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
