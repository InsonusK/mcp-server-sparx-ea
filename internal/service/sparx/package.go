package sparx

import (
	"fmt"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eaxmi"
)

// Packages: read a package's contents, and create / rename / move / delete a
// package. Packages are the model's containers, so they get their own method
// set even where the shape overlaps with elements.

// PackageInfo is the answer to "tell me about this package".
type PackageInfo struct {
	ID       string   `json:"id"`
	GUID     string   `json:"guid"`
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Parent   string   `json:"parent,omitempty"` // path of the parent package ("" for a root package)
	Packages []string `json:"packages"`         // child package names
	Elements []string `json:"elements"`         // child element names
	Diagrams []string `json:"diagrams"`         // child diagram names
}

// Package resolves ref (an id, a GUID, or a slash path) and returns its contents.
func (s *Service) Package(ref string) (*PackageInfo, error) {
	p := s.resolvePackage(ref)
	if p == nil {
		return nil, fmt.Errorf("sparx: no package for %q", ref)
	}
	info := &PackageInfo{
		ID: p.XMIID, GUID: p.GUID, Name: p.Name, Path: s.doc.PathOfPackage(p),
		Packages: []string{}, Elements: []string{}, Diagrams: []string{},
	}
	if p.Parent != nil && p.Parent != s.doc.Root {
		info.Parent = s.doc.PathOfPackage(p.Parent)
	}
	for _, sub := range p.Packages {
		info.Packages = append(info.Packages, sub.Name)
	}
	for _, e := range p.Elements {
		info.Elements = append(info.Elements, e.Name)
	}
	for _, g := range p.Diagrams {
		info.Diagrams = append(info.Diagrams, g.Name)
	}
	return info, nil
}

// CreatePackage adds a package named name inside the package at parentRef.
func (s *Service) CreatePackage(parentRef, name string) (*PackageInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("sparx: package name is required")
	}
	parent := s.resolvePackage(parentRef)
	if parent == nil {
		return nil, fmt.Errorf("sparx: no package for %q", parentRef)
	}
	if childPackage(parent, name) != nil {
		return nil, fmt.Errorf("sparx: package %q already has a sub-package named %q", parent.Name, name)
	}
	p, err := s.doc.AddPackage(parent.XMIID, name)
	if err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Package(p.XMIID)
}

// RenamePackage renames a package (by id or path). It keeps the package identity
// (use SetRootName to rename the EA root package with a fresh identity).
func (s *Service) RenamePackage(ref, newName string) (*PackageInfo, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return nil, fmt.Errorf("sparx: new name is required")
	}
	p := s.resolvePackage(ref)
	if p == nil {
		return nil, fmt.Errorf("sparx: no package for %q", ref)
	}
	if p.Parent != nil {
		if other := childPackage(p.Parent, newName); other != nil && other != p {
			return nil, fmt.Errorf("sparx: package %q already has a sub-package named %q", p.Parent.Name, newName)
		}
	}
	if err := s.doc.SetPackageName(p.XMIID, newName); err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Package(p.XMIID)
}

// MovePackage re-parents a package into the package at newParentRef.
func (s *Service) MovePackage(ref, newParentRef string) (*PackageInfo, error) {
	p := s.resolvePackage(ref)
	if p == nil {
		return nil, fmt.Errorf("sparx: no package for %q", ref)
	}
	parent := s.resolvePackage(newParentRef)
	if parent == nil {
		return nil, fmt.Errorf("sparx: no package for %q", newParentRef)
	}
	if p == parent {
		return nil, fmt.Errorf("sparx: cannot move a package into itself")
	}
	for a := parent; a != nil; a = a.Parent {
		if a == p {
			return nil, fmt.Errorf("sparx: cannot move a package into its own descendant")
		}
	}
	if p.Parent == parent {
		return s.Package(p.XMIID)
	}
	if other := childPackage(parent, p.Name); other != nil {
		return nil, fmt.Errorf("sparx: package %q already has a sub-package named %q", parent.Name, p.Name)
	}
	if err := s.doc.MovePackage(p.XMIID, parent.XMIID); err != nil {
		return nil, fmt.Errorf("sparx: %w", err)
	}
	return s.Package(p.XMIID)
}

// CopyPackage deep-copies a package (with a fresh identity, so nothing collides)
// into the package at destParentRef — anywhere in the same model, including
// under a different root package. Returns the new package.
func (s *Service) CopyPackage(ref, destParentRef string) (*PackageInfo, error) {
	src := s.resolvePackage(ref)
	if src == nil {
		return nil, fmt.Errorf("sparx: no package for %q", ref)
	}
	dest := s.resolvePackage(destParentRef)
	if dest == nil {
		return nil, fmt.Errorf("sparx: no package for %q", destParentRef)
	}
	for a := dest; a != nil; a = a.Parent {
		if a == src {
			return nil, fmt.Errorf("sparx: cannot copy a package into itself or its own descendant")
		}
	}
	if other := childPackage(dest, src.Name); other != nil {
		return nil, fmt.Errorf("sparx: package %q already has a sub-package named %q", dest.Name, src.Name)
	}
	np, err := s.doc.CopyPackage(s.doc, src.XMIID, dest.XMIID)
	if err != nil {
		return nil, errWrap(err)
	}
	return s.Package(np.XMIID)
}

// DeletePackage removes a package. With cascadeDelete=false it refuses a package
// that still contains sub-packages, elements or diagrams. With cascadeDelete=true
// it removes the whole subtree (and every relationship attached to a deleted
// element). It never deletes an EA root package. Returns the number of model
// objects removed (the package itself included).
func (s *Service) DeletePackage(ref string, cascadeDelete bool) (int, error) {
	p := s.resolvePackage(ref)
	if p == nil {
		return 0, fmt.Errorf("sparx: no package for %q", ref)
	}
	if p.Parent == nil || p.Parent == s.doc.Root {
		return 0, fmt.Errorf("sparx: %q is an EA root package; rename it instead of deleting", p.Name)
	}
	if !cascadeDelete && (len(p.Packages) > 0 || len(p.Elements) > 0 || len(p.Diagrams) > 0) {
		return 0, fmt.Errorf("sparx: package %q is not empty; pass cascadeDelete to remove its contents", p.Name)
	}
	n, err := s.doc.RemovePackage(p.XMIID)
	if err != nil {
		return 0, fmt.Errorf("sparx: %w", err)
	}
	return n, nil
}

func childPackage(p *eaxmi.Package, name string) *eaxmi.Package {
	for _, sub := range p.Packages {
		if sub.Name == name {
			return sub
		}
	}
	return nil
}
