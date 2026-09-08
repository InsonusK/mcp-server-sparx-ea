package eapx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture is the populated sample project shipped in example/.
func fixture(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("..", "example", "TestProject.eapx"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	return p
}

// TestConnectorAgainstFixture is an in-package integration test that drives the
// cgo binding (cgo_mdb.go) end to end. The Cucumber suite covers the same
// behaviour from the outside; this one exists so mutation testing on the eapx
// package has runnable mutants on the binding layer.
func TestConnectorAgainstFixture(t *testing.T) {
	c, err := Open(fixture(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer c.Close()

	t.Run("select returns columns, rows and count", func(t *testing.T) {
		rs, err := c.Query("select Object_ID, Name from t_object where Object_ID = 2")
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if want := []string{"Object_ID", "Name"}; !equalStr(rs.Columns, want) {
			t.Fatalf("columns = %v, want %v", rs.Columns, want)
		}
		if rs.RowCount != 1 || len(rs.Rows) != 1 {
			t.Fatalf("RowCount=%d rows=%d, want 1/1", rs.RowCount, len(rs.Rows))
		}
		if got := rs.Rows[0]; !equalStr(got, []string{"2", "Stakeholder1"}) {
			t.Fatalf("row = %v, want [2 Stakeholder1]", got)
		}
	})

	t.Run("empty result set has zero rows", func(t *testing.T) {
		rs, err := c.Query("select Object_ID from t_object where Object_ID = 999999")
		if err != nil {
			t.Fatalf("Query: %v", err)
		}
		if rs.RowCount != 0 || len(rs.Rows) != 0 {
			t.Fatalf("RowCount=%d rows=%d, want 0/0", rs.RowCount, len(rs.Rows))
		}
	})

	t.Run("engine error is surfaced", func(t *testing.T) {
		_, err := c.Query("select * from t_not_a_table")
		var e *Error
		if !errors.As(err, &e) || e.Op != "query" {
			t.Fatalf("expected *Error with Op=query, got %v", err)
		}
		if !strings.Contains(e.Error(), "(sql: select * from t_not_a_table)") {
			t.Fatalf("error string should echo the failing SQL, got %q", e.Error())
		}
	})

	t.Run("query after close fails", func(t *testing.T) {
		c2, err := Open(fixture(t))
		if err != nil {
			t.Fatal(err)
		}
		_ = c2.Close()
		if _, err := c2.Query("select Object_ID from t_object"); err == nil {
			t.Fatal("expected error querying a closed connector")
		}
	})
}

func TestOpenRejectsMissingAndEmpty(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("empty path must fail")
	}
	_, err := Open(filepath.Join(t.TempDir(), "nope.eapx"))
	if err == nil {
		t.Fatal("missing file must fail")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("want the error to unwrap to os.ErrNotExist, got %v", err)
	}
}

func equalStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
