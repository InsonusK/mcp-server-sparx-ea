Feature: Create relationships, with ArchiMate validation (method 5)
  As the MCP server
  I want the service to refuse a relationship ArchiMate does not permit or that
  already exists, so an agent cannot put the model into an invalid state.

  All scenarios accumulate into tmp/relationship_create.xml — one file with a
  worked example of every relationship the service currently supports, plus the
  helper elements they connect.

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_create.xml |
      | root   | relationship create     |

  Scenario: Helper elements the relationship examples connect
    When I create a "ArchiMate.Goal" named "GoalA" in "relationship create/Motivation_Package" with note ""
    And I create a "ArchiMate.Goal" named "GoalB" in "relationship create/Motivation_Package" with note ""
    And I create a "ArchiMate.Requirement" named "ReqA" in "relationship create/Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessProcess" named "ProcA" in "relationship create/Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessProcess" named "ProcB" in "relationship create/Motivation_Package" with note ""
    And I create a "ArchiMate.BusinessObject" named "ObjA" in "relationship create/Motivation_Package" with note ""
    Then the create succeeds

  Scenario Outline: A worked example of every supported relationship
    When I relate "relationship create/Motivation_Package/<source>" to "relationship create/Motivation_Package/<target>" as "<relation>"
    Then the relate succeeds

    Examples:
      | source | target | relation                 |
      | ReqA   | GoalA  | ArchiMate.Realization    |
      | GoalA  | GoalB  | ArchiMate.Specialization |
      | GoalA  | GoalB  | ArchiMate.Composition    |
      | GoalA  | GoalB  | ArchiMate.Aggregation    |
      | GoalA  | GoalB  | ArchiMate.Association    |
      | ProcA  | GoalA  | ArchiMate.Influence      |
      | ProcA  | ProcB  | ArchiMate.Triggering     |
      | ProcA  | ProcB  | ArchiMate.Flow           |
      | ProcA  | ObjA   | ArchiMate.Access         |
      | ProcA  | GoalB  | ArchiMate.Serving        |

  Scenario Outline: Relationships ArchiMate does not permit are refused
    When I relate "relationship create/Motivation_Package/<source>" to "relationship create/Motivation_Package/<target>" as "<relation>"
    Then the relate fails with "<message>"

    Examples:
      | source | target | relation                 | message                                             |
      | GoalA  | ProcA  | ArchiMate.Specialization | same type                                            |
      | GoalA  | ReqA   | ArchiMate.Composition    | different ArchiMate types                            |
      | GoalA  | GoalB  | ArchiMate.Triggering     | behaviour elements                                   |
      | ProcA  | ObjA   | ArchiMate.Influence      | must target a motivation element                     |
      | ReqA   | GoalA  | ArchiMate.Frobnicate     | not a known ArchiMate relationship type              |

  Scenario: The same (type, source, target) relationship cannot be created twice
    When I relate "relationship create/Motivation_Package/ReqA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Realization"
    And I relate "relationship create/Motivation_Package/ReqA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Realization"
    Then the relate fails with "already exists"

  Scenario: A named relationship round-trips and is visible from both ends
    When I relate "relationship create/Motivation_Package/Stakeholder1" to "relationship create/Motivation_Package/GoalA" as "ArchiMate.Association" named "cares about"
    Then the relate succeeds
    And after reload the element "relationship create/Motivation_Package/Stakeholder1" relations include:
      | type                 | name        | direction | otherName |
      | ArchiMate.Association | cares about | outgoing  | GoalA     |
    And after reload the element "relationship create/Motivation_Package/GoalA" relations include:
      | type                 | direction | otherName    |
      | ArchiMate.Association | incoming  | Stakeholder1 |
