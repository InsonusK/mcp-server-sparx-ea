package sparx

import (
	"fmt"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Repairing elements whose EA base type doesn't match their ArchiMate type —
// the class of defect fixed in eaForElement (archimate.go): an element
// created before that fix still carries the wrong uml:Class/uml:Interface/…
// base, so EA renders it as an anonymous class instead of its real ArchiMate
// shape. Validate reports these without touching the file; Fix repairs the
// ones it's given; ValidateAndFix does both in one step. All three return the
// same EAIssue shape.

// EAIssue is one ArchiMate element whose EA base type doesn't match what
// eaForElement expects for its ArchiMate type.
type EAIssue struct {
	ID       string `json:"id"`
	GUID     string `json:"guid"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`     // "ArchiMate.TechnologyInterface"
	GotType  string `json:"gotType"`  // "uml:Class" — EA's current base type
	WantType string `json:"wantType"` // "uml:Interface" — EA's correct base type
}

// eaIssueFor reports the issue for el, or ok=false if it's fine, or not a
// recognised ArchiMate element (out of scope: there's no expected
// representation to compare against).
func (s *Service) eaIssueFor(el *eaxmi.Element) (EAIssue, bool) {
	bare, ok := archimateTypeFromStereotype(el.Stereotype)
	if !ok || !archimateElements[bare] {
		return EAIssue{}, false
	}
	want := eaForElement(bare)
	if el.UMLType == want.UMLType {
		return EAIssue{}, false
	}
	return EAIssue{
		ID: el.XMIID, GUID: el.GUID, Name: el.Name,
		Path: s.doc.PathOfElement(el), Type: qualify(bare),
		GotType: el.UMLType, WantType: want.UMLType,
	}, true
}

// ValidateModel reports every element whose EA base type doesn't match its
// ArchiMate type. Read-only.
func (s *Service) ValidateModel() []EAIssue {
	issues := []EAIssue{}
	for _, el := range s.doc.AllElements() {
		if issue, ok := s.eaIssueFor(el); ok {
			issues = append(issues, issue)
		}
	}
	return issues
}

// FixElements repairs the EA representation of the given elements (ref: id,
// GUID or path — normally the ids ValidateModel reported). A ref that
// doesn't resolve, or that turns out to need no fix, is silently skipped.
// Returns the elements actually changed.
func (s *Service) FixElements(refs []string) ([]EAIssue, error) {
	fixed := []EAIssue{}
	for _, ref := range refs {
		el, _ := s.resolveElement(ref)
		if el == nil {
			continue
		}
		issue, ok := s.eaIssueFor(el)
		if !ok {
			continue
		}
		if err := s.doc.RetypeElement(el.XMIID, issue.WantType); err != nil {
			return fixed, fmt.Errorf("sparx: %w", err)
		}
		fixed = append(fixed, issue)
	}
	return fixed, nil
}

// ValidateAndFixModel repairs every element ValidateModel finds.
func (s *Service) ValidateAndFixModel() ([]EAIssue, error) {
	issues := s.ValidateModel()
	refs := make([]string, len(issues))
	for i, iss := range issues {
		refs[i] = iss.ID
	}
	return s.FixElements(refs)
}
