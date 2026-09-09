package eaxmi

// The parsed view of an EA "Export Package to XMI" (XMI 2.1) file. Only the
// pieces this project needs are modelled; the rest of the document is left
// untouched for round-tripping (see the etree document kept on Document).

// Package is a node in the model tree: the root <uml:Model>, or any
// <packagedElement xmi:type="uml:Package">.
type Package struct {
	XMIID     string // e.g. EAPK_91AD8E48_3AE7_49f7_AAAB_72CA99521F53 ("" for the synthetic model root)
	GUID      string // {91AD8E48-3AE7-49f7-AAAB-72CA99521F53}
	Name      string
	EALocalID string // t_package.Package_ID as a string, when present

	Parent   *Package
	Packages []*Package
	Elements []*Element
	Diagrams []*Diagram
}

// Element is a model element — <packagedElement xmi:type="uml:Class"|"uml:Activity"|…>
// enriched from the <xmi:Extension>/<elements> block.
type Element struct {
	XMIID         string
	GUID          string
	Name          string
	UMLType       string // "uml:Class", "uml:Activity", …
	Stereotype    string // "ArchiMate_Goal" ("" if none)
	Documentation string
	EALocalID     string // t_object.Object_ID
	Author        string

	Package *Package
}

// Connector is a relationship — an <edge> or a connector-shaped <packagedElement>
// (uml:Dependency / uml:Association / …), enriched from <xmi:Extension>/<connectors>.
type Connector struct {
	XMIID         string
	GUID          string
	Name          string
	Documentation string
	EAType        string // "ControlFlow", "Dependency", "Association", …
	Stereotype    string // "ArchiMate_Influence" ("" if none)
	Direction     string // "Source -> Destination" etc.
	EALocalID     string // t_connector.Connector_ID

	SourceID string // Element.XMIID of the source
	TargetID string // Element.XMIID of the target
}

// Diagram is a <diagram> from <xmi:Extension>/<diagrams>.
type Diagram struct {
	XMIID         string
	GUID          string
	Name          string
	Type          string // "Logical", …
	Documentation string
	EALocalID     string

	Package *Package
	Objects []*DiagramObject // element placements
	Links   []*DiagramLink   // connector placements
}

// DiagramObject places an element on a diagram, with its rectangle in EA's
// diagram coordinate space (origin top-left, y grows downward).
type DiagramObject struct {
	SubjectID string // Element.XMIID
	Seq       int
	Left      int
	Top       int
	Right     int
	Bottom    int
}

// DiagramLink records that a connector is shown on a diagram (EA auto-routes it;
// no rectangle).
type DiagramLink struct {
	ConnectorID string // Connector.XMIID
}
