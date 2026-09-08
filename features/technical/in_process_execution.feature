Feature: Queries execute in-process
  Architectural concern (base-task DoD): the server binds libmdbtools directly
  through cgo. It must not spawn child processes or write temporary runtime
  files while answering a query.

  Scenario: No project package shells out to an external process
    Then no package under "github.com/InsonusK/mcp-server-sparx-ea" imports "os/exec"

  Scenario: Answering queries writes no temporary files
    Given an empty directory registered as the process temp dir
    And a connection to the Sparx EA file "example/TestProject.eapx"
    When I run the query "select Object_ID, Name from t_object" 25 times
    Then the temp dir is still empty

  Scenario: The compiled binary links libmdb directly
    Given the server binary is built
    Then the binary is dynamically linked against "libmdb"
