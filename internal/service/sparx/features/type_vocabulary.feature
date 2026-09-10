Feature: The ArchiMate 3.2 type vocabulary
  Architectural concern: the service exposes ArchiMate types and refuses to
  create anything outside the hard-coded vocabulary.

  Scenario: Whole ArchiMate 3.2 element types are in the vocabulary
    Then "ArchiMate.Goal" is a known element type
    And "ArchiMate.BusinessProcess" is a known element type
    And "ArchiMate.ApplicationComponent" is a known element type
    And "ArchiMate.Node" is a known element type
    And "ArchiMate.Plateau" is a known element type

  Scenario: Non-ArchiMate names are not element types
    Then "ArchiMate.Frobnicator" is not a known element type
    And "uml:Class" is not a known element type
    And "ArchiMate.Junction" is not a known element type

  Scenario: The ArchiMate relationship vocabulary
    Then "ArchiMate.Realization" is a known relationship type
    And "ArchiMate.Serving" is a known relationship type
    And "ArchiMate.Influence" is a known relationship type

  Scenario: The relationship rules are queryable without touching a model
    Then "ArchiMate.Assignment" from "ArchiMate.ApplicationComponent" to "ArchiMate.ApplicationFunction" is "allow"
    And "ArchiMate.Realization" from "ArchiMate.ApplicationComponent" to "ArchiMate.ApplicationService" is "allow"
    And "ArchiMate.Triggering" from "ArchiMate.Plateau" to "ArchiMate.Plateau" is "warn"
    And "ArchiMate.Assignment" from "ArchiMate.Goal" to "ArchiMate.Node" is "deny"
    And "ArchiMate.Triggering" from "ArchiMate.BusinessObject" to "ArchiMate.BusinessProcess" is "deny"
    And "ArchiMate.Nonsense" from "ArchiMate.Goal" to "ArchiMate.Goal" is "deny"
