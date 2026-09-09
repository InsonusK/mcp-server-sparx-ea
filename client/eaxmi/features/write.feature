Feature: Editing the model and writing it back
  As internal/service/sparx
  I want to add / change / remove packages, elements, connectors and diagram
  placements and write a file that parses back to the same model
  So the service can produce an importable working copy

  Scenario: A new model starts with one root package and nothing else
    Given a new model with root "Fresh"
    Then the root package names are "Fresh"
    And there are 0 elements
    And there are 0 connectors

  Scenario: A blank root name is refused
    Then creating a new model with root "   " fails with "root name is required"

  Scenario: Build a package tree and an element from scratch, then round-trip it
    Given a new model with root "Fresh"
    When I add a package "Domain" under "Fresh"
    And I add a package "Sub" under "Fresh/Domain"
    And I add an element "Aim" of type "uml:Class" stereotype "ArchiMate_Goal" under "Fresh/Domain/Sub"
    And I set the documentation of the element "Fresh/Domain/Sub/Aim" to "the objective"
    And I write and reopen as "fresh_tree.xml"
    Then the element "Fresh/Domain/Sub/Aim" is:
      | field         | value                |
      | umlType       | uml:Class            |
      | stereotype    | ArchiMate_Goal       |
      | documentation | the objective        |
      | package       | Fresh/Domain/Sub     |
    And the reopened file has no dangling id references

  Scenario: Renaming and moving an element survives a round-trip
    Given the model file "TestProject.xml"
    When I add a package "Archive" under "Model"
    And I set the name of the element "Model/Motivation_Package/Value1" to "CustomerValue"
    And I move the element "Model/Motivation_Package/CustomerValue" under "Model/Archive"
    And I write and reopen as "rename_move.xml"
    Then the element "Model/Archive/CustomerValue" is:
      | field   | value         |
      | name    | CustomerValue |
      | package | Model/Archive |
    And resolving the element "Model/Motivation_Package/CustomerValue" finds nothing

  Scenario: Removing an element also removes its connectors and diagram objects
    Given the model file "TestProject.xml"
    When I remove the element "Model/Motivation_Package/Goal1"
    And I write and reopen as "remove_element.xml"
    Then there are 14 elements
    And there are 2 connectors
    And the reopened file has no dangling id references

  Scenario: Adding a connector of each representation writes a consistent file
    Given the model file "TestProject.xml"
    When I add a connector from "Model/Motivation_Package/Value1" to "Model/Motivation_Package/Meaning1" ea-type "Association" repr "association" stereotype "ArchiMate_Association"
    And I add a connector from "Model/Motivation_Package/Capability1" to "Model/Motivation_Package/Resource1" ea-type "Dependency" repr "dependency" stereotype "ArchiMate_Realization"
    And I add a connector from "Model/Motivation_Package/Driver1" to "Model/Motivation_Package/Outcome1" ea-type "ControlFlow" repr "controlflow" stereotype "ArchiMate_Influence"
    And I write and reopen as "add_connectors.xml"
    Then there are 10 connectors
    And the connector from "Model/Motivation_Package/Value1" to "Model/Motivation_Package/Meaning1" is:
      | field      | value                |
      | eaType     | Association          |
      | stereotype | ArchiMate_Association |
    And the reopened file has no dangling id references

  Scenario: Placing and removing a diagram object survives a round-trip
    Given the model file "TestProject.xml"
    When I add an element "Extra" of type "uml:Class" stereotype "ArchiMate_Goal" under "Model/Motivation_Package"
    And I place the element "Model/Motivation_Package/Extra" on the diagram "Model/Motivation_Package/Motivation_Diagram" at 10,20,110,90
    And I write and reopen as "place_object.xml"
    Then the diagram "Model/Motivation_Package/Motivation_Diagram" has 16 objects and 7 links
    And the diagram "Model/Motivation_Package/Motivation_Diagram" places "Model/Motivation_Package/Extra" at 10,20,110,90

  Scenario: Removing a non-empty package removes its whole subtree
    Given the model file "TestProject.xml"
    When I remove the package "Model/Motivation_Package"
    And I write and reopen as "remove_package.xml"
    Then there are 0 elements
    And there are 0 connectors
    And the reopened file has no dangling id references
