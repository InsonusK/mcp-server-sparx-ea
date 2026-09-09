package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Relationships (method 5): create / delete, with ArchiMate validation. A
// relationship is refused if ArchiMate does not permit it between the two
// element types, or if one of the same (type, source, target) already exists.

// CreateRelationship adds a relationship of ArchiMate type archimateRelType
// (e.g. "ArchiMate.Serving") from sourceRef to targetRef. It fails if the
// relationship is not permitted between the two element types.
func (s *Service) CreateRelationship(sourceRef, targetRef, archimateRelType, name, documentation string) (*Relation, error) {
	relBare, ok := unqualify(archimateRelType)
	if !ok || !archimateRelationships[relBare] {
		return nil, fmt.Errorf("sparx: %q is not a known ArchiMate relationship type", archimateRelType)
	}
	src, _ := s.resolveElement(sourceRef)
	tgt, _ := s.resolveElement(targetRef)
	if src == nil {
		return nil, fmt.Errorf("sparx: no source element for %q", sourceRef)
	}
	if tgt == nil {
		return nil, fmt.Errorf("sparx: no target element for %q", targetRef)
	}
	srcBare, sok := archimateTypeFromStereotype(src.Stereotype)
	tgtBare, tok := archimateTypeFromStereotype(tgt.Stereotype)
	if !sok || !tok {
		return nil, fmt.Errorf("sparx: %s or %s is not an ArchiMate element; refusing to add a relationship", src.Name, tgt.Name)
	}
	if ok, reason := relationshipAllowed(relBare, srcBare, tgtBare); !ok {
		return nil, fmt.Errorf("sparx: %s from %s (%s) to %s (%s) is not allowed: %s",
			relBare, src.Name, qualify(srcBare), tgt.Name, qualify(tgtBare), reason)
	}
	// (type, source, target) is unique: no duplicate relationship.
	stereo := "ArchiMate_" + relBare
	for _, c := range s.doc.Connectors() {
		if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID && c.Stereotype == stereo {
			return nil, fmt.Errorf("sparx: a %s relationship from %s to %s already exists (%s)",
				relBare, src.Name, tgt.Name, c.XMIID)
		}
	}

	ea := relEA[relBare]
	conn, err := s.doc.AddConnector(eaxmi.ConnectorSpec{
		SourceID: src.XMIID, TargetID: tgt.XMIID, Name: strings.TrimSpace(name),
		EAType: ea.EAType, ModelRepr: ea.ModelRepr,
		Stereotype: stereo, Direction: ea.Direction,
		Documentation: documentation,
	})
	if err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return &Relation{
		ID: conn.XMIID, Type: qualify(relBare), Name: conn.Name, Direction: "outgoing",
		OtherID: tgt.XMIID, OtherName: tgt.Name, OtherType: qualify(tgtBare),
	}, nil
}

// DeleteRelationshipsBetween removes every relationship whose source is
// sourceRef and target is targetRef. Returns how many were removed; it is an
// error if there were none.
func (s *Service) DeleteRelationshipsBetween(sourceRef, targetRef string) (int, error) {
	src, _ := s.resolveElement(sourceRef)
	tgt, _ := s.resolveElement(targetRef)
	if src == nil {
		return 0, fmt.Errorf("sparx: no source element for %q", sourceRef)
	}
	if tgt == nil {
		return 0, fmt.Errorf("sparx: no target element for %q", targetRef)
	}
	var ids []string
	for _, c := range s.doc.Connectors() {
		if c.SourceID == src.XMIID && c.TargetID == tgt.XMIID {
			ids = append(ids, c.XMIID)
		}
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("sparx: no relationship from %s to %s", src.Name, tgt.Name)
	}
	for _, id := range ids {
		if err := s.doc.RemoveConnector(id); err != nil {
			return 0, fmt.Errorf("sparx: %w", err)
		}
	}
	return len(ids), nil
}

// DeleteRelationship removes a relationship by its id.
func (s *Service) DeleteRelationship(id string) error {
	for _, c := range s.doc.Connectors() {
		if c.XMIID == id || strings.EqualFold(c.GUID, id) {
			if err := s.doc.RemoveConnector(c.XMIID); err != nil {
				return fmt.Errorf("sparx: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("sparx: no relationship %q", id)
}
