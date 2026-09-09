Feature: Delete an element (method 4)

  Every scenario is an independent package inside tmp/element_delete.xml.

  Background:
    Given the working model:
      | source | TestProject.xml    |
      | output | element_delete.xml |

  Scenario: Deleting an element also removes every relationship attached to it
    When I delete the element "Motivation_Package/Goal1"
    Then the delete succeeds
    And after reload the element "Motivation_Package/Goal1" cannot be found
    And after reload the element "Motivation_Package/Assessment1" has no relation to "Goal1"
    And after reload the element "Motivation_Package/Requirement1" has no relation to "Goal1"

  Scenario: Deleting a missing element is an error
    When I delete the element "Motivation_Package/Ghost"
    Then the delete fails with "no element for"
