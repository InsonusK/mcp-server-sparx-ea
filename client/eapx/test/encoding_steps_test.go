package eapxtest

// Steps for features/text_encoding.feature.

import (
	"context"
	"encoding/hex"
	"fmt"
	"unicode/utf8"

	"github.com/cucumber/godog"
)

func (w *world) everyResultValueIsValidUTF8(ctx context.Context) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	checked := 0
	for r, row := range w.rs.Rows {
		for c, cell := range row {
			if !utf8.ValidString(cell) {
				return fmt.Errorf("row %d column %q is not valid UTF-8: %x", r, w.rs.Columns[c], cell)
			}
			checked++
		}
	}
	logf(ctx, "all %d text value(s) are valid UTF-8", checked)
	return nil
}

func (w *world) theSingleCellIsUTF8Bytes(ctx context.Context, wantHex string) error {
	if err := w.ensureQuerySucceeded(); err != nil {
		return err
	}
	if len(w.rs.Rows) != 1 || len(w.rs.Rows[0]) != 1 {
		return fmt.Errorf("expected exactly one row with one column, got %d row(s) / columns %v",
			len(w.rs.Rows), w.rs.Columns)
	}
	got := hex.EncodeToString([]byte(w.rs.Rows[0][0]))
	if got != wantHex {
		return fmt.Errorf("cell %q: bytes %s, want %s", w.rs.Rows[0][0], got, wantHex)
	}
	logf(ctx, "cell %q decoded to the exact bytes %s", w.rs.Rows[0][0], got)
	return nil
}

func registerEncodingSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^every result value is valid UTF-8$`, w.everyResultValueIsValidUTF8)
	sc.Step(`^the single result cell is the UTF-8 bytes "([0-9a-fA-F]+)"$`, w.theSingleCellIsUTF8Bytes)
}
