Feature: Project-wide architectural guarantees
  Cross-cutting concerns that are about the whole module or the shipped binary,
  not any single package.

  Scenario: No project package shells out to an external process
    Then no package under "github.com/InsonusK/mcp-server-sparx-ea" imports "os/exec"

  Scenario: The binary is pure Go and cross-compiles
    Then the binary builds with CGO disabled
    And the binary builds for "windows/amd64"
    And the binary builds for "darwin/arm64"
