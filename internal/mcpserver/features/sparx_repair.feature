Feature: The EA-representation repair tools
  As an LLM agent talking to the MCP server
  I want to find and fix elements whose EA base type doesn't match their
  ArchiMate type — the kind of mismatch that makes Sparx EA render an element
  as an anonymous class instead of its real ArchiMate shape
  So that I can repair a model exported before that mapping was corrected

  Background:
    Given a running MCP server with a fake model

  Scenario: The repair tools are advertised
    When I list the MCP tools
    Then the tool "ea_validate_model" is available
    And the tool "ea_fix_elements" is available
    And the tool "ea_validate_and_fix_model" is available

  Scenario: ea_validate_model is read-only
    When I call "ea_validate_model" with:
      | file | model.xml |
    Then the tool call is not an error
    And the model was opened with "model.xml"
    And the model method "ValidateModel()" was called
    And the model was not saved

  Scenario: ea_fix_elements passes the refs through and saves the copy
    When I call "ea_fix_elements" with:
      | file   | model.xml                 |
      | output | edited.xml                |
      | refs   | ["EAID_1","EAID_2"]       |
    Then the tool call is not an error
    And the model method "FixElements([EAID_1 EAID_2])" was called
    And the model was saved to "edited.xml"

  Scenario: ea_validate_and_fix_model validates, fixes and saves in one step
    When I call "ea_validate_and_fix_model" with:
      | file   | model.xml  |
      | output | edited.xml |
    Then the tool call is not an error
    And the model method "ValidateAndFixModel()" was called
    And the model was saved to "edited.xml"

  Scenario: ea_fix_elements needs the output argument
    When I call "ea_fix_elements" with:
      | file | model.xml           |
      | refs | ["EAID_1"]          |
    Then the tool call is an error containing "missing required argument: output"
    And the model was not saved
