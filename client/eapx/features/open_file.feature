Feature: Opening a connection to a Sparx EA file
  As the MCP server
  I want Open(path) to connect to a readable JET database or fail with a clear reason
  So that a caller never queries against a broken handle

  Scenario: Opening the populated sample project
    When I open the Sparx EA file "TestProject.eapx"
    Then the connection opens successfully

  Scenario: Opening the empty sample project
    When I open the Sparx EA file "EmptyProject.eapx"
    Then the connection opens successfully

  Scenario Outline: Opening an unusable path fails with a specific message
    When I open the Sparx EA file "<file>"
    Then opening fails with "<message>"

    Examples:
      | file                | message                   |
      | missing-project.eapx | file not found            |
      | not-a-database.txt  | Unable to locate database |
      |                     | empty file path           |
