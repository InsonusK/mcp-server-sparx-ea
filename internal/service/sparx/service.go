// Package sparx is the ArchiMate-facing service over an EA XMI file. It gives
// the MCP server simple, validated operations instead of raw SQL / raw XML:
// element types are ArchiMate types (the uml:Class + stereotype pair is hidden),
// and every mutation is validated before it touches the file.
//
// It reads (and, iteratively, writes) EA "Export Package to XMI" files via
// client/eaxmi.
//
// The API is split by model concept, one file each:
//
//	tree.go          the model navigator tree
//	rootnode.go      the EA root package (name / identity)
//	package.go       packages: read + create / rename / move / delete
//	element.go       elements: read + create / rename / re-note / move / delete
//	relationship.go  relationships: create / delete, with ArchiMate validation
//	diagram.go       diagrams: read + place / move / remove an element
//	save.go          writing the working copy back out
//	resolve.go       ref (id / GUID / path) -> model object
//	archimate.go     the ArchiMate 3.2 vocabulary and relationship rules
package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Service is bound to one XMI file.
type Service struct {
	doc *eaxmi.Document
}

// Open loads the XMI file at path.
func Open(path string) (*Service, error) {
	doc, err := eaxmi.Open(path)
	if err != nil {
		return nil, err
	}
	return &Service{doc: doc}, nil
}

// ---------- shared helpers ----------

func elementType(e *eaxmi.Element) string {
	name, ok := archimateTypeFromStereotype(e.Stereotype)
	if !ok || !archimateElements[name] {
		if e.Stereotype == "" {
			return "unknown:" + strings.TrimPrefix(e.UMLType, "uml:")
		}
		return "unknown:" + e.Stereotype
	}
	return qualify(name)
}

func relationType(c *eaxmi.Connector) string {
	name, ok := archimateTypeFromStereotype(c.Stereotype)
	if !ok || !archimateRelationships[name] {
		if c.Stereotype == "" {
			return "unknown:" + c.EAType
		}
		return "unknown:" + c.Stereotype
	}
	return qualify(name)
}

func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

func errWrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("sparx: %w", err)
}
