package sparx

import (
	"sort"
	"strings"
)

// The service speaks ArchiMate: an element has one Type ("ArchiMate.Goal"), not
// a (uml:Class, stereotype) pair. This file is the ArchiMate 3.2 vocabulary,
// the conversion to/from EA's representation, and the relationship rules used to
// validate a mutation before it is written.

type elementTypeName = string

// archimateElements is the full ArchiMate 3.2 element vocabulary — the hard-coded
// whitelist: the service refuses to create anything outside this set.
var archimateElements = toSet(
	// Motivation
	"Stakeholder", "Driver", "Assessment", "Goal", "Outcome",
	"Principle", "Requirement", "Constraint", "Meaning", "Value",
	// Strategy
	"Resource", "Capability", "CourseOfAction", "ValueStream",
	// Business
	"BusinessActor", "BusinessRole", "BusinessCollaboration", "BusinessInterface",
	"BusinessProcess", "BusinessFunction", "BusinessInteraction", "BusinessEvent",
	"BusinessService", "BusinessObject", "Contract", "Representation", "Product",
	// Application
	"ApplicationComponent", "ApplicationCollaboration", "ApplicationInterface",
	"ApplicationFunction", "ApplicationInteraction", "ApplicationProcess",
	"ApplicationEvent", "ApplicationService", "DataObject",
	// Technology
	"Node", "Device", "SystemSoftware", "TechnologyCollaboration",
	"TechnologyInterface", "Path", "CommunicationNetwork", "TechnologyFunction",
	"TechnologyProcess", "TechnologyInteraction", "TechnologyEvent",
	"TechnologyService", "Artifact",
	// Physical
	"Equipment", "Facility", "DistributionNetwork", "Material",
	// Implementation & Migration
	"WorkPackage", "Deliverable", "ImplementationEvent", "Plateau", "Gap",
	// Other
	"Location", "Grouping",
)

// archimateRelationships is the full ArchiMate 3.2 relationship vocabulary.
var archimateRelationships = toSet(
	"Composition", "Aggregation", "Assignment", "Realization", "Serving",
	"Access", "Influence", "Triggering", "Flow", "Specialization", "Association",
)

// behaviorTypes is the ArchiMate behaviour vocabulary — the elements EA exports
// as uml:Activity rather than uml:Class (events included).
var behaviorTypes = toSet(
	"BusinessProcess", "BusinessFunction", "BusinessInteraction", "BusinessEvent", "BusinessService",
	"ApplicationFunction", "ApplicationInteraction", "ApplicationProcess", "ApplicationEvent", "ApplicationService",
	"TechnologyFunction", "TechnologyProcess", "TechnologyInteraction", "TechnologyEvent", "TechnologyService",
	"ImplementationEvent", "ValueStream",
)

// interfaceTypes are the ArchiMate elements EA exports as uml:Interface rather
// than uml:Class: the exposed-behaviour "interface" element of each active
// layer. Confirmed against a real EA export (tmp/examples/sandbox.xml) — EA
// only recognises the ArchiMate_*Interface stereotype on a uml:Interface base
// (base_Interface); applied to a uml:Class it renders as an anonymous class.
var interfaceTypes = toSet("ApplicationInterface", "BusinessInterface", "TechnologyInterface")

// componentTypes are the ArchiMate elements EA exports as uml:Component rather
// than uml:Class.
var componentTypes = toSet("ApplicationComponent")

// ---------- EA representation ----------

// eaRelation is how EA serialises an ArchiMate relationship in an XMI export.
type eaRelation struct {
	EAType    string // <properties ea_type="…">
	ModelRepr string // "association" | "controlflow" | "dependency" | "generalization"
	Direction string // <properties direction="…">
	verified  bool   // confirmed against a real EA export?
}

// relEA maps an ArchiMate relationship type to its EA serialisation. The
// verified entries come from real EA "Export Package to XMI" files
// (example/TestProject*-Model.xml); the rest follow EA's ArchiMate3 MDG
// conventions and are confirmed on the first import round-trip.
var relEA = map[elementTypeName]eaRelation{
	"Association":    {EAType: "Association", ModelRepr: "association", Direction: "Unspecified", verified: true},
	"Composition":    {EAType: "Association", ModelRepr: "association", Direction: "Unspecified", verified: true},
	"Aggregation":    {EAType: "Association", ModelRepr: "association", Direction: "Unspecified", verified: true},
	"Assignment":     {EAType: "Association", ModelRepr: "association", Direction: "Unspecified", verified: false},
	"Influence":      {EAType: "ControlFlow", ModelRepr: "controlflow", Direction: "Source -> Destination", verified: true},
	"Triggering":     {EAType: "ControlFlow", ModelRepr: "controlflow", Direction: "Source -> Destination", verified: false},
	"Flow":           {EAType: "ControlFlow", ModelRepr: "controlflow", Direction: "Source -> Destination", verified: false},
	"Realization":    {EAType: "Dependency", ModelRepr: "dependency", Direction: "Source -> Destination", verified: true},
	"Serving":        {EAType: "Dependency", ModelRepr: "dependency", Direction: "Source -> Destination", verified: false},
	"Access":         {EAType: "Dependency", ModelRepr: "dependency", Direction: "Source -> Destination", verified: false},
	"Specialization": {EAType: "Generalization", ModelRepr: "generalization", Direction: "", verified: true},
}

// EARelationship returns EA's serialisation of a bare ArchiMate relationship
// name ("Realization"): the EA connector type, the model representation and the
// direction. Exported for tooling that writes connectors directly via the eaxmi
// codec (see tools/relexamples).
func EARelationship(relBare string) (eaType, modelRepr, direction string, ok bool) {
	e, ok := relEA[relBare]
	return e.EAType, e.ModelRepr, e.Direction, ok
}

// EAElementForType returns EA's (uml type, stereotype) pair for a bare ArchiMate
// element type ("Goal"). Exported for tooling (see tools/relexamples).
func EAElementForType(typeBare string) (umlType, stereotype string) {
	e := eaForElement(typeBare)
	return e.UMLType, e.Stereotype
}

// eaElement is how EA serialises an ArchiMate element.
type eaElement struct {
	UMLType    string // <packagedElement xmi:type="…">
	Stereotype string // "ArchiMate_Goal"
}

// eaForElement returns EA's (uml type, stereotype) for an ArchiMate type. Most
// ArchiMate elements are uml:Class; behaviour elements are uml:Activity, the
// three ArchiMate "interface" elements are uml:Interface, and
// ApplicationComponent is uml:Component.
func eaForElement(t elementTypeName) eaElement {
	umlType := "uml:Class"
	switch {
	case behaviorTypes[t]:
		umlType = "uml:Activity"
	case interfaceTypes[t]:
		umlType = "uml:Interface"
	case componentTypes[t]:
		umlType = "uml:Component"
	}
	return eaElement{UMLType: umlType, Stereotype: "ArchiMate_" + t}
}

// ---------- name <-> EA stereotype ----------

func archimateTypeFromStereotype(stereotype string) (name string, ok bool) {
	const p = "ArchiMate_"
	if !strings.HasPrefix(stereotype, p) {
		return stereotype, false
	}
	return stereotype[len(p):], true
}

func qualify(name string) string { return "ArchiMate." + name }

func unqualify(qualified string) (string, bool) {
	const p = "ArchiMate."
	if !strings.HasPrefix(qualified, p) {
		return qualified, false
	}
	return qualified[len(p):], true
}

// IsKnownElementType reports whether qualified ("ArchiMate.Goal") is in the
// hard-coded ArchiMate element vocabulary.
func IsKnownElementType(qualified string) bool {
	name, ok := unqualify(qualified)
	return ok && archimateElements[name]
}

// IsKnownRelationshipType reports whether qualified ("ArchiMate.Realization") is
// in the hard-coded ArchiMate relationship vocabulary.
func IsKnownRelationshipType(qualified string) bool {
	name, ok := unqualify(qualified)
	return ok && archimateRelationships[name]
}

// RelationshipAllowed reports whether ea_create_relationship would accept an
// ArchiMate relationship of the qualified type relType ("ArchiMate.Realization")
// from sourceType to targetType — i.e. the verdict is not Deny. Any unknown name
// yields false. See RelationshipVerdict for the three-way answer.
func RelationshipAllowed(relType, sourceType, targetType string) bool {
	return RelationshipVerdictQualified(relType, sourceType, targetType) != VerdictDeny
}

// ElementTypes returns the qualified names of every ArchiMate element type the
// service accepts ("ArchiMate.Goal", …), sorted.
func ElementTypes() []string { return qualifiedSorted(archimateElements) }

// RelationshipTypes returns the qualified names of every ArchiMate relationship
// type the service accepts ("ArchiMate.Realization", …), sorted.
func RelationshipTypes() []string { return qualifiedSorted(archimateRelationships) }

func qualifiedSorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, qualify(name))
	}
	sort.Strings(out)
	return out
}

func toSet(vs ...string) map[string]bool {
	m := make(map[string]bool, len(vs))
	for _, v := range vs {
		m[v] = true
	}
	return m
}
