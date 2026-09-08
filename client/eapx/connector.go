// Package eapx is a connector for Sparx Enterprise Architect project files
// (.eapx / .eap), which are Microsoft Access JET databases. It binds directly
// to the mdbtools C library (libmdb / libmdbsql) via cgo and runs SQL against
// the file in-process — no subprocess, no temporary files.
package eapx

import (
	"os"
	"strings"
	"sync"
)

func init() {
	// mdbtools decodes JET4 (UCS-2LE) text through iconv; the target charset
	// comes from $MDBICONV and defaults to UTF-8. Pin it so the connector's
	// output does not depend on the caller's environment.
	if _, ok := os.LookupEnv("MDBICONV"); !ok {
		_ = os.Setenv("MDBICONV", "UTF-8")
	}
}

// ResultSet is the outcome of a successful SELECT. Every value is returned as a
// string, exactly as mdbtools renders it; an SQL NULL is the empty string.
type ResultSet struct {
	Columns  []string
	Rows     [][]string
	RowCount int
}

// Connector holds an open engine bound to a single .eapx file. All methods are
// safe for concurrent use — the underlying mdbtools engine is single-threaded,
// so calls are serialized.
type Connector struct {
	path string

	mu     sync.Mutex
	engine *mdbSQL
}

// Open connects to the Sparx EA file at path. It returns an error if the file
// does not exist or is not a readable JET database.
func Open(path string) (*Connector, error) {
	if strings.TrimSpace(path) == "" {
		return nil, &Error{Op: "open", Path: path, msg: "empty file path"}
	}
	if _, err := os.Stat(path); err != nil {
		return nil, &Error{Op: "open", Path: path, msg: "file not found"}
	}

	engine, err := mdbOpen(path)
	if err != nil {
		return nil, &Error{Op: "open", Path: path, msg: err.Error()}
	}
	return &Connector{path: path, engine: engine}, nil
}

// writeVerbs are the leading keywords of statements that would mutate the file.
// The connector is read-only, so these are rejected before reaching mdbtools
// (which would otherwise report an opaque "syntax error").
var writeVerbs = map[string]bool{
	"insert": true, "update": true, "delete": true, "replace": true,
	"create": true, "drop": true, "alter": true, "truncate": true,
	"grant": true, "revoke": true, "merge": true,
}

// Query runs one SQL statement against the file and returns its result set.
// Only read-only SELECT statements are supported.
func (c *Connector) Query(sql string) (*ResultSet, error) {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil, &Error{Op: "query", Path: c.path, msg: "empty SQL statement"}
	}
	// trimmed is non-empty and starts with a non-space rune, so Fields yields
	// at least one element.
	if verb := strings.ToLower(strings.Fields(trimmed)[0]); writeVerbs[verb] {
		return nil, &Error{
			Op:   "query",
			Path: c.path,
			SQL:  sql,
			msg:  "only read-only SELECT statements are supported",
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.engine == nil {
		return nil, &Error{Op: "query", Path: c.path, msg: "connector is closed"}
	}

	rs, err := c.engine.query(sql)
	if err != nil {
		return nil, &Error{Op: "query", Path: c.path, msg: err.Error(), SQL: sql}
	}
	return rs, nil
}

// Close releases the engine and the file handle. Further calls to Query fail.
// Calling Close more than once is a no-op.
func (c *Connector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.engine != nil {
		c.engine.close()
		c.engine = nil
	}
	return nil
}
