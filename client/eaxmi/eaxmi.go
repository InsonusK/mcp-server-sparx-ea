// Package eaxmi reads (and, later, writes) Sparx Enterprise Architect
// "Export Package to XMI" files — EA's XMI 2.1 dialect, in which every element
// appears both in the <uml:Model> semantic tree and in an EA-specific
// <xmi:Extension> block.
//
// This package is a codec: it exposes the model as plain Go structs (see
// model.go) and keeps the underlying document so edits can be written back
// without losing the parts it does not model. It knows nothing about ArchiMate
// — that lives in internal/service/sparx.
package eaxmi

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/beevik/etree"
	"golang.org/x/text/encoding/charmap"
)

// Document is a parsed XMI file.
type Document struct {
	Path            string
	ExporterVersion string

	// Root is the <uml:Model> node as a Package (its XMIID is "").
	Root *Package

	doc          *etree.Document // the full document, for round-tripping
	elementByID  map[string]*Element
	connectorAll []*Connector
	diagramByID  map[string]*Diagram
	packageByID  map[string]*Package
}

// Open parses the XMI file at path.
func Open(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("eaxmi: %w", err)
	}
	defer f.Close()
	d, err := Read(f)
	if err != nil {
		return nil, err
	}
	d.Path = path
	return d, nil
}

// Read parses XMI from r.
func Read(r io.Reader) (*Document, error) {
	doc := etree.NewDocument()
	doc.ReadSettings = etree.ReadSettings{
		CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
			switch strings.ToLower(charset) {
			case "windows-1252", "cp1252":
				return charmap.Windows1252.NewDecoder().Reader(input), nil
			case "windows-1251", "cp1251":
				return charmap.Windows1251.NewDecoder().Reader(input), nil
			case "utf-8", "":
				return input, nil
			default:
				return nil, fmt.Errorf("eaxmi: unsupported charset %q", charset)
			}
		},
	}
	if _, err := doc.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("eaxmi: parse: %w", err)
	}
	return parse(doc)
}

// Connectors returns every relationship in the document.
func (d *Document) Connectors() []*Connector { return d.connectorAll }

// ElementByID looks up an element by its xmi:id (e.g. "EAID_..."). For a lookup
// by id, GUID or path, use ResolveElement (navigate.go).
func (d *Document) ElementByID(id string) (*Element, bool) {
	e, ok := d.elementByID[id]
	return e, ok
}

// guidFromXMIID converts an EA xmi:id ("EAID_A_B_C_D_E" / "EAPK_…") to the
// braced GUID EA stores in the database ("{A-B-C-D-E}"). A value that does not
// match the shape is returned unchanged.
func guidFromXMIID(id string) string {
	body := id
	for _, p := range []string{"EAID_", "EAPK_", "EAID", "EAPK"} {
		if strings.HasPrefix(body, p) {
			body = body[len(p):]
			break
		}
	}
	parts := strings.Split(body, "_")
	if len(parts) != 5 {
		return id
	}
	return "{" + strings.Join(parts, "-") + "}"
}
