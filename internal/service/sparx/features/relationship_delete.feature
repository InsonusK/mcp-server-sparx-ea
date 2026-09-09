Feature: Delete a relationship (method 5)

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_delete.xml |
      | root   | relationship delete     |

  Scenario: Create then delete a relationship by id
    When I relate "relationship delete/Motivation_Package/Value1" to "relationship delete/Motivation_Package/Goal1" as "ArchiMate.Association"
    And I delete the last created relationship
    Then the delete succeeds
    And after reload the element "relationship delete/Motivation_Package/Value1" has no relation to "Goal1"

  Scenario: Deleting a relationship that does not exist is an error
    When I delete the relationship "EAID_00000000_0000_0000_0000_000000000000"
    Then the delete fails with "no relationship"

  Scenario: Deleting an existing fixture relationship
    When I delete the relationship between "relationship delete/Motivation_Package/Requirement1" and "relationship delete/Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "relationship delete/Motivation_Package/Requirement1" has no relation to "Goal1"
