Feature: Create and delete relationships, with ArchiMate validation (method 5)
  As the MCP server
  I want the service to refuse a relationship that ArchiMate does not permit
  So that an agent cannot put the model into an invalid state

  Background:
    Given a working copy of "TestProject.xml"

  Scenario Outline: Relationship validity between two elements
    When I relate "Model/Motivation_Package/<source>" to "Model/Motivation_Package/<target>" as "<relation>"
    Then the relate <outcome>

    Examples:
      | source       | target | relation               | outcome                                           |
      | Requirement1 | Goal1  | ArchiMate.Realization  | succeeds                                          |
      | Stakeholder1 | Goal1  | ArchiMate.Association  | succeeds                                          |
      | Driver1      | Goal1  | ArchiMate.Influence    | succeeds                                          |
      | Goal1        | Goal1  | ArchiMate.Specialization | succeeds                                        |
      | Goal1        | Value1 | ArchiMate.Specialization | fails with "same type"                          |
      | Goal1        | Driver1 | ArchiMate.Composition  | fails with "different ArchiMate types"           |
      | Goal1        | Driver1 | ArchiMate.Triggering   | fails with "behaviour elements"                  |
      | Requirement1 | Goal1  | ArchiMate.Frobnicate   | fails with "not a known ArchiMate relationship type" |

  Scenario: Influence must target a motivation element
    When I create a "ArchiMate.BusinessProcess" named "Proc" in "Model/Motivation_Package" with note ""
    And I relate "Model/Motivation_Package/Driver1" to "Model/Motivation_Package/Proc" as "ArchiMate.Influence"
    Then the relate fails with "must target a motivation element"

  Scenario: A created relationship round-trips and is visible from both ends
    When I relate "Model/Motivation_Package/Stakeholder1" to "Model/Motivation_Package/Goal1" as "ArchiMate.Association" named "cares about"
    Then the relate succeeds
    And after reload the element "Model/Motivation_Package/Stakeholder1" relations include:
      | type                 | name        | direction | otherName |
      | ArchiMate.Association | cares about | outgoing  | Goal1     |
    And after reload the element "Model/Motivation_Package/Goal1" relations include:
      | type                 | direction | otherName    |
      | ArchiMate.Association | incoming  | Stakeholder1 |

  Scenario: Delete a relationship by id
    When I relate "Model/Motivation_Package/Value1" to "Model/Motivation_Package/Goal1" as "ArchiMate.Association"
    And I delete the last created relationship
    Then the delete succeeds
    And after reload the element "Model/Motivation_Package/Value1" has no relation to "Goal1"
