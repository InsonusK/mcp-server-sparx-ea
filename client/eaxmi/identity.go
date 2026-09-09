package eaxmi

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/beevik/etree"
)

// idTokenRe matches an EA xmi:id / xmi:idref token: EAID_ or EAPK_ followed by a
// GUID with "_" separators, or an association-end token (…_src<6hex>… / …_dst…).
var idTokenRe = regexp.MustCompile(
	`EA(?:ID|PK)_(?:[0-9A-Fa-f]{8}|(?:src|dst)[0-9A-Fa-f]{6})(?:_[0-9A-Fa-f]{4}){3}_[0-9A-Fa-f]{12}`)

// bracedGUIDRe matches "{A-B-C-D-E}" as it appears inside <xrefs value="…">.
var bracedGUIDRe = regexp.MustCompile(
	`\{[0-9A-Fa-f]{8}-(?:[0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}\}`)

// RemapIdentity regenerates the GUID of every package, element, connector,
// diagram and cross-reference in the document, keeping all internal references
// consistent. After it the model is identity-disjoint from the file it was
// loaded from, so both can be imported into the same EA project side by side
// (an XMI import matches by GUID — without this, the second import moves the
// shared content out of the first).
//
// It works on the serialised form (every reference, in any attribute or in the
// packed <xrefs> strings, is a plain substring) and then re-parses.
func (d *Document) RemapIdentity() error {
	var buf bytes.Buffer
	if _, err := d.doc.WriteTo(&buf); err != nil {
		return fmt.Errorf("eaxmi: remap serialise: %w", err)
	}
	s := buf.String()

	// old GUID body ("A_B_C_D_E") -> new body; keyed on the body so EAID_x and
	// EAPK_x for the same package map to the same new GUID.
	newBody := map[string]string{}
	// association-end tokens are remapped whole (their body is not a real GUID).
	endTok := map[string]string{}

	for _, tok := range idTokenRe.FindAllString(s, -1) {
		body := tok[5:] // strip "EAID_" / "EAPK_"
		if strings.HasPrefix(body, "src") || strings.HasPrefix(body, "dst") {
			if _, ok := endTok[tok]; !ok {
				endTok[tok] = xmiIDFromGUID(NewGUID(), tok[:5])
			}
			continue
		}
		if _, ok := newBody[body]; !ok {
			newBody[body] = underscoreBody(NewGUID())
		}
	}
	for _, b := range bracedGUIDRe.FindAllString(s, -1) {
		body := strings.ReplaceAll(strings.Trim(b, "{}"), "-", "_")
		if _, ok := newBody[body]; !ok {
			newBody[body] = underscoreBody(NewGUID())
		}
	}

	type rep struct{ from, to string }
	var reps []rep
	for old, nw := range newBody {
		reps = append(reps,
			rep{"EAID_" + old, "EAID_" + nw},
			rep{"EAPK_" + old, "EAPK_" + nw},
			rep{"{" + dash(old) + "}", "{" + dash(nw) + "}"},
		)
	}
	for old, nw := range endTok {
		reps = append(reps, rep{old, nw})
	}
	sort.Slice(reps, func(i, j int) bool { return len(reps[i].from) > len(reps[j].from) })
	for _, r := range reps {
		s = strings.ReplaceAll(s, r.from, r.to)
	}

	nd := etree.NewDocument()
	nd.ReadSettings = d.doc.ReadSettings
	if err := nd.ReadFromString(s); err != nil {
		return fmt.Errorf("eaxmi: remap re-parse: %w", err)
	}
	np, err := parse(nd)
	if err != nil {
		return err
	}
	path := d.Path
	*d = *np
	d.Path = path
	return nil
}

func underscoreBody(guid string) string {
	return strings.ReplaceAll(strings.Trim(guid, "{}"), "-", "_")
}

func dash(underscoreBody string) string {
	return strings.ReplaceAll(underscoreBody, "_", "-")
}
