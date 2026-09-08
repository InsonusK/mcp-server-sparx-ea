package eapx

import "testing"

// firstWord is pure string logic on the hot path of every Query call; these
// cases pin its contract directly (the Cucumber suite exercises it only
// indirectly through write-statement rejection).
func TestFirstWord(t *testing.T) {
	cases := map[string]string{
		"select 1":      "select",
		"  SELECT  x":   "select",
		"\n\tSelect\n":  "select",
		"(select 1)":    "select",
		"DELETE from t": "delete",
		"Insert into t": "insert",
		"":              "",
		"UPDATE":        "update",
		"drop;":         "drop",
		";select 1":     "",
	}
	for in, want := range cases {
		if got := firstWord(in); got != want {
			t.Errorf("firstWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWriteVerbsRejectedBeforeEngine(t *testing.T) {
	c := &Connector{path: "unused"} // engine is nil: a write verb must be rejected before it is touched
	for _, sql := range []string{"insert into t values (1)", "UPDATE t SET a=1", "delete from t", "drop table t"} {
		rs, err := c.Query(sql)
		if rs != nil || err == nil {
			t.Fatalf("Query(%q): expected rejection, got rs=%v err=%v", sql, rs, err)
		}
		var e *Error
		if !asError(err, &e) || e.msg != "only read-only SELECT statements are supported" {
			t.Fatalf("Query(%q): unexpected error %v", sql, err)
		}
	}
}

// The low-level engine wrapper must tolerate being used with no handle: query
// reports "closed" and close is a safe no-op (also covering a double Close).
func TestMdbSQLZeroValueIsSafe(t *testing.T) {
	m := &mdbSQL{}
	if _, err := m.query("select 1"); err == nil {
		t.Fatal("query on a nil handle must error")
	}
	m.close()
	m.close() // must not panic
}

func asError(err error, target **Error) bool {
	if e, ok := err.(*Error); ok {
		*target = e
		return true
	}
	return false
}
