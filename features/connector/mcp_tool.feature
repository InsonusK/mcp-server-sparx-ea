Feature: The ea_query MCP tool
  As an LLM agent talking to the MCP server
  I want an ea_query tool that runs SQL against a .eapx file
  So that I can read Sparx EA model data over MCP

  Background:
    Given a running MCP server

  Scenario: The server advertises the ea_query tool
    When I list the MCP tools
    Then the tool "ea_query" is available

  Scenario: ea_query returns query rows as JSON
    When I call "ea_query" with file "example/TestProject.eapx" and sql "select Object_ID, Name from t_object"
    Then the tool call is not an error
    And the tool JSON field "rowCount" equals "16"
    And the tool JSON contains "columns"
    And the tool JSON contains "Stakeholder1"

  Scenario: ea_query surfaces an open failure as a tool error
    When I call "ea_query" with file "example/missing.eapx" and sql "select Object_ID from t_object"
    Then the tool call is an error containing "file not found"

  Scenario: ea_query rejects a write statement as a tool error
    When I call "ea_query" with file "example/TestProject.eapx" and sql "delete from t_object"
    Then the tool call is an error containing "only read-only SELECT statements are supported"
