package features

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/bddsupport"
)

// noProjectPackageImportsOsExec proves that no first-party package shells out to
// a subprocess.
func noProjectPackageImportsOsExec(module string) error {
	cmd := exec.Command("go", "list", "-deps",
		"-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./...")
	cmd.Dir = bddsupport.RepoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go list failed: %v\n%s", err, out)
	}
	var offenders []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], module) {
			continue
		}
		for _, imp := range fields[1:] {
			if imp == "os/exec" {
				offenders = append(offenders, fields[0])
			}
		}
	}
	if len(offenders) != 0 {
		return fmt.Errorf("packages import os/exec: %v", offenders)
	}
	return nil
}

func buildWith(env ...string) error {
	cmd := exec.Command("go", "build", "-o", os.DevNull, ".")
	cmd.Dir = bddsupport.RepoRoot()
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build (%v) failed: %v\n%s", env, err, out)
	}
	return nil
}

func binaryBuildsWithCGODisabled() error {
	return buildWith("CGO_ENABLED=0")
}

func binaryBuildsFor(target string) error {
	goos, goarch, ok := strings.Cut(target, "/")
	if !ok {
		return fmt.Errorf("target %q is not GOOS/GOARCH", target)
	}
	return buildWith("CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch)
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^no package under "([^"]*)" imports "os/exec"$`, noProjectPackageImportsOsExec)
	sc.Step(`^the binary builds with CGO disabled$`, binaryBuildsWithCGODisabled)
	sc.Step(`^the binary builds for "([^"]*)"$`, binaryBuildsFor)
}
