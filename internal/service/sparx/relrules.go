package sparx

import (
	_ "embed"
	"fmt"
	"strings"
)

// The ArchiMate relationship rules. archimate_relationships.csv holds one row
// per (source, target, relation) — every element type × every element type ×
// every relationship type — with a status of deny / warn / allow. It is
// generated once by tools/reltable -seed and then hand-maintained as Sparx EA
// behaviour is confirmed.
//
//	deny  → ea_create_relationship refuses the relationship
//	warn  → it is created, but the result carries a warning; a model that already
//	        contains such a relationship is flagged when read
//	allow → created silently
//
// A (source, target, relation) triple missing from the file is treated as deny.

//go:embed archimate_relationships.csv
var relationshipRulesCSV string

// Verdict is how the rules classify one (relation, source, target) triple.
type Verdict int

const (
	VerdictDeny Verdict = iota
	VerdictWarn
	VerdictAllow
)

func (v Verdict) String() string {
	switch v {
	case VerdictAllow:
		return "allow"
	case VerdictWarn:
		return "warn"
	default:
		return "deny"
	}
}

var relationshipRules = mustParseRelationshipRules(relationshipRulesCSV)

func relRuleKey(rel, src, tgt string) string { return rel + "\x00" + src + "\x00" + tgt }

func mustParseRelationshipRules(s string) map[string]Verdict {
	out := map[string]Verdict{}
	for i, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "source,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) != 4 {
			panic(fmt.Sprintf("archimate_relationships.csv:%d: want 4 fields, got %d (%q)", i+1, len(f), line))
		}
		src, tgt, rel, status := f[0], f[1], f[2], f[3]
		var v Verdict
		//# ArchiMate relationship rules — status is deny | warn | allow.
		//# Seeded by tools/reltable; edit directly as Sparx EA behaviour is confirmed.
		switch status {
		case "allow":
			// # allow = created silently
			v = VerdictAllow
		case "warn":
			//# warn  = created, but the tool returns a warning; flagged when a model is read
			v = VerdictWarn
		case "deny":
			//# deny  = ea_create_relationship refuses it
			v = VerdictDeny
		default:
			panic(fmt.Sprintf("archimate_relationships.csv:%d: unknown status %q", i+1, status))
		}
		out[relRuleKey(rel, src, tgt)] = v
	}
	return out
}

// RelationshipVerdict classifies an ArchiMate relationship between two element
// types. All names are bare ("Realization", "ApplicationComponent",
// "ApplicationService"). An unlisted triple is VerdictDeny.
func RelationshipVerdict(rel, source, target string) Verdict {
	if v, ok := relationshipRules[relRuleKey(rel, source, target)]; ok {
		return v
	}
	return VerdictDeny
}

// RelationshipVerdictQualified is RelationshipVerdict for qualified names
// ("ArchiMate.Realization"). An unknown name yields VerdictDeny.
func RelationshipVerdictQualified(relType, sourceType, targetType string) Verdict {
	rel, okR := unqualify(relType)
	src, okS := unqualify(sourceType)
	tgt, okT := unqualify(targetType)
	if !okR || !okS || !okT {
		return VerdictDeny
	}
	return RelationshipVerdict(rel, src, tgt)
}

// RelationshipVerdictName is RelationshipVerdictQualified as a string
// ("allow" / "warn" / "deny"). Exported for tooling and tests.
func RelationshipVerdictName(relType, sourceType, targetType string) string {
	return RelationshipVerdictQualified(relType, sourceType, targetType).String()
}
