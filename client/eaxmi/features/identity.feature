Feature: Regenerating the model identity
  As internal/service/sparx
  I want RemapIdentity to give every package, element, connector and diagram a
  fresh GUID while keeping all internal references consistent
  So two working copies of the same export import into one EA project side by side

  Scenario: Every GUID changes but the model stays internally consistent
    Given the model file "TestProject.xml"
    And I record every element GUID by path
    When I remap the model identity
    Then every recorded GUID has changed
    And there are 15 elements
    And there are 7 connectors
    And resolving the element "Model/Motivation_Package/Goal1" finds "Goal1"
    And the connector from "Model/Motivation_Package/Requirement1" to "Model/Motivation_Package/Goal1" still links those two elements

  Scenario: A remapped model still writes a file with no dangling references
    Given the model file "TestProject.xml"
    When I remap the model identity
    And I write and reopen as "remapped.xml"
    Then the reopened file has no dangling id references
    And there are 15 elements
