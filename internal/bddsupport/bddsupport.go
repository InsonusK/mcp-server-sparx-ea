// Package bddsupport holds the few helpers shared by the godog step
// definitions that live next to each package under test.
package bddsupport

import (
	"os"
	"path/filepath"
)

// RepoRoot walks up from the current working directory to the module root
// (the directory containing go.mod). godog runs each package's steps with that
// package's directory as the working directory, so callers need this to reach
// shared fixtures under example/ and to run project-wide commands.
func RepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

// ResolvePath maps a path as written in a .feature file to a real path: an
// absolute path and the empty string are returned unchanged, a relative path
// is taken relative to the repository root.
func ResolvePath(p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(RepoRoot(), p)
}
