Feature: The model-lifecycle tools
  As an LLM agent talking to the MCP server
  I want tools that build a model from scratch and manage its root package
  So that I can start a new Sparx EA model over MCP, not only edit an export

  Background:
    Given a running MCP server with a fake model

  Scenario: The lifecycle tools are advertised
    When I list the MCP tools
    Then the tool "ea_new_model" is available
    And the tool "ea_create_root_package" is available
    And the tool "ea_set_root_name" is available

  Scenario: ea_new_model builds a model from scratch and saves it
    When I call "ea_new_model" with:
      | root   | Sparx EA MCP Server |
      | output | fresh.xml           |
    Then the tool call is not an error
    And the model method "NewModel(Sparx EA MCP Server)" was called
    And the model method "Tree()" was called
    And the model was saved to "fresh.xml"
    And the tool JSON contains "saved"

  Scenario: ea_new_model needs the output argument
    When I call "ea_new_model" with:
      | root | Sparx EA MCP Server |
    Then the tool call is an error containing "missing required argument: output"
    And the model was not saved

  Scenario: ea_create_root_package maps its name and saves the copy
    When I call "ea_create_root_package" with:
      | file   | model.xml   |
      | output | edited.xml  |
      | name   | Motivation2 |
    Then the tool call is not an error
    And the model was opened with "model.xml"
    And the model method "CreateRootPackage(Motivation2)" was called
    And the model was saved to "edited.xml"

  Scenario: ea_set_root_name forwards the fresh_identity flag
    When I call "ea_set_root_name" with:
      | file           | model.xml |
      | output         | edited.xml |
      | name           | Renamed   |
      | fresh_identity | true      |
    Then the tool call is not an error
    And the model method "SetRootName(Renamed,true)" was called
    And the model was saved to "edited.xml"
    And the tool JSON contains "rootName"

  Scenario: ea_set_root_name defaults fresh_identity to false
    When I call "ea_set_root_name" with:
      | file   | model.xml  |
      | output | edited.xml |
      | name   | Renamed    |
    Then the model method "SetRootName(Renamed,false)" was called
