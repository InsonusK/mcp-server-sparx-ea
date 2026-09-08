Feature: Open a connection to a Sparx EA file
  As an MCP server component
  I want to open a connection to a .eapx file by path
  So that I can run SQL against it or fail clearly when the file is unusable

  Scenario Outline: Opening a connection by file path
    Given the file path "<path>"
    When I open the connection
    Then opening <outcome>

    Examples:
      | path                            | outcome                                     |
      | example/TestProject.eapx        | succeeds                                     |
      | example/EmptyProject.eapx       | succeeds                                     |
      | example/does-not-exist.eapx     | fails with an error containing "file not found" |
      | /tmp/nonexistent-dir/x.eapx     | fails with an error containing "file not found" |
      | go.mod                          | fails with an error containing "Unable to locate database" |
      |                                 | fails with an error containing "empty file path" |
