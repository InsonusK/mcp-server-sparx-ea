package eaxmitest

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var (
	idDefRe = regexp.MustCompile(`xmi:id="((?:EAID|EAPK)_[^"]+)"`)
	idRefRe = regexp.MustCompile(`(?:package|package2|base_[A-Za-z]+|source|target|subject|client|supplier|general|start|end)="((?:EAID|EAPK)_[^"]+)"`)
)

// noDanglingRefs reads the last-saved file and checks every id-shaped attribute
// value resolves to a defined xmi:id (matching by GUID body, so EA's
// EAID_/EAPK_ mirror and package2 convention are not flagged).
func noDanglingRefs(_ context.Context, w *world) error {
	if w.savedPath == "" {
		return fmt.Errorf("nothing has been written yet")
	}
	b, err := os.ReadFile(w.savedPath)
	if err != nil {
		return err
	}
	s := string(b)

	defined := map[string]bool{}
	for _, m := range idDefRe.FindAllStringSubmatch(s, -1) {
		defined[body(m[1])] = true
	}
	missing := map[string]bool{}
	for _, m := range idRefRe.FindAllStringSubmatch(s, -1) {
		if !defined[body(m[1])] {
			missing[m[1]] = true
		}
	}
	if len(missing) == 0 {
		return nil
	}
	var list []string
	for k := range missing {
		list = append(list, k)
	}
	sort.Strings(list)
	return fmt.Errorf("%d dangling id reference(s): %s", len(list), strings.Join(list, ", "))
}

func body(id string) string {
	id = strings.TrimPrefix(strings.TrimPrefix(id, "EAID_"), "EAPK_")
	return strings.TrimPrefix(strings.TrimPrefix(id, "src"), "dst")
}
