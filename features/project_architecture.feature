Feature: Project-wide architectural guarantees
  Cross-cutting concerns from the base-task Definition of Done that are about the
  whole module or the shipped binary, not any single package.

  Scenario: No project package shells out to an external process
    Then no package under "github.com/InsonusK/mcp-server-sparx-ea" imports "os/exec"

  Scenario: The compiled binary links libmdb directly
    Given the server binary is built
    Then the binary is dynamically linked against "libmdb"
