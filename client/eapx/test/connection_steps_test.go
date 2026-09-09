package eapxtest

// Steps for features/open_file.feature (opening / closing a connection).
// Shared by the Background of the other feature files.

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"github.com/InsonusK/mcp-server-sparx-ea/client/eapx"
)

func (w *world) iOpenTheFile(ctx context.Context, name string) error {
	path := fixturePath(name)
	w.conn, w.openErr = eapx.Open(path)
	if w.openErr != nil {
		logf(ctx, "open %q → error: %v", path, w.openErr)
	} else {
		logf(ctx, "open %q → connection established", path)
	}
	return nil
}

func (w *world) connectionOpensSuccessfully(ctx context.Context) error {
	if w.openErr != nil {
		return fmt.Errorf("expected open to succeed, got: %v", w.openErr)
	}
	if w.conn == nil {
		return fmt.Errorf("expected a connection, got nil")
	}
	logf(ctx, "connection is usable")
	return nil
}

func (w *world) openingFailsWith(ctx context.Context, msg string) error {
	if w.openErr == nil {
		return fmt.Errorf("expected open to fail with %q, but it succeeded", msg)
	}
	if !strings.Contains(w.openErr.Error(), msg) {
		return fmt.Errorf("open error %q does not contain %q", w.openErr.Error(), msg)
	}
	logf(ctx, "open failed as expected: %v", w.openErr)
	return nil
}

// givenAConnection is the "Given a connection to ..." setup step: it fails the
// scenario immediately if the fixture cannot be opened.
func (w *world) givenAConnection(ctx context.Context, name string) error {
	path := fixturePath(name)
	conn, err := eapx.Open(path)
	if err != nil {
		return fmt.Errorf("could not open fixture %q: %w", path, err)
	}
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.conn = conn
	logf(ctx, "connected to %q", path)
	return nil
}

func (w *world) iCloseTheConnection(ctx context.Context) error {
	if w.conn == nil {
		return fmt.Errorf("no open connection to close")
	}
	err := w.conn.Close()
	logf(ctx, "closed the connection (err=%v)", err)
	return err
}

func registerConnectionSteps(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^I open the Sparx EA file "([^"]*)"$`, w.iOpenTheFile)
	sc.Step(`^the connection opens successfully$`, w.connectionOpensSuccessfully)
	sc.Step(`^opening fails with "([^"]*)"$`, w.openingFailsWith)
	sc.Step(`^a connection to the Sparx EA file "([^"]*)"$`, w.givenAConnection)
	sc.Step(`^I close the connection$`, w.iCloseTheConnection)
}
