Feature: The ea_query MCP tool
  As an LLM agent talking to the MCP server
  I want an ea_query tool that runs SQL against a .eapx file
  So that I can read Sparx EA model data over MCP

  Background:
    Given a running MCP server

  Scenario: The server advertises exactly the ea_query tool
    When I list the MCP tools
    Then the tool "ea_query" is available
    And exactly 1 tool is advertised

  Scenario: ea_query returns query rows as JSON
    When I call "ea_query" with file "TestProject.eapx" and sql "select Object_ID, Name from t_object"
    Then the tool call is not an error
    And the tool JSON field "rowCount" equals "16"
    And the tool JSON contains "columns"
    And the tool JSON contains "Stakeholder1"

  Scenario: ea_query returns an empty row list, not null, when nothing matches
    When I call "ea_query" with file "TestProject.eapx" and sql "select Object_ID from t_object where Object_ID = 999999"
    Then the tool call is not an error
    And the tool JSON field "rowCount" equals "0"
    And the tool JSON field "rows" equals "[]"

  Scenario Outline: ea_query needs both file and sql
    When I call "ea_query" with only <arg> set
    Then the tool call is an error containing "<message>"

    Examples:
      | arg  | message                      |
      | file | missing required argument: sql  |
      | sql  | missing required argument: file |

  Scenario: ea_query surfaces an open failure as a tool error
    When I call "ea_query" with file "missing.eapx" and sql "select Object_ID from t_object"
    Then the tool call is an error containing "file not found"

  Scenario: ea_query rejects a write statement as a tool error
    When I call "ea_query" with file "TestProject.eapx" and sql "delete from t_object"
    Then the tool call is an error containing "only read-only SELECT statements are supported"
