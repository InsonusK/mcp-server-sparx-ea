Feature: Create relationships, with ArchiMate validation (method 5)
  As the MCP server
  I want the service to refuse a relationship ArchiMate does not permit or that
  already exists, so an agent cannot put the model into an invalid state.

  Every scenario creates its own copy of the helper elements (GoalA, GoalB,
  ReqA, ProcA, ProcB, ObjA) and writes its own tmp file, so no scenario depends
  on another.

  Scenario: A worked example of every supported relationship
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_create.xml |
      | root   | relationship create     |
    And the relationship helper elements in "relationship create/Motivation_Package"
    When I relate "relationship create/Motivation_Package/ReqA" to "relationship create/Motivation_Package/GoalA" as "ArchiMate.Realization"
    And I relate "relationship create/Motivation_Package/GoalA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Specialization"
    And I relate "relationship create/Motivation_Package/GoalA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Composition"
    And I relate "relationship create/Motivation_Package/GoalA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Aggregation"
    And I relate "relationship create/Motivation_Package/GoalA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Association"
    And I relate "relationship create/Motivation_Package/ProcA" to "relationship create/Motivation_Package/GoalA" as "ArchiMate.Influence"
    And I relate "relationship create/Motivation_Package/ProcA" to "relationship create/Motivation_Package/ProcB" as "ArchiMate.Triggering"
    And I relate "relationship create/Motivation_Package/ProcA" to "relationship create/Motivation_Package/ProcB" as "ArchiMate.Flow"
    And I relate "relationship create/Motivation_Package/ProcA" to "relationship create/Motivation_Package/ObjA" as "ArchiMate.Access"
    And I relate "relationship create/Motivation_Package/ProcA" to "relationship create/Motivation_Package/GoalB" as "ArchiMate.Serving"
    Then the relate succeeds

  Scenario Outline: Relationships ArchiMate does not permit are refused
    Given the working model:
      | source | TestProject.xml                 |
      | output | relationship_create_refused.xml |
      | root   | relationship create refused     |
    And the relationship helper elements in "relationship create refused/Motivation_Package"
    When I relate "relationship create refused/Motivation_Package/<source>" to "relationship create refused/Motivation_Package/<target>" as "<relation>"
    Then the relate fails with "<message>"

    Examples:
      | source | target | relation                 | message                                 |
      | GoalA  | ProcA  | ArchiMate.Specialization | same type                               |
      | GoalA  | ReqA   | ArchiMate.Composition    | different ArchiMate types               |
      | GoalA  | GoalB  | ArchiMate.Triggering     | behaviour elements                      |
      | ProcA  | ObjA   | ArchiMate.Influence      | must target a motivation element        |
      | ReqA   | GoalA  | ArchiMate.Frobnicate     | not a known ArchiMate relationship type |

  Scenario: The same (type, source, target) relationship cannot be created twice
    Given the working model:
      | source | TestProject.xml             |
      | output | relationship_create_dup.xml |
      | root   | relationship create dup     |
    And the relationship helper elements in "relationship create dup/Motivation_Package"
    When I relate "relationship create dup/Motivation_Package/ReqA" to "relationship create dup/Motivation_Package/GoalB" as "ArchiMate.Realization"
    And I relate "relationship create dup/Motivation_Package/ReqA" to "relationship create dup/Motivation_Package/GoalB" as "ArchiMate.Realization"
    Then the relate fails with "already exists"

  Scenario: A named relationship round-trips and is visible from both ends
    Given the working model:
      | source | TestProject.xml               |
      | output | relationship_create_named.xml |
      | root   | relationship create named     |
    And the relationship helper elements in "relationship create named/Motivation_Package"
    When I relate "relationship create named/Motivation_Package/Stakeholder1" to "relationship create named/Motivation_Package/GoalA" as "ArchiMate.Association" named "cares about"
    Then the relate succeeds
    And after reload the element "relationship create named/Motivation_Package/Stakeholder1" relations include:
      | type                 | name        | direction | otherName |
      | ArchiMate.Association | cares about | outgoing  | GoalA     |
    And after reload the element "relationship create named/Motivation_Package/GoalA" relations include:
      | type                 | direction | otherName    |
      | ArchiMate.Association | incoming  | Stakeholder1 |
