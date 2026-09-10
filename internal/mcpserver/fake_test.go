package mcpserver_test

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

// fakeModel is a spy + stub for mcpserver.Model. Tool tests configure its canned
// returns and assert which method a tool called, with which arguments — the MCP
// layer is exercised without touching an XMI file.
type fakeModel struct {
	openedWith string
	calls      []string
	savedTo    string

	err     error // returned by every read/mutate method when set
	saveErr error

	tree     *sparx.Node
	element  *sparx.ElementInfo
	pkg      *sparx.PackageInfo
	diagram  *sparx.DiagramInfo
	relation *sparx.Relation
	count    int
	rootName string
}

func defaultFake() *fakeModel {
	return &fakeModel{
		tree: &sparx.Node{Kind: sparx.KindPackage, Name: "Model", Path: "", Children: []*sparx.Node{
			{Kind: sparx.KindPackage, Name: "Motivation", ID: "EAPK_1", Path: "Model/Motivation"},
		}},
		element: &sparx.ElementInfo{
			ID: "EAID_1", GUID: "{1}", Name: "Goal1", Type: "ArchiMate.Goal",
			Path: "Model/Motivation/Goal1", Documentation: "an aim", Relations: []sparx.Relation{},
		},
		pkg: &sparx.PackageInfo{
			ID: "EAPK_1", GUID: "{2}", Name: "Motivation", Path: "Model/Motivation",
			Packages: []string{}, Elements: []string{"Goal1"}, Diagrams: []string{},
		},
		diagram: &sparx.DiagramInfo{
			ID: "EAID_D", GUID: "{3}", Name: "Overview", Path: "Model/Motivation/Overview",
			DiagramType: "Logical", Objects: []sparx.PlacedElement{}, Links: []sparx.PlacedLink{},
		},
		relation: &sparx.Relation{
			ID: "EAID_R", Type: "ArchiMate.Realization", Direction: "outgoing", OtherName: "Goal1",
			Verdict: "warn", Warning: "Realization from ArchiMate.X to ArchiMate.Y is discouraged by the ArchiMate rules",
		},
		count:    2,
		rootName: "Model",
	}
}

func (f *fakeModel) rec(format string, a ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, a...))
}

func (f *fakeModel) called(sub string) bool {
	for _, c := range f.calls {
		if containsSub(c, sub) {
			return true
		}
	}
	return false
}

func containsSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return len(sub) == 0
}

func (f *fakeModel) Tree() *sparx.Node { f.rec("Tree()"); return f.tree }

func (f *fakeModel) CreateRootPackage(name string) (*sparx.PackageInfo, error) {
	f.rec("CreateRootPackage(%s)", name)
	return f.pkg, f.err
}
func (f *fakeModel) SetRootName(newName string, freshIdentity bool) (string, error) {
	f.rec("SetRootName(%s,%v)", newName, freshIdentity)
	if f.err != nil {
		return "", f.err
	}
	f.rootName = newName
	return newName, nil
}

func (f *fakeModel) Element(ref string) (*sparx.ElementInfo, error) {
	f.rec("Element(%s)", ref)
	return f.element, f.err
}
func (f *fakeModel) Package(ref string) (*sparx.PackageInfo, error) {
	f.rec("Package(%s)", ref)
	return f.pkg, f.err
}
func (f *fakeModel) Diagram(ref string) (*sparx.DiagramInfo, error) {
	f.rec("Diagram(%s)", ref)
	return f.diagram, f.err
}

func (f *fakeModel) CreateElement(parent, typ, name, doc string) (*sparx.ElementInfo, error) {
	f.rec("CreateElement(%s,%s,%s,%s)", parent, typ, name, doc)
	return f.element, f.err
}
func (f *fakeModel) RenameElement(ref, name string) (*sparx.ElementInfo, error) {
	f.rec("RenameElement(%s,%s)", ref, name)
	return f.element, f.err
}
func (f *fakeModel) SetElementDocumentation(ref, doc string) (*sparx.ElementInfo, error) {
	f.rec("SetElementDocumentation(%s,%s)", ref, doc)
	return f.element, f.err
}
func (f *fakeModel) MoveElement(ref, parent string) (*sparx.ElementInfo, error) {
	f.rec("MoveElement(%s,%s)", ref, parent)
	return f.element, f.err
}
func (f *fakeModel) DeleteElement(ref string) error {
	f.rec("DeleteElement(%s)", ref)
	return f.err
}

func (f *fakeModel) CreateRelationship(src, tgt, typ, name, doc string) (*sparx.Relation, error) {
	f.rec("CreateRelationship(%s,%s,%s,%s,%s)", src, tgt, typ, name, doc)
	return f.relation, f.err
}
func (f *fakeModel) DeleteRelationship(id string) error {
	f.rec("DeleteRelationship(%s)", id)
	return f.err
}
func (f *fakeModel) DeleteRelationshipsBetween(src, tgt string) (int, error) {
	f.rec("DeleteRelationshipsBetween(%s,%s)", src, tgt)
	return f.count, f.err
}

func (f *fakeModel) CreatePackage(parent, name string) (*sparx.PackageInfo, error) {
	f.rec("CreatePackage(%s,%s)", parent, name)
	return f.pkg, f.err
}
func (f *fakeModel) RenamePackage(ref, name string) (*sparx.PackageInfo, error) {
	f.rec("RenamePackage(%s,%s)", ref, name)
	return f.pkg, f.err
}
func (f *fakeModel) MovePackage(ref, parent string) (*sparx.PackageInfo, error) {
	f.rec("MovePackage(%s,%s)", ref, parent)
	return f.pkg, f.err
}
func (f *fakeModel) CopyPackage(ref, dest string) (*sparx.PackageInfo, error) {
	f.rec("CopyPackage(%s,%s)", ref, dest)
	return f.pkg, f.err
}
func (f *fakeModel) DeletePackage(ref string, cascade bool) (int, error) {
	f.rec("DeletePackage(%s,%v)", ref, cascade)
	return f.count, f.err
}

func (f *fakeModel) CreateDiagram(parent, name, layer string) (*sparx.DiagramInfo, error) {
	f.rec("CreateDiagram(%s,%s,%s)", parent, name, layer)
	return f.diagram, f.err
}

func (f *fakeModel) AddToDiagram(d, e string, at sparx.Rect) error {
	f.rec("AddToDiagram(%s,%s,%v)", d, e, at)
	return f.err
}
func (f *fakeModel) MoveOnDiagram(d, e string, at sparx.Rect) error {
	f.rec("MoveOnDiagram(%s,%s,%v)", d, e, at)
	return f.err
}
func (f *fakeModel) RemoveFromDiagram(d, e string) error {
	f.rec("RemoveFromDiagram(%s,%s)", d, e)
	return f.err
}

func (f *fakeModel) Save(path string) error {
	f.rec("Save(%s)", path)
	f.savedTo = path
	return f.saveErr
}
