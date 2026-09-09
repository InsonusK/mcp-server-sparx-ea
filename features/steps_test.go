package features

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/internal/bddsupport"
)

type world struct {
	binPath string
}

func (w *world) reset() { *w = world{} }

// noProjectPackageImportsOsExec proves the DoD guarantee that no first-party
// package shells out to a subprocess.
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

func (w *world) theServerBinaryIsBuilt() error {
	root := bddsupport.RepoRoot()
	outPath := filepath.Join(root, "tmp", "bin", "mcp-server-sparx-ea")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %v\n%s", err, out)
	}
	w.binPath = outPath
	return nil
}

func (w *world) binaryLinkedAgainst(lib string) error {
	if w.binPath == "" {
		return fmt.Errorf("the binary has not been built yet")
	}
	out, err := exec.Command("ldd", w.binPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ldd failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), lib) {
		return fmt.Errorf("binary is not linked against %q:\n%s", lib, out)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	w := &world{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.reset()
		return ctx, nil
	})

	sc.Step(`^no package under "([^"]*)" imports "os/exec"$`, noProjectPackageImportsOsExec)
	sc.Step(`^the server binary is built$`, w.theServerBinaryIsBuilt)
	sc.Step(`^the binary is dynamically linked against "([^"]*)"$`, w.binaryLinkedAgainst)
}
