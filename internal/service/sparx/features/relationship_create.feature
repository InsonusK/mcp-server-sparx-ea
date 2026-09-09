Feature: Create relationships, with ArchiMate validation (method 5)
  As the MCP server
  I want the service to refuse a relationship ArchiMate does not permit or that
  already exists, so an agent cannot put the model into an invalid state.

  Every scenario (and every Examples row) is an independent package inside
  tmp/relationship_create.xml, each carrying its own copy of the helper elements
  created in the Background — so no scenario depends on another's success.

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_create.xml |
    And I create a "ArchiMate.Goal" named "GoalA" in "Motivation_Package" with note ""
    And I create a "ArchiMate.Goal" named "GoalB" in "Motivation_Package" with note ""
    And I create a "ArchiMate.Requirement" named "ReqA" in "Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessProcess" named "ProcA" in "Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessProcess" named "ProcB" in "Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessObject" named "ObjA" in "Motivation_Package" with note ""

  Scenario: A worked example of every supported relationship
    When I relate "Motivation_Package/ReqA" to "Motivation_Package/GoalA" as "ArchiMate.Realization"
    And I relate "Motivation_Package/GoalA" to "Motivation_Package/GoalB" as "ArchiMate.Specialization"
    And I relate "Motivation_Package/GoalA" to "Motivation_Package/GoalB" as "ArchiMate.Composition"
    And I relate "Motivation_Package/GoalA" to "Motivation_Package/GoalB" as "ArchiMate.Aggregation"
    And I relate "Motivation_Package/GoalA" to "Motivation_Package/GoalB" as "ArchiMate.Association"
    And I relate "Motivation_Package/ProcA" to "Motivation_Package/GoalA" as "ArchiMate.Influence"
    And I relate "Motivation_Package/ProcA" to "Motivation_Package/ProcB" as "ArchiMate.Triggering"
    And I relate "Motivation_Package/ProcA" to "Motivation_Package/ProcB" as "ArchiMate.Flow"
    And I relate "Motivation_Package/ProcA" to "Motivation_Package/ObjA" as "ArchiMate.Access"
    And I relate "Motivation_Package/ProcA" to "Motivation_Package/GoalB" as "ArchiMate.Serving"
    Then the relate succeeds

  Scenario Outline: Relationships ArchiMate does not permit are refused
    When I relate "Motivation_Package/<source>" to "Motivation_Package/<target>" as "<relation>"
    Then the relate fails with "<message>"

    Examples:
      | source | target | relation                 | message                                |
      | GoalA  | ProcA  | ArchiMate.Specialization | same type                              |
      | GoalA  | ReqA   | ArchiMate.Composition    | different ArchiMate types              |
      | GoalA  | GoalB  | ArchiMate.Triggering     | behaviour elements                     |
      | ProcA  | ObjA   | ArchiMate.Influence      | must target a motivation element       |
      | ReqA   | GoalA  | ArchiMate.Frobnicate     | not a known ArchiMate relationship type |

  Scenario: The same (type, source, target) relationship cannot be created twice
    When I relate "Motivation_Package/ReqA" to "Motivation_Package/GoalB" as "ArchiMate.Realization"
    And I relate "Motivation_Package/ReqA" to "Motivation_Package/GoalB" as "ArchiMate.Realization"
    Then the relate fails with "already exists"

  Scenario: A named relationship round-trips and is visible from both ends
    When I relate "Motivation_Package/Stakeholder1" to "Motivation_Package/GoalA" as "ArchiMate.Association" named "cares about"
    Then the relate succeeds
    And after reload the element "Motivation_Package/Stakeholder1" relations include:
      | type                 | name        | direction | otherName |
      | ArchiMate.Association | cares about | outgoing  | GoalA     |
    And after reload the element "Motivation_Package/GoalA" relations include:
      | type                 | direction | otherName    |
      | ArchiMate.Association | incoming  | Stakeholder1 |
