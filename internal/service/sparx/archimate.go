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

// ---------- element aspect (for relationship rules) ----------

type aspect string

const (
	aspectActive     aspect = "active"   // active structure
	aspectBehavior   aspect = "behavior" // behaviour
	aspectPassive    aspect = "passive"  // passive structure
	aspectMotivation aspect = "motivation"
	aspectStrategy   aspect = "strategy"
	aspectImpl       aspect = "implementation"
	aspectComposite  aspect = "composite" // Location, Grouping
)

var elementAspect = map[elementTypeName]aspect{
	"Stakeholder": aspectMotivation, "Driver": aspectMotivation, "Assessment": aspectMotivation,
	"Goal": aspectMotivation, "Outcome": aspectMotivation, "Principle": aspectMotivation,
	"Requirement": aspectMotivation, "Constraint": aspectMotivation, "Meaning": aspectMotivation,
	"Value": aspectMotivation,

	"Resource": aspectStrategy, "Capability": aspectStrategy,
	"CourseOfAction": aspectStrategy, "ValueStream": aspectStrategy,

	"BusinessActor": aspectActive, "BusinessRole": aspectActive,
	"BusinessCollaboration": aspectActive, "BusinessInterface": aspectActive,
	"BusinessProcess": aspectBehavior, "BusinessFunction": aspectBehavior,
	"BusinessInteraction": aspectBehavior, "BusinessEvent": aspectBehavior,
	"BusinessService": aspectBehavior,
	"BusinessObject":  aspectPassive, "Contract": aspectPassive, "Representation": aspectPassive,
	"Product": aspectPassive,

	"ApplicationComponent": aspectActive, "ApplicationCollaboration": aspectActive,
	"ApplicationInterface": aspectActive,
	"ApplicationFunction":  aspectBehavior, "ApplicationInteraction": aspectBehavior,
	"ApplicationProcess": aspectBehavior, "ApplicationEvent": aspectBehavior,
	"ApplicationService": aspectBehavior,
	"DataObject":         aspectPassive,

	"Node": aspectActive, "Device": aspectActive, "SystemSoftware": aspectActive,
	"TechnologyCollaboration": aspectActive, "TechnologyInterface": aspectActive,
	"Path": aspectActive, "CommunicationNetwork": aspectActive,
	"TechnologyFunction": aspectBehavior, "TechnologyProcess": aspectBehavior,
	"TechnologyInteraction": aspectBehavior, "TechnologyEvent": aspectBehavior,
	"TechnologyService": aspectBehavior,
	"Artifact":          aspectPassive,

	"Equipment": aspectActive, "Facility": aspectActive, "DistributionNetwork": aspectActive,
	"Material": aspectPassive,

	"WorkPackage": aspectImpl, "Deliverable": aspectImpl, "ImplementationEvent": aspectImpl,
	"Plateau": aspectImpl, "Gap": aspectImpl,

	"Location": aspectComposite, "Grouping": aspectComposite,
}

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

// eaElement is how EA serialises an ArchiMate element.
type eaElement struct {
	UMLType    string // <packagedElement xmi:type="…">
	Stereotype string // "ArchiMate_Goal"
}

// eaForElement returns EA's (uml type, stereotype) for an ArchiMate type. Every
// ArchiMate element is uml:Class except behaviour elements, which EA exports as
// uml:Activity.
func eaForElement(t elementTypeName) eaElement {
	umlType := "uml:Class"
	if elementAspect[t] == aspectBehavior {
		umlType = "uml:Activity"
	}
	return eaElement{UMLType: umlType, Stereotype: "ArchiMate_" + t}
}

// ---------- relationship rules ----------

// relationshipAllowed reports whether an ArchiMate relationship of type rel may
// connect a source element of type src to a target of type tgt. Conservative:
// it rejects a combination it is unsure about rather than let an invalid model
// be written (EA re-validates on import as a backstop). All names are bare
// ("Goal", "Realization").
func relationshipAllowed(rel, src, tgt elementTypeName) (bool, string) {
	if !archimateElements[src] {
		return false, "unknown source element type"
	}
	if !archimateElements[tgt] {
		return false, "unknown target element type"
	}
	if !archimateRelationships[rel] {
		return false, "unknown relationship type"
	}

	sa, ta := elementAspect[src], elementAspect[tgt]
	switch rel {
	case "Association":
		return true, "" // ArchiMate's universal fallback
	case "Specialization":
		if src != tgt {
			return false, "specialization only connects two elements of the same type"
		}
		return true, ""
	case "Composition", "Aggregation":
		if src != tgt {
			return false, rel + " between different ArchiMate types is not supported yet"
		}
		return true, ""
	case "Influence":
		if ta != aspectMotivation {
			return false, "influence must target a motivation element"
		}
		return true, ""
	case "Realization":
		// A more concrete element realizes a more abstract one. Within
		// motivation this is common (Requirement/Principle/Outcome realize a
		// Goal); a motivation element does not realize a core element.
		if ta == aspectMotivation {
			return true, ""
		}
		if sa == aspectMotivation {
			return false, "a motivation element only realizes another motivation element"
		}
		return true, ""
	case "Assignment":
		if sa == aspectActive && (ta == aspectBehavior || ta == aspectActive || ta == aspectPassive) {
			return true, ""
		}
		return false, "assignment goes from an active-structure element to a behaviour, interface or object"
	case "Serving":
		if (sa == aspectActive || sa == aspectBehavior) && ta != aspectPassive {
			return true, ""
		}
		return false, "serving goes from an active-structure or behaviour element"
	case "Access":
		if sa == aspectBehavior && ta == aspectPassive {
			return true, ""
		}
		return false, "access goes from a behaviour element to a passive-structure element"
	case "Triggering", "Flow":
		if sa == aspectBehavior && ta == aspectBehavior {
			return true, ""
		}
		return false, rel + " connects two behaviour elements"
	}
	return false, "relationship not permitted between these element types"
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
