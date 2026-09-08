package eapx

import "fmt"

// Error is returned by every Connector operation that fails. Op is the logical
// operation ("open" or "query"), Path is the file involved, and SQL is the
// statement text when the failure happened during Query.
type Error struct {
	Op   string
	Path string
	SQL  string
	msg  string
}

func (e *Error) Error() string {
	if e.SQL != "" {
		return fmt.Sprintf("eapx %s %q: %s (sql: %s)", e.Op, e.Path, e.msg, e.SQL)
	}
	return fmt.Sprintf("eapx %s %q: %s", e.Op, e.Path, e.msg)
}
