Feature: The ArchiMate editing tools
  As an LLM agent talking to the MCP server
  I want tools that edit a copy of a model and save it for re-import into EA
  So that I can change a Sparx EA model over MCP

  Background:
    Given a running MCP server with a fake model

  Scenario: The editing tools are advertised
    When I list the MCP tools
    Then the tool "ea_create_element" is available
    And the tool "ea_update_element" is available
    And the tool "ea_delete_element" is available
    And the tool "ea_create_relationship" is available
    And the tool "ea_delete_relationship" is available
    And the tool "ea_create_package" is available
    And the tool "ea_update_package" is available
    And the tool "ea_copy_package" is available
    And the tool "ea_delete_package" is available
    And the tool "ea_place_on_diagram" is available
    And the tool "ea_move_on_diagram" is available
    And the tool "ea_remove_from_diagram" is available

  Scenario: ea_create_element maps its arguments and saves the copy
    When I call "ea_create_element" with:
      | file   | model.xml        |
      | output | edited.xml       |
      | parent | Model/Motivation |
      | type   | ArchiMate.Goal   |
      | name   | NewAim           |
      | note   | the objective    |
    Then the tool call is not an error
    And the model was opened with "model.xml"
    And the model method "CreateElement(Model/Motivation,ArchiMate.Goal,NewAim,the objective)" was called
    And the model was saved to "edited.xml"
    And the tool JSON contains "saved"

  Scenario: ea_update_element applies only the fields it is given
    When I call "ea_update_element" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | ref    | Model/Motivation/Goal1 |
      | name   | Goal1Renamed           |
    Then the tool call is not an error
    And the model method "RenameElement(Model/Motivation/Goal1,Goal1Renamed)" was called
    And the model method "SetElementDocumentation" was not called
    And the model method "MoveElement" was not called
    And the model was saved to "edited.xml"

  Scenario: ea_delete_element reports success and saves
    When I call "ea_delete_element" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | ref    | Model/Motivation/Goal1 |
    Then the tool call is not an error
    And the model method "DeleteElement(Model/Motivation/Goal1)" was called
    And the model was saved to "edited.xml"

  Scenario: ea_create_relationship maps its arguments
    When I call "ea_create_relationship" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | source | Model/Motivation/Req1  |
      | target | Model/Motivation/Goal1 |
      | type   | ArchiMate.Realization  |
    Then the tool call is not an error
    And the model method "CreateRelationship(Model/Motivation/Req1,Model/Motivation/Goal1,ArchiMate.Realization,,)" was called

  Scenario: ea_delete_relationship by two endpoints
    When I call "ea_delete_relationship" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | source | Model/Motivation/Req1  |
      | target | Model/Motivation/Goal1 |
    Then the tool call is not an error
    And the model method "DeleteRelationshipsBetween(Model/Motivation/Req1,Model/Motivation/Goal1)" was called

  Scenario: ea_delete_relationship needs an id or both endpoints
    When I call "ea_delete_relationship" with:
      | file   | model.xml             |
      | output | edited.xml            |
      | source | Model/Motivation/Req1 |
    Then the tool call is an error containing "pass either 'id' or both"

  Scenario: ea_copy_package maps ref and dest
    When I call "ea_copy_package" with:
      | file   | model.xml        |
      | output | edited.xml       |
      | ref    | Model/Motivation |
      | dest   | Model/Archive    |
    Then the model method "CopyPackage(Model/Motivation,Model/Archive)" was called

  Scenario: ea_delete_package passes the cascade flag
    When I call "ea_delete_package" with:
      | file    | model.xml        |
      | output  | edited.xml       |
      | ref     | Model/Motivation |
      | cascade | true             |
    Then the model method "DeletePackage(Model/Motivation,true)" was called

  Scenario: ea_place_on_diagram passes the rectangle
    When I call "ea_place_on_diagram" with:
      | file    | model.xml                 |
      | output  | edited.xml                |
      | diagram | Model/Motivation/Overview |
      | element | Model/Motivation/Goal1    |
      | left    | 10                        |
      | top     | 20                        |
      | right   | 110                       |
      | bottom  | 90                        |
    Then the tool call is not an error
    And the model method "AddToDiagram(Model/Motivation/Overview,Model/Motivation/Goal1,{10 20 110 90})" was called
    And the model was saved to "edited.xml"

  Scenario: an editing tool refuses to write over the source file
    When I call "ea_delete_element" with:
      | file   | model.xml              |
      | output | model.xml              |
      | ref    | Model/Motivation/Goal1 |
    Then the tool call is an error containing "output must be a different path"
    And the model was not saved

  Scenario: an editing tool needs the output argument
    When I call "ea_delete_element" with:
      | file | model.xml              |
      | ref  | Model/Motivation/Goal1 |
    Then the tool call is an error containing "missing required argument: output"

  Scenario: a service error stops the save
    Given the model returns the error "Realization not allowed between these types"
    When I call "ea_create_relationship" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | source | Model/Motivation/Goal1 |
      | target | Model/Motivation/Goal2 |
      | type   | ArchiMate.Realization  |
    Then the tool call is an error containing "not allowed"
    And the model was not saved

  Scenario: a save failure is reported as a tool error
    Given saving the model fails with "disk full"
    When I call "ea_delete_element" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | ref    | Model/Motivation/Goal1 |
    Then the tool call is an error containing "disk full"

  Scenario: ea_update_element with note and parent runs both mutations
    When I call "ea_update_element" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | ref    | Model/Motivation/Goal1 |
      | note   | revised                |
      | parent | Model/Archive          |
    Then the tool call is not an error
    And the model method "SetElementDocumentation(Model/Motivation/Goal1,revised)" was called
    And the model method "MoveElement(Model/Motivation/Goal1,Model/Archive)" was called
    And the model method "RenameElement" was not called

  Scenario: ea_update_element with no changes just reads the element back
    When I call "ea_update_element" with:
      | file   | model.xml              |
      | output | edited.xml             |
      | ref    | Model/Motivation/Goal1 |
    Then the tool call is not an error
    And the model method "Element(Model/Motivation/Goal1)" was called
    And the model was saved to "edited.xml"

  Scenario: ea_create_package maps its arguments
    When I call "ea_create_package" with:
      | file   | model.xml   |
      | output | edited.xml  |
      | parent | Model       |
      | name   | Realisation |
    Then the tool call is not an error
    And the model method "CreatePackage(Model,Realisation)" was called

  Scenario: ea_update_package renames then moves
    When I call "ea_update_package" with:
      | file   | model.xml        |
      | output | edited.xml       |
      | ref    | Model/Motivation |
      | name   | Motivation2      |
      | parent | Model/Archive    |
    Then the model method "RenamePackage(Model/Motivation,Motivation2)" was called
    And the model method "MovePackage(Model/Motivation,Model/Archive)" was called

  Scenario: ea_delete_relationship by id
    When I call "ea_delete_relationship" with:
      | file   | model.xml |
      | output | edited.xml |
      | id     | EAID_R    |
    Then the model method "DeleteRelationship(EAID_R)" was called

  Scenario: ea_move_on_diagram and ea_remove_from_diagram reach the model
    When I call "ea_move_on_diagram" with:
      | file    | model.xml                 |
      | output  | edited.xml                |
      | diagram | Model/Motivation/Overview |
      | element | Model/Motivation/Goal1    |
      | left    | 5                         |
      | top     | 5                         |
      | right   | 55                        |
      | bottom  | 45                        |
    Then the model method "MoveOnDiagram(Model/Motivation/Overview,Model/Motivation/Goal1,{5 5 55 45})" was called
    When I call "ea_remove_from_diagram" with:
      | file    | model.xml                 |
      | output  | edited.xml                |
      | diagram | Model/Motivation/Overview |
      | element | Model/Motivation/Goal1    |
    Then the model method "RemoveFromDiagram(Model/Motivation/Overview,Model/Motivation/Goal1)" was called
