Feature: Delete a relationship (method 5)

  Each scenario declares its fixture and its own tmp file.

  Scenario: Create then delete a relationship by id
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_delete.xml |
      | root   | relationship delete     |
    When I relate "relationship delete/Motivation_Package/Value1" to "relationship delete/Motivation_Package/Goal1" as "ArchiMate.Association"
    And I delete the last created relationship
    Then the delete succeeds
    And after reload the element "relationship delete/Motivation_Package/Value1" has no relation to "Goal1"

  Scenario: Deleting a relationship that does not exist is an error
    Given the working model:
      | source | TestProject.xml             |
      | output | relationship_delete_bad.xml |
      | root   | relationship delete bad     |
    When I delete the relationship "EAID_00000000_0000_0000_0000_000000000000"
    Then the delete fails with "no relationship"

  Scenario: Deleting an existing fixture relationship
    Given the working model:
      | source | TestProject.xml                 |
      | output | relationship_delete_fixture.xml |
      | root   | relationship delete fixture     |
    When I delete the relationship between "relationship delete fixture/Motivation_Package/Requirement1" and "relationship delete fixture/Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "relationship delete fixture/Motivation_Package/Requirement1" has no relation to "Goal1"
