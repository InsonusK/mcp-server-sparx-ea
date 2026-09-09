Feature: Delete a relationship (method 5)

  Every scenario is an independent package inside tmp/relationship_delete.xml.

  Background:
    Given the working model:
      | source | TestProject.xml         |
      | output | relationship_delete.xml |

  Scenario: Create then delete a relationship by id
    When I relate "Motivation_Package/Value1" to "Motivation_Package/Goal1" as "ArchiMate.Association"
    And I delete the last created relationship
    Then the delete succeeds
    And after reload the element "Motivation_Package/Value1" has no relation to "Goal1"

  Scenario: Deleting a relationship that does not exist is an error
    When I delete the relationship "EAID_00000000_0000_0000_0000_000000000000"
    Then the delete fails with "no relationship"

  Scenario: Deleting an existing fixture relationship
    When I delete the relationship between "Motivation_Package/Requirement1" and "Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "Motivation_Package/Requirement1" has no relation to "Goal1"
