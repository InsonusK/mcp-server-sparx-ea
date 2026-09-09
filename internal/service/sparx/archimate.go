package sparx

import "strings"

// The service speaks ArchiMate: an element has one Type ("ArchiMate.Goal"), not
// a (uml:Class, stereotype) pair. This file is the ArchiMate 3.2 vocabulary and
// the conversion to/from EA's representation.

// elementTypeName is an ArchiMate 3.2 element type without the "ArchiMate."
// prefix, e.g. "Goal". Kept as its own type for readability.
type elementTypeName = string

// archimateElements is the full ArchiMate 3.2 element vocabulary. Membership is
// what "hard-coded whitelist" means for elements: the service refuses to create
// anything outside this set. (Junction is a relationship connector, not here.)
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
	"Access", "Influence", "Triggering", "Flow", "Specialization",
	"Association", "Junction",
)

// IsKnownElementType reports whether qualified ("ArchiMate.Goal") is in the
// hard-coded ArchiMate element vocabulary the service will create.
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

// eaStereotypeForElement / eaStereotypeForRelationship: EA names the stereotype
// "ArchiMate_<Type>".
func eaStereotypeForElement(t elementTypeName) string      { return "ArchiMate_" + t }
func eaStereotypeForRelationship(t elementTypeName) string { return "ArchiMate_" + t }

// archimateTypeFromStereotype turns EA's "ArchiMate_Goal" into "Goal". It returns
// ok=false for a stereotype that is not an ArchiMate one (so callers can report
// it as unknown rather than silently mislabel it).
func archimateTypeFromStereotype(stereotype string) (name string, ok bool) {
	const p = "ArchiMate_"
	if !strings.HasPrefix(stereotype, p) {
		return stereotype, false
	}
	return stereotype[len(p):], true
}

// qualify / unqualify move between the wire form ("ArchiMate.Goal") and the bare
// name ("Goal").
func qualify(name string) string { return "ArchiMate." + name }

func unqualify(qualified string) (string, bool) {
	const p = "ArchiMate."
	if !strings.HasPrefix(qualified, p) {
		return qualified, false
	}
	return qualified[len(p):], true
}

func toSet(vs ...string) map[string]bool {
	m := make(map[string]bool, len(vs))
	for _, v := range vs {
		m[v] = true
	}
	return m
}
