package eapxtest

// Steps for features/in_process_execution.feature.

import (
	"context"
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

func (w *world) recordOnDiskState(ctx context.Context, name string) error {
	path := fixturePath(name)
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	w.recordedPath = path
	w.recordedStat = st
	logf(ctx, "recorded %s: size=%d mtime=%s", path, st.Size(), st.ModTime().Format("15:04:05.000000000"))
	return nil
}

func (w *world) emptyDirAsTMPDIR(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "eapx-tmpdir-*")
	if err != nil {
		return err
	}
	w.prevTMPDIR, w.hadTMPDIR = os.LookupEnv("TMPDIR")
	if err := os.Setenv("TMPDIR", dir); err != nil {
		return err
	}
	w.tmpDir = dir
	w.tmpDirSet = true
	logf(ctx, "TMPDIR set to fresh empty dir %s", dir)
	return nil
}

func (w *world) fileOnDiskUnchanged(ctx context.Context, name string) error {
	if w.recordedStat == nil {
		return fmt.Errorf("on-disk state was not recorded first")
	}
	path := fixturePath(name)
	if path != w.recordedPath {
		return fmt.Errorf("recorded %q but asked about %q", w.recordedPath, path)
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.Size() != w.recordedStat.Size() {
		return fmt.Errorf("%s size changed: %d → %d", path, w.recordedStat.Size(), st.Size())
	}
	if !st.ModTime().Equal(w.recordedStat.ModTime()) {
		return fmt.Errorf("%s mtime changed: %s → %s", path, w.recordedStat.ModTime(), st.ModTime())
	}
	logf(ctx, "%s unchanged on disk (size=%d, mtime intact)", path, st.Size())
	return nil
}

func (w *world) tmpdirStillEmpty(ctx context.Context) error {
	if !w.tmpDirSet {
		return fmt.Errorf("TMPDIR was not set")
	}
	entries, err := os.ReadDir(w.tmpDir)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		return fmt.Errorf("TMPDIR is not empty: %v", names)
	}
	logf(ctx, "TMPDIR %s is still empty", w.tmpDir)
	return nil
}

func registerArchitectureSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the on-disk state of "([^"]*)" is recorded$`, w.recordOnDiskState)
	sc.Step(`^an empty directory is set as TMPDIR$`, w.emptyDirAsTMPDIR)
	sc.Step(`^the "([^"]*)" file on disk is unchanged$`, w.fileOnDiskUnchanged)
	sc.Step(`^TMPDIR is still empty$`, w.tmpdirStillEmpty)
}
