// Command relexamples writes small EA "Import Package from XMI" models that
// exercise the ArchiMate relationship rules:
//
//	docs/examples/relationships-allow.xml — relationships the rules allow
//	docs/examples/relationships-warn.xml  — relationships the rules only warn about
//	docs/examples/relationships-deny.xml  — relationships the rules deny
//	internal/service/sparx/test/testdata/rule_violations.xml — a mixed fixture
//
// A deny model can only be built by writing connectors directly through the
// eaxmi codec (ea_create_relationship would refuse them), so every file is built
// that way. Feed the docs/examples/ models to the MCP server to see
// ea_model_tree / ea_element flag the warn / deny relationships, and try
// ea_create_relationship against them.
//
//	go run ./tools/relexamples
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

type link struct{ src, tgt, rel string }

var files = []struct {
	dir   string
	name  string
	root  string
	want  string // "" = a mixed fixture (per-link verdict not checked against one value)
	links []link
}{
	{
		dir: "internal/service/sparx/test/testdata", name: "rule_violations.xml",
		root: "Rule violations", want: "",
		links: []link{
			{"Goal1", "Node1", "Assignment"},                      // deny
			{"BusinessObject1", "BusinessProcess1", "Triggering"}, // deny
			{"Plateau1", "Plateau2", "Triggering"},                // warn
			{"Node1", "Device1", "Composition"},                   // warn
			{"Requirement1", "Goal1", "Realization"},              // allow
		},
	},
	{
		dir: "docs/examples", name: "relationships-allow.xml", root: "Relationships — allow", want: "allow",
		links: []link{
			{"ApplicationComponent1", "ApplicationService1", "Realization"},
			{"BusinessRole1", "BusinessProcess1", "Assignment"},
			{"BusinessProcess1", "BusinessObject1", "Access"},
			{"Requirement1", "Goal1", "Realization"},
			{"ApplicationService1", "BusinessProcess1", "Serving"},
			{"Stakeholder1", "Goal1", "Association"},
		},
	},
	{
		dir: "docs/examples", name: "relationships-warn.xml", root: "Relationships — warn", want: "warn",
		links: []link{
			{"Plateau1", "Plateau2", "Triggering"},
			{"Node1", "Device1", "Composition"},
			{"Resource1", "Capability1", "Assignment"},
			{"Grouping1", "Goal1", "Aggregation"},
			{"ValueStream1", "Capability1", "Aggregation"},
		},
	},
	{
		dir: "docs/examples", name: "relationships-deny.xml", root: "Relationships — deny", want: "deny",
		links: []link{
			{"Goal1", "Node1", "Assignment"},
			{"BusinessObject1", "BusinessProcess1", "Triggering"},
			{"ApplicationService1", "ApplicationComponent1", "Realization"},
			{"Goal1", "Goal2", "Triggering"},
			{"BusinessProcess1", "BusinessRole1", "Assignment"},
		},
	},
}

// typeOf maps an example element name to its bare ArchiMate type.
func typeOf(name string) string {
	// strip a trailing digit or two
	i := len(name)
	for i > 0 && name[i-1] >= '0' && name[i-1] <= '9' {
		i--
	}
	return name[:i]
}

func main() {
	for _, f := range files {
		if err := os.MkdirAll(f.dir, 0o755); err != nil {
			log.Fatal(err)
		}
		build(filepath.Join(f.dir, f.name), f.root, f.want, f.links)
	}
}

func build(out, root, want string, links []link) {
	doc, err := eaxmi.NewModel(root)
	if err != nil {
		log.Fatal(err)
	}
	p, err := doc.AddPackage(doc.Root.Packages[0].XMIID, "Elements")
	if err != nil {
		log.Fatal(err)
	}

	// collect the element names used and create each once
	seen := map[string]string{} // name -> xmi:id
	add := func(elName string) string {
		if id, ok := seen[elName]; ok {
			return id
		}
		umlType, stereo := sparx.EAElementForType(typeOf(elName))
		el, err := doc.AddElement(p.XMIID, eaxmi.ElementSpec{Name: elName, UMLType: umlType, Stereotype: stereo})
		if err != nil {
			log.Fatalf("%s: add %s: %v", out, elName, err)
		}
		seen[elName] = el.XMIID
		return el.XMIID
	}

	for _, l := range links {
		got := sparx.RelationshipVerdictName("ArchiMate."+l.rel, "ArchiMate."+typeOf(l.src), "ArchiMate."+typeOf(l.tgt))
		if want != "" && got != want {
			log.Fatalf("%s: %s %s->%s is %q, want %q — fix the example list", out, l.rel, l.src, l.tgt, got, want)
		}
		sid, tid := add(l.src), add(l.tgt)
		eaType, repr, dir, ok := sparx.EARelationship(l.rel)
		if !ok {
			log.Fatalf("no EA mapping for %s", l.rel)
		}
		if _, err := doc.AddConnector(eaxmi.ConnectorSpec{
			SourceID: sid, TargetID: tid, EAType: eaType, ModelRepr: repr,
			Stereotype: "ArchiMate_" + l.rel, Direction: dir,
		}); err != nil {
			log.Fatalf("%s: connect %s->%s: %v", out, l.src, l.tgt, err)
		}
	}

	if err := doc.WriteFile(out); err != nil {
		log.Fatal(err)
	}
	label := want
	if label == "" {
		label = "mixed"
	}
	fmt.Printf("wrote %s (%d elements, %d %s relationships)\n", out, len(seen), len(links), label)
}
