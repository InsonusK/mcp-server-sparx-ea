package mcpserver

import "github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"

// Model is the slice of the ArchiMate service the sparx tool handlers use.
// *sparx.Service satisfies it; tests supply a fake so the MCP layer can be
// exercised without touching an XMI file.
type Model interface {
	Tree() *sparx.Node
	Element(ref string) (*sparx.ElementInfo, error)
	Package(ref string) (*sparx.PackageInfo, error)
	Diagram(ref string) (*sparx.DiagramInfo, error)

	CreateRootPackage(name string) (*sparx.PackageInfo, error)
	SetRootName(newName string, freshIdentity bool) (string, error)

	CreateElement(parentRef, archimateType, name, documentation string) (*sparx.ElementInfo, error)
	RenameElement(ref, newName string) (*sparx.ElementInfo, error)
	SetElementDocumentation(ref, documentation string) (*sparx.ElementInfo, error)
	MoveElement(ref, newParentRef string) (*sparx.ElementInfo, error)
	DeleteElement(ref string) error

	CreateRelationship(sourceRef, targetRef, archimateRelType, name, documentation string) (*sparx.Relation, error)
	DeleteRelationship(id string) error
	DeleteRelationshipsBetween(sourceRef, targetRef string) (int, error)

	CreatePackage(parentRef, name string) (*sparx.PackageInfo, error)
	RenamePackage(ref, newName string) (*sparx.PackageInfo, error)
	MovePackage(ref, newParentRef string) (*sparx.PackageInfo, error)
	CopyPackage(ref, destParentRef string) (*sparx.PackageInfo, error)
	DeletePackage(ref string, cascadeDelete bool) (int, error)

	CreateDiagram(parentRef, name, layer string) (*sparx.DiagramInfo, error)
	AddToDiagram(diagramRef, elementRef string, at sparx.Rect) error
	MoveOnDiagram(diagramRef, elementRef string, at sparx.Rect) error
	RemoveFromDiagram(diagramRef, elementRef string) error

	Save(path string) error
}

// SparxOpener resolves a model-file path to a Model. Options.Sparx overrides it;
// the default opens the real XMI service.
type SparxOpener func(path string) (Model, error)

func defaultSparxOpener(path string) (Model, error) { return sparx.Open(path) }

// ModelFactory builds a new, empty Model from a root package name.
// Options.NewModel overrides it; the default builds the real XMI service.
type ModelFactory func(rootName string) (Model, error)

func defaultModelFactory(rootName string) (Model, error) {
	m, err := sparx.NewModel(rootName)
	if err != nil {
		return nil, err
	}
	return m, nil
}
