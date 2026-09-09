package eaxmi

import "strings"

// Navigation over the parsed model. A ref is one of:
//   - an EA xmi:id      "EAID_F3C2209E_B146_4328_863E_622BABD2B43E" / "EAPK_…"
//   - a braced GUID     "{F3C2209E-B146-4328-863E-622BABD2B43E}"
//   - a slash path      "Model/Motivation_Package/Goal1" (from an EA root package)

// IsRefID reports whether ref is an id/GUID rather than a path.
func IsRefID(ref string) bool {
	return strings.HasPrefix(ref, "EAID_") || strings.HasPrefix(ref, "EAPK_") ||
		(strings.HasPrefix(ref, "{") && strings.HasSuffix(ref, "}"))
}

// SplitPath splits "A/B/C" into non-empty, trimmed segments.
func SplitPath(ref string) []string {
	var out []string
	for _, seg := range strings.Split(ref, "/") {
		if seg = strings.TrimSpace(seg); seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

// ResolveElement returns the element identified by ref, or nil.
func (d *Document) ResolveElement(ref string) *Element {
	if IsRefID(ref) {
		for _, e := range d.AllElements() {
			if e.XMIID == ref || strings.EqualFold(e.GUID, ref) {
				return e
			}
		}
		return nil
	}
	segs := SplitPath(ref)
	if len(segs) == 0 {
		return nil
	}
	pkg := d.walkPackages(segs[:len(segs)-1])
	if pkg == nil {
		return nil
	}
	for _, e := range pkg.Elements {
		if e.Name == segs[len(segs)-1] {
			return e
		}
	}
	return nil
}

// ResolveDiagram returns the diagram identified by ref, or nil.
func (d *Document) ResolveDiagram(ref string) *Diagram {
	if IsRefID(ref) {
		for _, p := range d.AllPackages() {
			for _, g := range p.Diagrams {
				if g.XMIID == ref || strings.EqualFold(g.GUID, ref) {
					return g
				}
			}
		}
		return nil
	}
	segs := SplitPath(ref)
	if len(segs) == 0 {
		return nil
	}
	pkg := d.walkPackages(segs[:len(segs)-1])
	if pkg == nil {
		return nil
	}
	for _, g := range pkg.Diagrams {
		if g.Name == segs[len(segs)-1] {
			return g
		}
	}
	return nil
}

// ResolvePackage returns the package identified by ref, or nil.
func (d *Document) ResolvePackage(ref string) *Package {
	if IsRefID(ref) {
		for _, p := range d.AllPackages() {
			if p.XMIID == ref || strings.EqualFold(p.GUID, ref) {
				return p
			}
		}
		return nil
	}
	return d.walkPackages(SplitPath(ref))
}

// walkPackages follows package-name segments down from the synthetic model root.
func (d *Document) walkPackages(segs []string) *Package {
	cur := d.Root
	for _, seg := range segs {
		var next *Package
		for _, p := range cur.Packages {
			if p.Name == seg {
				next = p
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// AllPackages returns every package below the model root (depth-first).
func (d *Document) AllPackages() []*Package {
	var out []*Package
	var rec func(p *Package)
	rec = func(p *Package) {
		for _, sub := range p.Packages {
			out = append(out, sub)
			rec(sub)
		}
	}
	rec(d.Root)
	return out
}

// AllElements returns every element in the model.
func (d *Document) AllElements() []*Element {
	var out []*Element
	for _, p := range d.AllPackages() {
		out = append(out, p.Elements...)
	}
	return out
}

// PathOfPackage returns "A/B/C" from an EA root package down to p.
func (d *Document) PathOfPackage(p *Package) string {
	var segs []string
	for cur := p; cur != nil && cur != d.Root; cur = cur.Parent {
		segs = append([]string{cur.Name}, segs...)
	}
	return strings.Join(segs, "/")
}

// PathOfElement returns the slash path to e.
func (d *Document) PathOfElement(e *Element) string {
	if e == nil {
		return ""
	}
	return joinSeg(d.PathOfPackage(e.Package), e.Name)
}

// PathOfDiagram returns the slash path to g.
func (d *Document) PathOfDiagram(g *Diagram) string {
	if g == nil {
		return ""
	}
	return joinSeg(d.PathOfPackage(g.Package), g.Name)
}

func joinSeg(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}
