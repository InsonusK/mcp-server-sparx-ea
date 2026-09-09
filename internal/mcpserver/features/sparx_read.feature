Feature: The ArchiMate read tools
  As an LLM agent talking to the MCP server
  I want tools that read an exported .xml model
  So that I can inspect a Sparx EA model without SQL

  Background:
    Given a running MCP server with a fake model

  Scenario: The read tools are advertised
    When I list the MCP tools
    Then exactly 17 tools are advertised
    And the tool "ea_model_tree" is available
    And the tool "ea_element" is available
    And the tool "ea_package" is available
    And the tool "ea_diagram" is available
    And the tool "ea_archimate_types" is available

  Scenario: ea_model_tree opens the file and returns the tree
    When I call "ea_model_tree" with:
      | file | model.xml |
    Then the tool call is not an error
    And the model was opened with "model.xml"
    And the model method "Tree()" was called
    And the tool JSON contains "Motivation"

  Scenario: ea_element passes the ref through and returns the element
    When I call "ea_element" with:
      | file | model.xml              |
      | ref  | Model/Motivation/Goal1 |
    Then the tool call is not an error
    And the model method "Element(Model/Motivation/Goal1)" was called
    And the tool JSON field "type" equals "ArchiMate.Goal"

  Scenario: ea_package and ea_diagram pass the ref through
    When I call "ea_package" with:
      | file | model.xml        |
      | ref  | Model/Motivation |
    Then the model method "Package(Model/Motivation)" was called
    When I call "ea_diagram" with:
      | file | model.xml                 |
      | ref  | Model/Motivation/Overview |
    Then the model method "Diagram(Model/Motivation/Overview)" was called

  Scenario: ea_archimate_types lists the vocabulary without opening a file
    When I call "ea_archimate_types" with:
      | dummy | x |
    Then the tool call is not an error
    And no model method was called
    And the tool JSON contains "ArchiMate.Goal"
    And the tool JSON contains "ArchiMate.Realization"

  Scenario: a service error becomes a tool error, not a transport error
    Given the model returns the error "no element for that ref"
    When I call "ea_element" with:
      | file | model.xml |
      | ref  | Ghost     |
    Then the tool call is an error containing "no element for that ref"

  Scenario: a read tool needs the file argument
    When I call "ea_element" with:
      | ref | Model/Motivation/Goal1 |
    Then the tool call is an error containing "missing required argument: file"
