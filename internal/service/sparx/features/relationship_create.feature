Feature: Create relationships, with ArchiMate validation (method 5)
  As the MCP server
  I want the service to refuse a relationship ArchiMate does not permit
  So that an agent cannot put the model into an invalid state

  Scenario Outline: Relationship validity between two elements
    Given the working model:
      | source | TestProject.xml           |
      | output | relationship_create_matrix.xml |
      | root   | relationship create matrix |
    When I relate "relationship create matrix/Motivation_Package/<source>" to "relationship create matrix/Motivation_Package/<target>" as "<relation>"
    Then the relate <outcome>

    Examples:
      | source       | target  | relation                 | outcome                                              |
      | Requirement1 | Goal1   | ArchiMate.Realization    | succeeds                                             |
      | Stakeholder1 | Goal1   | ArchiMate.Association    | succeeds                                             |
      | Driver1      | Goal1   | ArchiMate.Influence      | succeeds                                             |
      | Goal1        | Goal1   | ArchiMate.Specialization | succeeds                                             |
      | Goal1        | Value1  | ArchiMate.Specialization | fails with "same type"                               |
      | Goal1        | Driver1 | ArchiMate.Composition    | fails with "different ArchiMate types"               |
      | Goal1        | Driver1 | ArchiMate.Triggering     | fails with "behaviour elements"                      |
      | Requirement1 | Goal1   | ArchiMate.Frobnicate     | fails with "not a known ArchiMate relationship type" |

  Scenario: Influence must target a motivation element
    Given the working model:
      | source | TestProject.xml                   |
      | output | relationship_create_influence.xml |
      | root   | relationship create influence     |
    When I create a "ArchiMate.BusinessProcess" named "Proc" in "relationship create influence/Motivation_Package" with note ""
    And I relate "relationship create influence/Motivation_Package/Driver1" to "relationship create influence/Motivation_Package/Proc" as "ArchiMate.Influence"
    Then the relate fails with "must target a motivation element"

  Scenario: A created relationship round-trips and is visible from both ends
    Given the working model:
      | source | TestProject.xml                |
      | output | relationship_create.xml        |
      | root   | relationship create            |
    When I relate "relationship create/Motivation_Package/Stakeholder1" to "relationship create/Motivation_Package/Goal1" as "ArchiMate.Association" named "cares about"
    Then the relate succeeds
    And after reload the element "relationship create/Motivation_Package/Stakeholder1" relations include:
      | type                 | name        | direction | otherName |
      | ArchiMate.Association | cares about | outgoing  | Goal1     |
    And after reload the element "relationship create/Motivation_Package/Goal1" relations include:
      | type                 | direction | otherName    |
      | ArchiMate.Association | incoming  | Stakeholder1 |
