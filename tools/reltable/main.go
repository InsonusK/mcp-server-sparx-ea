// Command reltable maintains the ArchiMate relationship rules:
//
//	go run ./tools/reltable -seed [-force]
//	    Seed internal/service/sparx/archimate_relationships.csv — one row per
//	    (source, target, relation): every element type × every element type ×
//	    every relationship type. Status is "deny" where the ArchiMate 3.2
//	    notation forbids the relationship (forbidden() below), "allow" where the
//	    classic aspect rules clearly permit it, and "warn" for everything in
//	    between — permitted but to be verified against Sparx EA. Refuses to
//	    overwrite an existing file unless -force.
//
//	go run ./tools/reltable
//	    Render docs/archimate-relationship-matrix.md from that CSV (a read-only
//	    view: rows = target element, columns = relationship type, "+" allow,
//	    "!" warn, "-" deny).
//
// The CSV is the source of truth the service embeds; edit it directly as Sparx
// EA behaviour is confirmed. Re-seeding discards manual edits.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/service/sparx"
)

const (
	csvPath = "internal/service/sparx/archimate_relationships.csv"
	mdPath  = "docs/archimate-relationship-matrix.md"
)

// layers is the ArchiMate layer grouping (presentation + iteration order). Every
// element sparx.ElementTypes() returns must appear in exactly one layer.
var layers = []struct {
	name  string
	types []string
}{
	{"Motivation", []string{
		"Stakeholder", "Driver", "Assessment", "Goal", "Outcome",
		"Principle", "Requirement", "Constraint", "Meaning", "Value"}},
	{"Strategy", []string{
		"Resource", "Capability", "CourseOfAction", "ValueStream"}},
	{"Business", []string{
		"BusinessActor", "BusinessRole", "BusinessCollaboration", "BusinessInterface",
		"BusinessProcess", "BusinessFunction", "BusinessInteraction", "BusinessEvent",
		"BusinessService", "BusinessObject", "Contract", "Representation", "Product"}},
	{"Application", []string{
		"ApplicationComponent", "ApplicationCollaboration", "ApplicationInterface",
		"ApplicationFunction", "ApplicationInteraction", "ApplicationProcess",
		"ApplicationEvent", "ApplicationService", "DataObject"}},
	{"Technology", []string{
		"Node", "Device", "SystemSoftware", "TechnologyCollaboration", "TechnologyInterface",
		"Path", "CommunicationNetwork", "TechnologyFunction", "TechnologyProcess",
		"TechnologyInteraction", "TechnologyEvent", "TechnologyService", "Artifact"}},
	{"Physical", []string{
		"Equipment", "Facility", "DistributionNetwork", "Material"}},
	{"Implementation & Migration", []string{
		"WorkPackage", "Deliverable", "ImplementationEvent", "Plateau", "Gap"}},
	{"Other", []string{"Location", "Grouping"}},
}

// rels is the relationship column order.
var rels = []struct{ label, name string }{
	{"Comp", "Composition"},
	{"Aggr", "Aggregation"},
	{"Assign", "Assignment"},
	{"Real", "Realization"},
	{"Serv", "Serving"},
	{"Access", "Access"},
	{"Infl", "Influence"},
	{"Trig", "Triggering"},
	{"Flow", "Flow"},
	{"Spec", "Specialization"},
	{"Assoc", "Association"},
}

// aspectOf classifies an element for the seed rules. Finer than the service's
// own aspects: events, location and grouping are split out.
var aspectOf = map[string]string{}

func init() {
	set := func(a string, ts ...string) {
		for _, t := range ts {
			aspectOf[t] = a
		}
	}
	set("motivation", "Stakeholder", "Driver", "Assessment", "Goal", "Outcome",
		"Principle", "Requirement", "Constraint", "Meaning", "Value")
	set("strategy", "Resource", "Capability", "CourseOfAction", "ValueStream")
	set("active",
		"BusinessActor", "BusinessRole", "BusinessCollaboration", "BusinessInterface",
		"ApplicationComponent", "ApplicationCollaboration", "ApplicationInterface",
		"Node", "Device", "SystemSoftware", "TechnologyCollaboration", "TechnologyInterface",
		"Path", "CommunicationNetwork", "Equipment", "Facility", "DistributionNetwork")
	set("behavior",
		"BusinessProcess", "BusinessFunction", "BusinessInteraction", "BusinessService",
		"ApplicationFunction", "ApplicationInteraction", "ApplicationProcess", "ApplicationService",
		"TechnologyFunction", "TechnologyProcess", "TechnologyInteraction", "TechnologyService")
	set("event", "BusinessEvent", "ApplicationEvent", "TechnologyEvent")
	set("passive", "BusinessObject", "Contract", "Representation", "Product",
		"DataObject", "Artifact", "Material")
	set("impl", "WorkPackage", "Deliverable", "ImplementationEvent", "Plateau", "Gap")
	set("location", "Location")
	set("grouping", "Grouping")
}

// coarse folds the seed aspects back to the ones the classic aspect rules use.
func coarse(a string) string {
	switch a {
	case "event":
		return "behavior"
	case "location", "grouping":
		return "composite"
	}
	return a
}

// forbidden: the relationship is *definitely* not permitted by ArchiMate 3.2.
// Conservative — only high-confidence rules.
func forbidden(rel, src, tgt string) bool {
	as, at := aspectOf[src], aspectOf[tgt]
	mot := func(a string) bool { return a == "motivation" }

	switch rel {
	case "Association":
		return false
	case "Specialization":
		return src != tgt
	}

	// Location and Grouping take part only in Association, Composition,
	// Aggregation and Specialization.
	if as == "location" || as == "grouping" || at == "location" || at == "grouping" {
		return rel != "Composition" && rel != "Aggregation"
	}

	// Motivation elements never take part in the dynamic / allocation relations.
	if mot(as) || mot(at) {
		switch rel {
		case "Assignment", "Serving", "Access", "Triggering", "Flow":
			return true
		}
	}

	switch rel {
	case "Composition", "Aggregation":
		return mot(as) != mot(at) // motivation only nests within motivation

	case "Influence":
		return !mot(at) // influence always targets a motivation element

	case "Realization":
		if as == "event" || at == "event" {
			return true
		}
		if tgt == "Driver" || tgt == "Assessment" || tgt == "Stakeholder" {
			return true
		}
		if mot(as) {
			switch src {
			case "Outcome", "Requirement", "Principle", "Constraint":
				return !mot(at)
			default:
				return true
			}
		}
		if at == "passive" && as != "passive" {
			return true
		}
		if as == "behavior" && at == "active" {
			return true
		}
		switch tgt {
		case "Goal":
			return src != "Requirement" && src != "Constraint" && src != "Principle" &&
				src != "Outcome" && src != "CourseOfAction"
		case "Meaning":
			return src != "Representation"
		}
		return false

	case "Access":
		return (as != "behavior" && as != "event") || at != "passive"

	case "Triggering", "Flow":
		return as == "passive" || at == "passive"

	case "Assignment":
		if as == "behavior" || as == "event" || as == "passive" || as == "impl" {
			return true
		}
		return as == "strategy" && src != "Resource"

	case "Serving":
		if as == "passive" || as == "impl" {
			return true
		}
		return at == "passive"
	}
	return false
}

// seedAllows: the classic aspect rules clearly permit the relationship.
func seedAllows(rel, src, tgt string) bool {
	sa, ta := coarse(aspectOf[src]), coarse(aspectOf[tgt])
	switch rel {
	case "Association":
		return true
	case "Specialization", "Composition", "Aggregation":
		return src == tgt
	case "Influence":
		return ta == "motivation"
	case "Realization":
		if ta == "motivation" {
			return true
		}
		if sa == "motivation" {
			return false
		}
		return true
	case "Assignment":
		return sa == "active" && (ta == "behavior" || ta == "active" || ta == "passive")
	case "Serving":
		return (sa == "active" || sa == "behavior") && ta != "passive"
	case "Access":
		return sa == "behavior" && ta == "passive"
	case "Triggering", "Flow":
		return sa == "behavior" && ta == "behavior"
	}
	return false
}

func verdict(rel, src, tgt string) string {
	switch {
	case forbidden(rel, src, tgt):
		return "deny"
	case seedAllows(rel, src, tgt):
		return "allow"
	default:
		return "warn"
	}
}

func main() {
	seed := flag.Bool("seed", false, "seed "+csvPath+" from the rules (instead of rendering the view)")
	force := flag.Bool("force", false, "with -seed: overwrite an existing CSV")
	flag.Parse()

	allTypes := checkLayers()

	if *seed {
		seedCSV(*force)
		return
	}
	renderMarkdown(allTypes)
}

func checkLayers() []string {
	placed := map[string]int{}
	for _, l := range layers {
		for _, t := range l.types {
			placed[t]++
		}
	}
	var all []string
	for _, q := range sparx.ElementTypes() {
		name := strings.TrimPrefix(q, "ArchiMate.")
		if placed[name] != 1 {
			log.Fatalf("element %q placed in %d layers (want 1)", name, placed[name])
		}
		if aspectOf[name] == "" {
			log.Fatalf("element %q has no aspect", name)
		}
		all = append(all, name)
	}
	return all
}

func seedCSV(force bool) {
	if _, err := os.Stat(csvPath); err == nil && !force {
		log.Fatalf("%s already exists — edit it directly, or pass -force to reseed", csvPath)
	}

	f, err := os.Create(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "# ArchiMate relationship rules — status is deny | warn | allow.")
	fmt.Fprintln(w, "# deny  = ea_create_relationship refuses it")
	fmt.Fprintln(w, "# warn  = created, but the tool returns a warning; flagged when a model is read")
	fmt.Fprintln(w, "# allow = created silently")
	fmt.Fprintln(w, "# Seeded by tools/reltable; edit directly as Sparx EA behaviour is confirmed.")
	fmt.Fprintln(w, "source,target,relation,status")

	var counts = map[string]int{}
	for _, sl := range layers {
		for _, src := range sl.types {
			for _, tl := range layers {
				for _, tgt := range tl.types {
					for _, r := range rels {
						v := verdict(r.name, src, tgt)
						counts[v]++
						fmt.Fprintf(w, "%s,%s,%s,%s\n", src, tgt, r.name, v)
					}
				}
			}
		}
	}
	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("seeded %s (%d allow, %d warn, %d deny)\n",
		csvPath, counts["allow"], counts["warn"], counts["deny"])
}

func renderMarkdown(all []string) {
	rules, err := readCSV(csvPath)
	if err != nil {
		log.Fatal(err)
	}

	sym := map[string]string{"allow": "+", "warn": "!", "deny": "-"}

	var b strings.Builder
	b.WriteString(header)

	colHead := "| Цель \\ Связь |"
	colSep := "| --- |"
	for _, r := range rels {
		colHead += " " + r.label + " |"
		colSep += " --- |"
	}

	for _, l := range layers {
		fmt.Fprintf(&b, "# %s\n\n", l.name)
		for _, src := range l.types {
			fmt.Fprintf(&b, "## %s → …\n\n%s\n%s\n", src, colHead, colSep)
			for _, ll := range layers {
				for _, tgt := range ll.types {
					row := "| " + tgt + " |"
					for _, r := range rels {
						row += " " + sym[rules[key(src, tgt, r.name)]] + " |"
					}
					b.WriteString(row + "\n")
				}
			}
			b.WriteString("\n")
		}
	}
	if err := os.WriteFile(mdPath, []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
	n := len(all)
	fmt.Printf("wrote %s (%d×%d×%d from %s)\n", mdPath, n, n, len(rels), csvPath)
}

func key(src, tgt, rel string) string { return src + "|" + tgt + "|" + rel }

func readCSV(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "source,") {
			continue
		}
		p := strings.Split(line, ",")
		if len(p) != 4 {
			return nil, fmt.Errorf("bad row %q", line)
		}
		out[key(p[0], p[1], p[2])] = p[3]
	}
	return out, sc.Err()
}

var _ = sort.Strings

const header = `# Матрица связей ArchiMate — вьюха

Генерируется из [internal/service/sparx/archimate_relationships.csv](../internal/service/sparx/archimate_relationships.csv)
(` + "`go run ./tools/reltable`" + `) — **правь CSV, не этот файл**.

Для каждого элемента-источника: строки — цели, колонки — типы связей.

| символ | статус | поведение сервиса |
| --- | --- | --- |
| ` + "`+`" + ` | allow | ` + "`ea_create_relationship`" + ` создаёт молча |
| ` + "`!`" + ` | warn | создаёт, но возвращает предупреждение; помечается при чтении модели |
| ` + "`-`" + ` | deny | ` + "`ea_create_relationship`" + ` отказывает |

Seed: ` + "`-`" + ` — где нотация ArchiMate 3.2 точно запрещает; ` + "`+`" + ` — где
классические аспектные правила точно разрешают; ` + "`!`" + ` — всё между (создаётся,
но требует сверки с Sparx EA).

Столбцы: **Comp**osition · **Aggr**egation · **Assign**ment · **Real**ization ·
**Serv**ing · **Access** · **Infl**uence · **Trig**gering · **Flow** ·
**Spec**ialization · **Assoc**iation.

`
