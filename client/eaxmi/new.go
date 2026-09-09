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

func escapeXMLAttr(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
