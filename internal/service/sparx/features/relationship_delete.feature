Feature: Delete a relationship (method 5)

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_delete.xml |
      | root   | relationship delete     |

  Scenario: Delete a just-created relationship by id
    When I relate "relationship delete/Motivation_Package/Value1" to "relationship delete/Motivation_Package/Goal1" as "ArchiMate.Association"
    And I delete the last created relationship
    Then the delete succeeds
    And after reload the element "relationship delete/Motivation_Package/Value1" has no relation to "Goal1"
